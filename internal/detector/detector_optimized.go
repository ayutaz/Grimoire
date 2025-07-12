package detector

import (
	"image"
	"math"
	"runtime"
	"sync"
)

// ParallelDetector is an optimized detector using parallel processing
type ParallelDetector struct {
	*Detector
	numWorkers int
	cache      *DetectorCache
}

// DetectorCache stores intermediate results for reuse
type DetectorCache struct {
	mu           sync.RWMutex
	preprocessed map[string]*image.Gray
	contours     map[string][]Contour
	symbols      map[string][]*Symbol
	maxCacheSize int
	accessOrder  []string
}

// NewParallelDetector creates a new parallel detector
func NewParallelDetector(cfg Config) *ParallelDetector {
	return &ParallelDetector{
		Detector:   NewDetector(cfg),
		numWorkers: runtime.NumCPU(),
		cache:      NewDetectorCache(100), // Cache up to 100 entries
	}
}

// NewDetectorCache creates a new cache
func NewDetectorCache(maxSize int) *DetectorCache {
	return &DetectorCache{
		preprocessed: make(map[string]*image.Gray),
		contours:     make(map[string][]Contour),
		symbols:      make(map[string][]*Symbol),
		maxCacheSize: maxSize,
		accessOrder:  make([]string, 0, maxSize),
	}
}

// detectSymbolsFromContoursParallel processes contours in parallel
func (d *ParallelDetector) detectSymbolsFromContoursParallel(contours []Contour, binary *image.Gray) []*Symbol {
	if len(contours) == 0 {
		return []*Symbol{}
	}

	// First, find outer circle (sequential)
	var outerCircle *Symbol
	for _, contour := range contours {
		if contour.Area < float64(d.minContourArea) {
			continue
		}
		symbolType := d.classifyContour(contour)
		if symbolType == OuterCircle {
			outerCircle = &Symbol{
				Type:       symbolType,
				Position:   Position{X: float64(contour.Center.X), Y: float64(contour.Center.Y)},
				Size:       math.Sqrt(contour.Area),
				Confidence: contour.Circularity,
				Pattern:    "empty",
				Properties: make(map[string]interface{}),
			}
			break
		}
	}

	// Process remaining contours in parallel
	type symbolResult struct {
		symbol *Symbol
		index  int
	}

	resultChan := make(chan symbolResult, len(contours))
	var wg sync.WaitGroup

	// Create worker pool
	workChan := make(chan int, len(contours))

	// Start workers
	for w := 0; w < d.numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range workChan {
				contour := contours[idx]
				if contour.Area < float64(d.minContourArea) {
					continue
				}

				symbolType := d.classifyContour(contour)
				if symbolType == OuterCircle || symbolType == Unknown {
					continue
				}

				// Detect internal pattern
				pattern := PatternEmpty
				if symbolType == Square || symbolType == Circle || symbolType == Pentagon ||
					symbolType == Hexagon || symbolType == Star {
					pattern = d.detectInternalPattern(contour, binary)
				}

				symbol := &Symbol{
					Type:       symbolType,
					Position:   Position{X: float64(contour.Center.X), Y: float64(contour.Center.Y)},
					Size:       math.Sqrt(contour.Area),
					Confidence: 0.7,
					Pattern:    pattern,
					Properties: make(map[string]interface{}),
				}

				// Check if within outer circle
				if outerCircle != nil {
					centerDist := math.Sqrt(math.Pow(symbol.Position.X-outerCircle.Position.X, 2) +
						math.Pow(symbol.Position.Y-outerCircle.Position.Y, 2))
					if centerDist < outerCircle.Size*0.9 {
						if symbolType == Star {
							if centerDist < outerCircle.Size*0.3 {
								resultChan <- symbolResult{symbol: symbol, index: idx}
							}
						} else {
							resultChan <- symbolResult{symbol: symbol, index: idx}
						}
					}
				} else {
					resultChan <- symbolResult{symbol: symbol, index: idx}
				}
			}
		}()
	}

	// Send work to workers
	for i := range contours {
		workChan <- i
	}
	close(workChan)

	// Wait for workers and close result channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	symbols := make([]*Symbol, 0, len(contours))
	if outerCircle != nil {
		symbols = append(symbols, outerCircle)
	}

	for result := range resultChan {
		symbols = append(symbols, result.symbol)
	}

	return symbols
}

// findContoursParallel finds contours using parallel processing
func (d *ParallelDetector) findContoursParallel(binary *image.Gray) []Contour {
	bounds := binary.Bounds()
	height := bounds.Dy()

	// Split image into horizontal strips for parallel processing
	stripsPerWorker := height / d.numWorkers
	if stripsPerWorker < 50 { // Minimum strip height
		stripsPerWorker = 50
	}

	numStrips := height / stripsPerWorker
	if numStrips < 1 {
		numStrips = 1
	}

	// Process strips in parallel
	var wg sync.WaitGroup
	contoursChan := make(chan []Contour, numStrips)

	for i := 0; i < numStrips; i++ {
		wg.Add(1)
		startY := i * stripsPerWorker
		endY := startY + stripsPerWorker
		if i == numStrips-1 {
			endY = height
		}

		go func(sy, ey int) {
			defer wg.Done()
			// Extract sub-image
			subBounds := image.Rect(0, sy, bounds.Dx(), ey)
			subImage := image.NewGray(subBounds)

			for y := sy; y < ey; y++ {
				for x := 0; x < bounds.Dx(); x++ {
					subImage.Set(x, y, binary.At(x, y))
				}
			}

			// Find contours in strip
			stripContours := d.findContours(subImage)

			// Adjust Y coordinates
			for i := range stripContours {
				for j := range stripContours[i].Points {
					stripContours[i].Points[j].Y += sy
				}
				stripContours[i].calculateProperties()
			}

			contoursChan <- stripContours
		}(startY, endY)
	}

	go func() {
		wg.Wait()
		close(contoursChan)
	}()

	// Collect all contours
	var allContours []Contour
	for contours := range contoursChan {
		allContours = append(allContours, contours...)
	}

	return allContours
}

// Detect performs optimized symbol detection
func (d *ParallelDetector) Detect(imagePath string) ([]*Symbol, []Connection, error) {
	// Check cache first
	if cached := d.cache.getSymbols(imagePath); cached != nil {
		connections := d.improvedDetectConnections(nil, cached)
		return cached, connections, nil
	}

	// Load and validate image (same as original)
	img, err := d.loadAndValidateImage(imagePath)
	if err != nil {
		return nil, nil, err
	}

	// Convert to grayscale
	gray := d.toGrayscale(img)

	// Check cache for preprocessed image
	binary := d.cache.getPreprocessed(imagePath)
	if binary == nil {
		binary = d.preprocessImage(gray)
		d.cache.setPreprocessed(imagePath, binary)
	}

	// Find outer circle
	outerCircle := d.findOuterCircleFromGrayscale(gray)

	// Find contours in parallel
	contours := d.findContoursParallel(binary)

	// Add outer circle if found
	if outerCircle != nil {
		contours = append([]Contour{*outerCircle}, contours...)
	}

	// Detect symbols in parallel
	symbols := d.detectSymbolsFromContoursParallel(contours, binary)

	// Deduplicate nearby stars
	symbols = d.deduplicateNearbyStars(symbols)

	// Cache symbols
	d.cache.setSymbols(imagePath, symbols)

	// Detect connections
	connections := d.improvedDetectConnections(binary, symbols)

	// Validate results
	if err := d.validateResults(symbols, imagePath); err != nil {
		return nil, nil, err
	}

	return symbols, connections, nil
}

// Cache methods
func (c *DetectorCache) getPreprocessed(key string) *image.Gray {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.updateAccessOrder(key)
	return c.preprocessed[key]
}

func (c *DetectorCache) setPreprocessed(key string, img *image.Gray) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest if needed
	if len(c.preprocessed) >= c.maxCacheSize {
		c.evictOldest()
	}

	c.preprocessed[key] = img
	c.updateAccessOrder(key)
}

func (c *DetectorCache) getSymbols(key string) []*Symbol {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.updateAccessOrder(key)
	return c.symbols[key]
}

func (c *DetectorCache) setSymbols(key string, symbols []*Symbol) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest if needed
	if len(c.symbols) >= c.maxCacheSize {
		c.evictOldest()
	}

	c.symbols[key] = symbols
	c.updateAccessOrder(key)
}

func (c *DetectorCache) updateAccessOrder(key string) {
	// Remove if exists
	for i, k := range c.accessOrder {
		if k == key {
			c.accessOrder = append(c.accessOrder[:i], c.accessOrder[i+1:]...)
			break
		}
	}
	// Add to end
	c.accessOrder = append(c.accessOrder, key)
}

func (c *DetectorCache) evictOldest() {
	if len(c.accessOrder) == 0 {
		return
	}

	oldest := c.accessOrder[0]
	c.accessOrder = c.accessOrder[1:]

	delete(c.preprocessed, oldest)
	delete(c.contours, oldest)
	delete(c.symbols, oldest)
}

// loadAndValidateImage loads and validates the image file
func (d *ParallelDetector) loadAndValidateImage(imagePath string) (image.Image, error) {
	// Reuse the original detector's validation and loading logic
	return d.Detector.loadAndValidateImage(imagePath)
}

// validateResults validates the detection results
func (d *ParallelDetector) validateResults(symbols []*Symbol, imagePath string) error {
	// Reuse the original detector's validation logic
	return d.Detector.validateResults(symbols, imagePath)
}
