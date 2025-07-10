package detector

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectorCreation tests that we can create a new detector
func TestDetectorCreation(t *testing.T) {
	detector := NewDetector()
	assert.NotNil(t, detector)
	assert.Equal(t, 100, detector.minContourArea)
	assert.Equal(t, 0.8, detector.circleThreshold)
}

// TestDetectSymbols_NoOuterCircle tests detection fails without outer circle
func TestDetectSymbols_NoOuterCircle(t *testing.T) {
	// Create a test image without outer circle
	img := createTestImage(100, 100)
	path := saveTestImage(t, img, "no_outer_circle.png")
	defer os.Remove(path)

	symbols, err := DetectSymbols(path)
	
	// Should detect something but no outer circle
	assert.NoError(t, err)
	assert.NotNil(t, symbols)
	
	// Check that no outer circle is detected
	hasOuterCircle := false
	for _, symbol := range symbols {
		if symbol.Type == OuterCircle {
			hasOuterCircle = true
			break
		}
	}
	assert.False(t, hasOuterCircle, "Should not detect outer circle in empty image")
}

// TestDetectSymbols_MinimalProgram tests detection of minimal valid program
func TestDetectSymbols_MinimalProgram(t *testing.T) {
	// Create a test image with just an outer circle
	img := createTestImageWithCircle(200, 200, 90)
	path := saveTestImage(t, img, "minimal_program.png")
	defer os.Remove(path)

	symbols, err := DetectSymbols(path)
	
	require.NoError(t, err)
	require.NotEmpty(t, symbols)
	
	// Should detect at least the outer circle
	hasOuterCircle := false
	for _, symbol := range symbols {
		if symbol.Type == OuterCircle {
			hasOuterCircle = true
			assert.InDelta(t, 100, symbol.Position.X, 10)
			assert.InDelta(t, 100, symbol.Position.Y, 10)
			assert.True(t, symbol.Confidence > 0.7)
			break
		}
	}
	assert.True(t, hasOuterCircle, "Should detect outer circle")
}

// TestToGrayscale tests grayscale conversion
func TestToGrayscale(t *testing.T) {
	detector := NewDetector()
	
	// Create a color image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Set a red pixel
	img.Set(5, 5, color.RGBA{255, 0, 0, 255})
	
	gray := detector.toGrayscale(img)
	
	assert.NotNil(t, gray)
	assert.Equal(t, img.Bounds(), gray.Bounds())
	
	// Check that the red pixel was converted to gray
	grayPixel := gray.GrayAt(5, 5)
	assert.Greater(t, grayPixel.Y, uint8(0))
}

// TestThreshold tests binary threshold
func TestThreshold(t *testing.T) {
	detector := NewDetector()
	
	// Create a grayscale image with gradient
	gray := image.NewGray(image.Rect(0, 0, 10, 10))
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			gray.Set(x, y, color.Gray{uint8(x * 25)})
		}
	}
	
	binary := detector.threshold(gray)
	
	assert.NotNil(t, binary)
	assert.Equal(t, gray.Bounds(), binary.Bounds())
	
	// Check threshold effect
	// Left side should be black, right side should be white
	assert.Equal(t, color.Gray{0}, binary.GrayAt(0, 5))
	assert.Equal(t, color.Gray{255}, binary.GrayAt(9, 5))
}

// Helper functions

func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with white background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	return img
}

func createTestImageWithCircle(width, height, radius int) *image.RGBA {
	img := createTestImage(width, height)
	
	// Draw a black circle
	centerX, centerY := width/2, height/2
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			distance := dx*dx + dy*dy
			radiusSq := float64(radius * radius)
			
			// Draw circle outline (thickness ~5 pixels)
			if distance >= radiusSq-float64(radius*10) && distance <= radiusSq+float64(radius*10) {
				img.Set(x, y, color.Black)
			}
		}
	}
	
	return img
}

func saveTestImage(t *testing.T, img image.Image, filename string) string {
	t.Helper()
	
	// Create temp directory
	dir := t.TempDir()
	path := filepath.Join(dir, filename)
	
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()
	
	// For now, save as PNG using a simple format
	// In real implementation, we'd use image/png
	err = savePNG(file, img)
	require.NoError(t, err)
	
	return path
}

// savePNG saves image as PNG
func savePNG(file *os.File, img image.Image) error {
	return png.Encode(file, img)
}