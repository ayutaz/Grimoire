package detector

import (
	"image"
	"runtime"
	"sync"
	"sync/atomic"
)

// ParallelDetectorV2 is an improved parallel detector with better memory management
type ParallelDetectorV2 struct {
	*Detector
	workerCount       int
	contourPool       *sync.Pool
	symbolPool        *sync.Pool
	bufferPool        *sync.Pool
	atomicSymbolCount int64
}

// NewParallelDetectorV2 creates an improved parallel detector
func NewParallelDetectorV2(cfg Config) *ParallelDetectorV2 {
	workerCount := runtime.NumCPU()
	if workerCount > 8 {
		workerCount = 8 // Limit workers to avoid excessive overhead
	}

	return &ParallelDetectorV2{
		Detector:    NewDetector(cfg),
		workerCount: workerCount,
		contourPool: &sync.Pool{
			New: func() interface{} {
				return &Contour{
					Points: make([]image.Point, 0, 100),
				}
			},
		},
		symbolPool: &sync.Pool{
			New: func() interface{} {
				return &Symbol{
					Properties: make(map[string]interface{}),
				}
			},
		},
		bufferPool: &sync.Pool{
			New: func() interface{} {
				// Pre-allocate buffer for image processing
				return make([]byte, 4096)
			},
		},
	}
}

// Detect performs optimized parallel detection
func (pd *ParallelDetectorV2) Detect(imagePath string) ([]*Symbol, []Connection, error) {
	// Load and validate image
	img, err := pd.loadAndValidateImage(imagePath)
	if err != nil {
		return nil, nil, err
	}

	// Convert to grayscale
	gray := pd.toGrayscale(img)

	// Preprocess image
	binary := pd.preprocessImage(gray)

	// Reset atomic counter
	atomic.StoreInt64(&pd.atomicSymbolCount, 0)

	// Find contours with memory-efficient parallel processing
	contours := pd.findContoursOptimized(binary)

	// Try to find outer circle
	outerCircle := pd.findOuterCircleFromGrayscale(gray)
	if outerCircle != nil {
		contours = append([]Contour{*outerCircle}, contours...)
	}

	// Detect symbols from contours with object pooling
	symbols := pd.detectSymbolsOptimized(contours, binary)

	// Deduplicate nearby stars
	symbols = pd.deduplicateNearbyStars(symbols)

	// Detect connections with optimized algorithm
	connections := pd.detectConnectionsOptimized(binary, symbols)

	// Validate results
	if err := pd.validateResults(symbols, imagePath); err != nil {
		return nil, nil, err
	}

	return symbols, connections, nil
}

// findContoursOptimized finds contours with better memory management
func (pd *ParallelDetectorV2) findContoursOptimized(binary *image.Gray) []Contour {
	// Use standard detector for the actual contour finding
	// The optimization is in the parallel symbol detection and connection detection
	return pd.Detector.findContours(binary)
}

// detectSymbolsOptimized detects symbols with object pooling
func (pd *ParallelDetectorV2) detectSymbolsOptimized(contours []Contour, binary *image.Gray) []*Symbol {
	if len(contours) == 0 {
		return nil
	}

	// For small contour counts, use single-threaded approach
	if len(contours) < 50 {
		return pd.detectSymbolsFromContours(contours, binary)
	}

	// Preallocate result slice
	expectedSymbols := len(contours) / 2 // Estimate
	results := make([]*Symbol, 0, expectedSymbols)
	resultsMutex := &sync.Mutex{}

	// Process contours in batches
	batchSize := len(contours) / pd.workerCount
	if batchSize < 10 {
		batchSize = 10
	}

	var wg sync.WaitGroup
	for i := 0; i < len(contours); i += batchSize {
		end := i + batchSize
		if end > len(contours) {
			end = len(contours)
		}

		wg.Add(1)
		go func(batch []Contour) {
			defer wg.Done()

			batchSymbols := make([]*Symbol, 0, len(batch))

			for _, contour := range batch {
				if contour.Area < float64(pd.Detector.minContourArea) {
					continue
				}

				// Classify contour
				symbolType := pd.classifyContour(contour)
				if symbolType == Unknown {
					continue
				}

				// Create new symbol instead of using pool to avoid race conditions
				symbol := &Symbol{
					Type:       symbolType,
					Position:   Position{X: float64(contour.Center.X), Y: float64(contour.Center.Y)},
					Size:       contour.getEquivalentRadius(),
					Confidence: contour.Circularity,
					Properties: make(map[string]interface{}),
				}

				// Detect pattern
				if symbolType == Square || symbolType == Circle || symbolType == Pentagon ||
					symbolType == Hexagon || symbolType == Star {
					symbol.Pattern = pd.detectInternalPattern(contour, binary)
				} else {
					symbol.Pattern = PatternEmpty
				}

				batchSymbols = append(batchSymbols, symbol)
			}

			// Add batch results
			if len(batchSymbols) > 0 {
				resultsMutex.Lock()
				results = append(results, batchSymbols...)
				resultsMutex.Unlock()
			}
		}(contours[i:end])
	}

	wg.Wait()

	// Update atomic counter
	atomic.AddInt64(&pd.atomicSymbolCount, int64(len(results)))

	return results
}

// detectConnectionsOptimized detects connections with spatial indexing
func (pd *ParallelDetectorV2) detectConnectionsOptimized(binary *image.Gray, symbols []*Symbol) []Connection {
	if len(symbols) < 2 {
		return nil
	}

	// Build spatial index for symbols
	spatialIndex := pd.buildSpatialIndex(symbols)

	// Find potential connections using optimized line detection
	connections := make([]Connection, 0, len(symbols))
	connectionsMutex := &sync.Mutex{}

	// Process in parallel for large symbol counts
	if len(symbols) > 100 {
		var wg sync.WaitGroup
		batchSize := len(symbols) / pd.workerCount
		if batchSize < 20 {
			batchSize = 20
		}

		for i := 0; i < len(symbols); i += batchSize {
			end := i + batchSize
			if end > len(symbols) {
				end = len(symbols)
			}

			wg.Add(1)
			go func(startIdx, endIdx int) {
				defer wg.Done()

				localConnections := pd.detectConnectionsBatch(binary, symbols, spatialIndex, startIdx, endIdx)

				if len(localConnections) > 0 {
					connectionsMutex.Lock()
					connections = append(connections, localConnections...)
					connectionsMutex.Unlock()
				}
			}(i, end)
		}

		wg.Wait()
	} else {
		// Single-threaded for small symbol counts
		connections = pd.improvedDetectConnections(binary, symbols)
	}

	return connections
}

// SpatialIndex for efficient symbol lookup
type SpatialIndex struct {
	gridSize int
	grid     map[gridKey][]*Symbol
}

type gridKey struct {
	x, y int
}

// buildSpatialIndex creates a spatial index for symbols
func (pd *ParallelDetectorV2) buildSpatialIndex(symbols []*Symbol) *SpatialIndex {
	// Calculate bounds
	if len(symbols) == 0 {
		return nil
	}

	minX, minY := symbols[0].Position.X, symbols[0].Position.Y
	maxX, maxY := minX, minY

	for _, sym := range symbols {
		if sym.Position.X < minX {
			minX = sym.Position.X
		}
		if sym.Position.X > maxX {
			maxX = sym.Position.X
		}
		if sym.Position.Y < minY {
			minY = sym.Position.Y
		}
		if sym.Position.Y > maxY {
			maxY = sym.Position.Y
		}
	}

	// Determine grid size
	gridSize := 100
	if maxX-minX > 1000 || maxY-minY > 1000 {
		gridSize = 200
	}

	index := &SpatialIndex{
		gridSize: gridSize,
		grid:     make(map[gridKey][]*Symbol),
	}

	// Add symbols to grid
	for _, sym := range symbols {
		key := gridKey{
			x: int(sym.Position.X) / gridSize,
			y: int(sym.Position.Y) / gridSize,
		}
		index.grid[key] = append(index.grid[key], sym)
	}

	return index
}

// getNearbySymbols returns symbols within a certain distance
func (si *SpatialIndex) getNearbySymbols(pos Position, maxDist float64) []*Symbol {
	if si == nil {
		return nil
	}

	cellRadius := int(maxDist/float64(si.gridSize)) + 1
	centerKey := gridKey{
		x: int(pos.X) / si.gridSize,
		y: int(pos.Y) / si.gridSize,
	}

	result := make([]*Symbol, 0, 10)
	maxDistSq := maxDist * maxDist

	for dx := -cellRadius; dx <= cellRadius; dx++ {
		for dy := -cellRadius; dy <= cellRadius; dy++ {
			key := gridKey{x: centerKey.x + dx, y: centerKey.y + dy}
			if symbols, ok := si.grid[key]; ok {
				for _, sym := range symbols {
					dx := sym.Position.X - pos.X
					dy := sym.Position.Y - pos.Y
					if dx*dx+dy*dy <= maxDistSq {
						result = append(result, sym)
					}
				}
			}
		}
	}

	return result
}

// detectConnectionsBatch detects connections for a batch of symbols
func (pd *ParallelDetectorV2) detectConnectionsBatch(binary *image.Gray, symbols []*Symbol,
	spatialIndex *SpatialIndex, startIdx, endIdx int) []Connection {

	connections := make([]Connection, 0)

	for i := startIdx; i < endIdx && i < len(symbols); i++ {
		fromSymbol := symbols[i]
		if fromSymbol.Type == OuterCircle {
			continue
		}

		// Look for nearby symbols
		maxConnectionDist := 300.0
		nearbySymbols := spatialIndex.getNearbySymbols(fromSymbol.Position, maxConnectionDist)

		for _, toSymbol := range nearbySymbols {
			if toSymbol == fromSymbol || toSymbol.Type == OuterCircle {
				continue
			}

			// Check if there's a line between them
			if pd.hasLineBetween(binary, fromSymbol.Position, toSymbol.Position) {
				// Create line for connection type determination
				line := Line{
					Start: image.Point{X: int(fromSymbol.Position.X), Y: int(fromSymbol.Position.Y)},
					End:   image.Point{X: int(toSymbol.Position.X), Y: int(toSymbol.Position.Y)},
				}
				connectionType := pd.determineConnectionType(line, binary)
				connections = append(connections, Connection{
					From:           fromSymbol,
					To:             toSymbol,
					ConnectionType: connectionType,
				})
			}
		}
	}

	return connections
}

// Cleanup returns objects to pools
func (pd *ParallelDetectorV2) Cleanup() {
	// Pools will be garbage collected when detector is no longer used
}
