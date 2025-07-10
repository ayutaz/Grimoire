package detector

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

// Detector handles symbol detection from images
type Detector struct {
	minContourArea  int
	circleThreshold float64
	binaryThreshold uint8
	blurKernelSize  int
	adaptiveBlockSize int
	morphKernelSize int
}

// NewDetector creates a new detector with default settings
func NewDetector() *Detector {
	return &Detector{
		minContourArea:  100,
		circleThreshold: 0.7,  // Lower threshold to detect more circles
		binaryThreshold: 128,
		blurKernelSize:  5,
		adaptiveBlockSize: 11,
		morphKernelSize: 3,
	}
}

// DetectSymbols detects all symbols in the given image file
func DetectSymbols(imagePath string) ([]*Symbol, error) {
	detector := NewDetector()
	return detector.Detect(imagePath)
}

// Detect performs symbol detection on the image
func (d *Detector) Detect(imagePath string) ([]*Symbol, error) {
	// Open image file
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Convert to grayscale
	gray := d.toGrayscale(img)

	// Preprocess image
	binary := d.preprocessImage(gray)

	// Try to find outer circle in original grayscale image
	outerCircle := d.findOuterCircleFromGrayscale(gray)

	// Find contours
	contours := d.findContours(binary)
	
	// Add outer circle if found
	if outerCircle != nil {
		contours = append([]Contour{*outerCircle}, contours...)
	}

	// Debug: print contour information
	if os.Getenv("GRIMOIRE_DEBUG") != "" {
		d.DebugPrintContours(contours)
		// Save preprocessed image for debugging
		d.DebugSaveContours(binary, contours, "debug_binary.png")
	}

	// Detect symbols from contours
	symbols := d.detectSymbolsFromContours(contours, binary)

	// Detect connections
	// TODO: Implement connection detection

	return symbols, nil
}

// toGrayscale converts an image to grayscale
func (d *Detector) toGrayscale(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := img.At(x, y)
			grayColor := color.GrayModel.Convert(oldColor)
			gray.Set(x, y, grayColor)
		}
	}

	return gray
}



// detectSymbolsFromContours analyzes contours to identify symbols
func (d *Detector) detectSymbolsFromContours(contours []Contour, binary *image.Gray) []*Symbol {
	symbols := make([]*Symbol, 0)

	// First, look for the outer circle
	var outerCircle *Symbol
	for _, contour := range contours {
		if contour.Area < float64(d.minContourArea) {
			continue
		}

		// Classify contour
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
			symbols = append(symbols, outerCircle)
			break
		}
	}

	// Then detect other symbols
	for _, contour := range contours {
		if contour.Area < float64(d.minContourArea) {
			continue
		}

		// Skip if it's the outer circle
		symbolType := d.classifyContour(contour)
		if symbolType == OuterCircle || symbolType == Unknown {
			continue
		}

		// Detect internal pattern
		pattern := d.detectInternalPattern(contour, binary)

		symbol := &Symbol{
			Type:       symbolType,
			Position:   Position{X: float64(contour.Center.X), Y: float64(contour.Center.Y)},
			Size:       math.Sqrt(contour.Area),
			Confidence: 0.7,
			Pattern:    pattern,
			Properties: make(map[string]interface{}),
		}

		// Only add symbols within the outer circle if one exists
		if outerCircle != nil {
			centerDist := math.Sqrt(math.Pow(symbol.Position.X-outerCircle.Position.X, 2) + 
				math.Pow(symbol.Position.Y-outerCircle.Position.Y, 2))
			if centerDist < outerCircle.Size*0.9 {
				symbols = append(symbols, symbol)
			}
		} else {
			symbols = append(symbols, symbol)
		}
	}

	return symbols
}

// classifyContour determines the type of symbol from contour shape
func (d *Detector) classifyContour(contour Contour) SymbolType {
	return d.classifyShape(contour)
}


// preprocessImage applies preprocessing steps to improve detection
func (d *Detector) preprocessImage(gray *image.Gray) *image.Gray {
	// Apply Gaussian blur to reduce noise
	blurred := gaussianBlur(gray, d.blurKernelSize)
	
	// Apply adaptive threshold
	binary := adaptiveThreshold(blurred, d.adaptiveBlockSize, 2)
	
	// Apply morphological operations to clean up
	binary = morphologyClose(binary, d.morphKernelSize)
	binary = morphologyOpen(binary, d.morphKernelSize)
	
	return binary
}

// isOuterCircle checks if a contour is the outer circle
func (d *Detector) isOuterCircle(contour Contour) bool {
	// Check if it's circular
	if !contour.isCircle(d.circleThreshold) {
		return false
	}
	
	// Check if it's large enough relative to total contour area
	// The outer circle should be one of the largest contours
	// This check is done in classifyShape by checking relative size
	
	// Check aspect ratio
	if contour.getAspectRatio() > 1.2 {
		return false
	}
	
	return true
}

// findOuterCircleFromGrayscale attempts to find the outer circle from the original grayscale image
func (d *Detector) findOuterCircleFromGrayscale(gray *image.Gray) *Contour {
	bounds := gray.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	center := image.Point{X: width / 2, Y: height / 2}
	
	// Scan from center outward to find the circle
	maxRadius := min(width, height) / 2
	var circlePoints []image.Point
	
	// Sample points around the expected circle
	numSamples := 360
	for r := maxRadius - 50; r < maxRadius; r++ {
		blackPixels := 0
		var candidatePoints []image.Point
		
		for i := 0; i < numSamples; i++ {
			angle := float64(i) * 2 * math.Pi / float64(numSamples)
			x := center.X + int(float64(r)*math.Cos(angle))
			y := center.Y + int(float64(r)*math.Sin(angle))
			
			if x >= 0 && x < width && y >= 0 && y < height {
				// Check if pixel is dark (part of circle)
				if gray.GrayAt(x, y).Y < 128 {
					blackPixels++
					candidatePoints = append(candidatePoints, image.Point{X: x, Y: y})
				}
			}
		}
		
		// If we found a circle at this radius (most pixels are black)
		if float64(blackPixels) > float64(numSamples)*0.8 {
			circlePoints = candidatePoints
			break
		}
	}
	
	if len(circlePoints) > 100 {
		contour := Contour{Points: circlePoints}
		contour.calculateProperties()
		return &contour
	}
	
	return nil
}