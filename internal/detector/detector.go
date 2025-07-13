package detector

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // for jpeg image format support
	_ "image/png"  // for png image format support
	"math"
	"os"
	"path/filepath"
	"strings"

	grimoireErrors "github.com/ayutaz/grimoire/internal/errors"
	"github.com/ayutaz/grimoire/internal/security"
)

// Config holds detector configuration
type Config struct {
	Debug bool
}

// Detector handles symbol detection from images
type Detector struct {
	minContourArea    int
	circleThreshold   float64
	binaryThreshold   uint8
	blurKernelSize    int
	adaptiveBlockSize int
	morphKernelSize   int
	debug             bool
}

// NewDetector creates a new detector with default settings
func NewDetector(cfg Config) *Detector {
	return &Detector{
		minContourArea:    50,   // Lower to detect small stars
		circleThreshold:   0.85, // Higher threshold to distinguish squares from circles
		binaryThreshold:   128,
		blurKernelSize:    3, // Reduced blur to preserve edges
		adaptiveBlockSize: 11,
		morphKernelSize:   2, // Reduced to prevent breaking thin lines
		debug:             cfg.Debug,
	}
}

// DetectSymbols detects all symbols in the given image file
func DetectSymbols(imagePath string) ([]*Symbol, []Connection, error) {
	detector := NewDetector(Config{Debug: false})
	return detector.Detect(imagePath)
}

// Detect performs symbol detection on the image
func (d *Detector) Detect(imagePath string) ([]*Symbol, []Connection, error) {
	// Load and validate image
	img, err := d.loadAndValidateImage(imagePath)
	if err != nil {
		return nil, nil, err
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
		if err := d.DebugSaveContours(binary, contours, "debug_binary.png"); err != nil {
			fmt.Printf("Failed to save debug contours: %v\n", err)
		}
		// Print detailed info for specific contours
		fmt.Println("\nDebug: Checking contours in square regions:")
		for i, contour := range contours {
			// Check both square regions
			if (contour.Center.X > 350 && contour.Center.X < 400 &&
				contour.Center.Y > 170 && contour.Center.Y < 220) ||
				(contour.Center.X > 200 && contour.Center.X < 250 &&
					contour.Center.Y > 170 && contour.Center.Y < 220) {
				bbox := contour.getBoundingBox()
				// approximatePolygon is in shape_classifier.go
				vertices := len(contour.Points)
				fmt.Printf("[%d] Square region candidate:\n", i)
				fmt.Printf("  Center: (%d,%d)\n", contour.Center.X, contour.Center.Y)
				fmt.Printf("  BBox: (%d,%d,%d,%d), w=%d, h=%d\n",
					bbox.Min.X, bbox.Min.Y, bbox.Max.X, bbox.Max.Y,
					bbox.Dx(), bbox.Dy())
				fmt.Printf("  Area: %.1f, Perimeter: %.1f\n", contour.Area, contour.Perimeter)
				fmt.Printf("  Circularity: %.2f, Aspect: %.2f\n",
					contour.Circularity, contour.getAspectRatio())
				fmt.Printf("  Points: %d\n", vertices)
			}
		}
		fmt.Println()
	}

	// Detect symbols from contours
	symbols := d.detectSymbolsFromContours(contours, binary)

	// Deduplicate nearby stars
	symbols = d.deduplicateNearbyStars(symbols)

	// Detect connections
	connections := d.improvedDetectConnections(binary, symbols)

	// Validate detection results
	if err := d.validateResults(symbols, imagePath); err != nil {
		return nil, nil, err
	}

	return symbols, connections, nil
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
			if os.Getenv("GRIMOIRE_DEBUG") != "" && symbolType == Unknown {
				fmt.Printf("Unknown symbol at (%d,%d), area=%.2f, circularity=%.2f\n",
					contour.Center.X, contour.Center.Y, contour.Area, contour.Circularity)
			}
			continue
		}

		// Detect internal pattern for shapes that can contain patterns
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

		if os.Getenv("GRIMOIRE_DEBUG") != "" && pattern != "empty" {
			fmt.Printf("Symbol %s at (%d,%d) has pattern: %s\n",
				symbolType, contour.Center.X, contour.Center.Y, pattern)
		}

		// Only add symbols within the outer circle if one exists
		if outerCircle != nil {
			centerDist := math.Sqrt(math.Pow(symbol.Position.X-outerCircle.Position.X, 2) +
				math.Pow(symbol.Position.Y-outerCircle.Position.Y, 2))
			if centerDist < outerCircle.Size*0.9 {
				// For stars, only accept those near the center
				if symbolType == Star {
					if centerDist < outerCircle.Size*0.3 { // Within 30% of radius from center
						symbols = append(symbols, symbol)
					}
				} else {
					symbols = append(symbols, symbol)
				}
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

	// Apply adaptive threshold with adjusted constant
	binary := adaptiveThreshold(blurred, d.adaptiveBlockSize, 5) // Increased constant for better edge preservation

	// Apply morphological operations to clean up
	// Only apply closing to connect nearby components
	binary = morphologyClose(binary, d.morphKernelSize)
	// Skip opening to avoid breaking thin lines
	// binary = morphologyOpen(binary, d.morphKernelSize)

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

// deduplicateNearbyStars removes duplicate star detections
func (d *Detector) deduplicateNearbyStars(symbols []*Symbol) []*Symbol {
	filtered := []*Symbol{}

	// Group stars by proximity
	starGroups := [][]*Symbol{}
	for _, symbol := range symbols {
		if symbol.Type != Star {
			filtered = append(filtered, symbol)
			continue
		}

		// Check if this star belongs to an existing group
		addedToGroup := false
		for i, group := range starGroups {
			for _, existingStar := range group {
				dist := math.Sqrt(math.Pow(symbol.Position.X-existingStar.Position.X, 2) +
					math.Pow(symbol.Position.Y-existingStar.Position.Y, 2))
				if dist < 50 { // Within 50 pixels
					starGroups[i] = append(starGroups[i], symbol)
					addedToGroup = true
					break
				}
			}
			if addedToGroup {
				break
			}
		}

		if !addedToGroup {
			starGroups = append(starGroups, []*Symbol{symbol})
		}
	}

	// Keep only the largest star from each group
	for _, group := range starGroups {
		if len(group) > 0 {
			largest := group[0]
			for _, star := range group[1:] {
				if star.Size > largest.Size {
					largest = star
				}
			}
			filtered = append(filtered, largest)
		}
	}

	return filtered
}

// loadAndValidateImage loads and validates the image file with security checks
func (d *Detector) loadAndValidateImage(imagePath string) (image.Image, error) {
	// Create image validator with default settings
	validator := security.NewImageValidator()

	// Create safe image decoder
	decoder := security.NewSafeImageDecoder(validator)

	// Decode image with all security validations
	img, err := decoder.DecodeImage(imagePath)
	if err != nil {
		// Convert security errors to grimoire errors for consistency
		errStr := err.Error()

		// Check for file not found errors
		if strings.Contains(errStr, "file not found") || os.IsNotExist(err) {
			return nil, grimoireErrors.FileNotFoundError(imagePath)
		}
		if strings.Contains(errStr, "unsupported file extension") || strings.Contains(errStr, "unsupported file format") {
			ext := filepath.Ext(imagePath)
			return nil, grimoireErrors.UnsupportedFormatError(ext).
				WithDetails(fmt.Sprintf("File: %s", filepath.Base(imagePath)))
		}

		if strings.Contains(errStr, "path traversal") {
			// Don't expose the actual path in error message for security
			safeFileName := filepath.Base(imagePath)
			if strings.Contains(safeFileName, "..") {
				safeFileName = "invalid-path"
			}
			return nil, grimoireErrors.NewError(grimoireErrors.ValidationError, "Invalid file path detected").
				WithLocation(safeFileName, 0, 0).
				WithSuggestion("Use a valid file path without directory traversal attempts")
		}

		if strings.Contains(errStr, "exceeds maximum") || strings.Contains(errStr, "exceeds safe limits") {
			return nil, grimoireErrors.NewError(grimoireErrors.ValidationError, "Image exceeds size limits").
				WithInnerError(err).
				WithLocation(imagePath, 0, 0).
				WithSuggestion("Use a smaller image (max 50MB file size, 10000x10000 pixels)")
		}

		// Check for permission errors
		if strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "access is denied") {
			return nil, grimoireErrors.NewError(grimoireErrors.FileReadError, "Failed to read image file").
				WithInnerError(err).
				WithLocation(imagePath, 0, 0)
		}

		// Generic image processing error
		return nil, grimoireErrors.NewError(grimoireErrors.ImageProcessingError, "Failed to validate and decode image").
			WithInnerError(err).
			WithLocation(imagePath, 0, 0).
			WithSuggestion("Ensure the image is a valid PNG or JPEG file and not corrupted")
	}

	return img, nil
}

// validateResults validates the detection results
func (d *Detector) validateResults(symbols []*Symbol, imagePath string) error {
	if len(symbols) == 0 {
		return grimoireErrors.NoSymbolsError().
			WithLocation(imagePath, 0, 0)
	}

	// Check for outer circle
	hasOuterCircle := false
	for _, sym := range symbols {
		if sym.Type == OuterCircle {
			hasOuterCircle = true
			break
		}
	}

	if !hasOuterCircle {
		return grimoireErrors.NoOuterCircleError().
			WithLocation(imagePath, 0, 0)
	}

	return nil
}

// hasLineBetween checks if there's a line between two positions in the image
func (d *Detector) hasLineBetween(binary *image.Gray, from, to Position) bool {
	// Convert positions to image points
	x1, y1 := int(from.X), int(from.Y)
	x2, y2 := int(to.X), int(to.Y)

	// Calculate distance
	dx := x2 - x1
	dy := y2 - y1
	distance := math.Sqrt(float64(dx*dx + dy*dy))

	// Sample points along the line
	numSamples := int(distance / 2.0) // Sample every 2 pixels
	if numSamples < 10 {
		numSamples = 10
	}

	blackPixels := 0
	for i := 0; i <= numSamples; i++ {
		t := float64(i) / float64(numSamples)
		x := x1 + int(t*float64(dx))
		y := y1 + int(t*float64(dy))

		// Check if pixel is within bounds
		if x >= 0 && x < binary.Bounds().Dx() && y >= 0 && y < binary.Bounds().Dy() {
			if binary.GrayAt(x, y).Y == 0 {
				blackPixels++
			}
		}
	}

	// Consider it a line if more than 60% of sampled pixels are black
	return float64(blackPixels) > float64(numSamples)*0.6
}
