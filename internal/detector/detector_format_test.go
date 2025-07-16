package detector

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestImageWithShapes creates a test image with basic shapes
func createTestImageWithShapes(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with white background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw a simple circle in the center
	centerX, centerY := width/2, height/2
	radius := 50

	for y := centerY - radius; y <= centerY+radius; y++ {
		for x := centerX - radius; x <= centerX+radius; x++ {
			dx, dy := x-centerX, y-centerY
			if dx*dx+dy*dy <= radius*radius {
				// Draw a thick circle outline
				if dx*dx+dy*dy >= (radius-5)*(radius-5) {
					img.Set(x, y, color.Black)
				}
			}
		}
	}

	// Draw a star inside the circle
	starPoints := 5
	outerRadius := 30
	innerRadius := 15

	for i := 0; i < starPoints*2; i++ {
		angle1 := float64(i) * 3.14159 / float64(starPoints)
		angle2 := float64(i+1) * 3.14159 / float64(starPoints)

		var r1, r2 int
		if i%2 == 0 {
			r1, r2 = outerRadius, innerRadius
		} else {
			r1, r2 = innerRadius, outerRadius
		}

		x1 := centerX + int(float64(r1)*math.Cos(angle1))
		y1 := centerY + int(float64(r1)*math.Sin(angle1))
		x2 := centerX + int(float64(r2)*math.Cos(angle2))
		y2 := centerY + int(float64(r2)*math.Sin(angle2))

		// Draw line (simplified)
		drawLineFormat(img, x1, y1, x2, y2, color.Black)
	}

	return img
}

func drawLineFormat(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx := absFormat(x1 - x0)
	dy := absFormat(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func absFormat(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestDetector_MultipleImageFormats(t *testing.T) {
	detector := NewDetector(Config{Debug: false})
	tempDir := t.TempDir()

	// Create test image
	img := createTestImageWithShapes(300, 300)

	tests := []struct {
		name      string
		format    string
		encoder   func(*image.RGBA, string) error
		extension string
	}{
		{
			name:      "PNG format",
			format:    "png",
			extension: ".png",
			encoder: func(img *image.RGBA, path string) error {
				f, err := os.Create(path)
				if err != nil {
					return err
				}
				defer f.Close()
				return png.Encode(f, img)
			},
		},
		{
			name:      "JPEG format",
			format:    "jpeg",
			extension: ".jpg",
			encoder: func(img *image.RGBA, path string) error {
				f, err := os.Create(path)
				if err != nil {
					return err
				}
				defer f.Close()
				return jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
			},
		},
		{
			name:      "GIF format",
			format:    "gif",
			extension: ".gif",
			encoder: func(img *image.RGBA, path string) error {
				f, err := os.Create(path)
				if err != nil {
					return err
				}
				defer f.Close()
				return gif.Encode(f, img, &gif.Options{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save image in the specific format
			imagePath := filepath.Join(tempDir, "test"+tt.extension)
			err := tt.encoder(img, imagePath)
			require.NoError(t, err)

			// Detect symbols from the saved image
			symbols, _, err := detector.Detect(imagePath)
			assert.NoError(t, err)

			// We should detect at least the circle
			assert.NotEmpty(t, symbols, "Should detect at least one symbol in %s format", tt.format)

			// Check if we can also detect from bytes
			data, err := os.ReadFile(imagePath)
			require.NoError(t, err)

			symbolsFromBytes, _, err := detector.DetectFromBytes(data)
			assert.NoError(t, err)
			assert.Equal(t, len(symbols), len(symbolsFromBytes), "Should detect same number of symbols from file and bytes")
		})
	}
}

func TestDetector_WebPFormat(t *testing.T) {
	// WebP test requires a pre-made WebP file since we can't encode WebP
	t.Skip("WebP encoding not available, would need a pre-made WebP test file")
}

func TestDetector_UnsupportedFormat(t *testing.T) {
	detector := NewDetector(Config{Debug: false})
	tempDir := t.TempDir()

	// Create a file with BMP header (unsupported by default Go image)
	bmpPath := filepath.Join(tempDir, "test.bmp")
	bmpHeader := []byte{'B', 'M'} // BMP header
	err := os.WriteFile(bmpPath, bmpHeader, 0644)
	require.NoError(t, err)

	// Should fail to decode
	_, _, err = detector.Detect(bmpPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode image")
}

// TestDetector_ActualGrimoireImages tests with actual Grimoire pattern images
func TestDetector_ActualGrimoireImages(t *testing.T) {
	detector := NewDetector(Config{Debug: false})

	// Get absolute path for sample directory
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Test with existing sample images
	sampleDir := filepath.Join(cwd, "../../web/static/samples")
	samples := []string{"hello-world.png", "calculator.png", "fibonacci.png", "loop.png"}

	for _, sample := range samples {
		t.Run(sample, func(t *testing.T) {
			imagePath := filepath.Join(sampleDir, sample)

			// Skip if file doesn't exist
			if _, err := os.Stat(imagePath); os.IsNotExist(err) {
				t.Skipf("Sample image %s not found", sample)
			}

			symbols, connections, err := detector.Detect(imagePath)
			assert.NoError(t, err)
			assert.NotEmpty(t, symbols, "Should detect symbols in %s", sample)

			// Log detected symbols for debugging
			t.Logf("Detected %d symbols and %d connections in %s", len(symbols), len(connections), sample)
			for i, sym := range symbols {
				t.Logf("  Symbol %d: Type=%s, Position=(%.0f,%.0f), Pattern=%s",
					i, sym.Type, sym.Position.X, sym.Position.Y, sym.Pattern)
			}
		})
	}
}
