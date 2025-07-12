package detector

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestPNGFile(t *testing.T, path string, width, height int) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with white background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}

	// Draw a simple black circle (outer circle) if image is large enough
	if width >= 20 && height >= 20 {
		centerX, centerY := width/2, height/2
		radius := min(width, height)/2 - 5

		// Simple circle drawing
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dx := x - centerX
				dy := y - centerY
				distSq := dx*dx + dy*dy
				radiusSq := radius * radius
				innerRadiusSq := (radius - 3) * (radius - 3)

				// Draw circle outline (ring)
				if distSq <= radiusSq && distSq >= innerRadiusSq {
					img.Set(x, y, color.RGBA{0, 0, 0, 255})
				}
			}
		}
	}

	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	err = png.Encode(file, img)
	require.NoError(t, err)
}

func TestDetector_SecurityValidation(t *testing.T) {
	detector := NewDetector(Config{})
	tempDir := t.TempDir()

	t.Run("Path traversal attack prevention", func(t *testing.T) {
		// Try various path traversal attempts
		attacks := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\config\\sam",
			"images/../../../etc/shadow",
			"./././../../../etc/hosts",
		}

		for _, attack := range attacks {
			t.Run(attack, func(t *testing.T) {
				_, _, err := detector.Detect(attack)
				assert.Error(t, err)
				// Error message should exist but not reveal full system paths
				// Just check that we got an error, not the specific content
			})
		}
	})

	t.Run("File size limit enforcement", func(t *testing.T) {
		// Create a file larger than 50MB limit
		hugePath := filepath.Join(tempDir, "huge.png")
		file, err := os.Create(hugePath)
		require.NoError(t, err)

		// Write PNG header
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		_, err = file.Write(pngHeader)
		require.NoError(t, err)

		// Write 51MB of data
		data := make([]byte, 1024*1024) // 1MB buffer
		for i := 0; i < 51; i++ {
			_, err = file.Write(data)
			require.NoError(t, err)
		}
		file.Close()

		_, _, err = detector.Detect(hugePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "size limits")
	})

	t.Run("Image dimension limits", func(t *testing.T) {
		// This test would require creating an actual large PNG
		// For now, we'll test that the validation is in place
		// In a real scenario, you'd create a 10001x10001 image
		t.Skip("Skipping large image test to save resources")
	})

	t.Run("Invalid file format with correct extension", func(t *testing.T) {
		// Create a file with .png extension but invalid content
		fakePNG := filepath.Join(tempDir, "fake.png")
		err := os.WriteFile(fakePNG, []byte("This is not a PNG file"), 0644)
		require.NoError(t, err)

		_, _, err = detector.Detect(fakePNG)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a valid PNG")
	})

	t.Run("JPEG file with PNG extension", func(t *testing.T) {
		// Create a file with .png extension but JPEG content
		jpegAsPNG := filepath.Join(tempDir, "jpeg_as_png.png")
		jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0}
		err := os.WriteFile(jpegAsPNG, append(jpegHeader, make([]byte, 100)...), 0644)
		require.NoError(t, err)

		_, _, err = detector.Detect(jpegAsPNG)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a valid PNG")
	})

	t.Run("Valid small image", func(t *testing.T) {
		// Create a valid small PNG
		validPath := filepath.Join(tempDir, "valid.png")
		createTestPNGFile(t, validPath, 100, 100)

		// This should succeed
		symbols, connections, err := detector.Detect(validPath)
		assert.NoError(t, err)
		assert.NotNil(t, symbols)
		assert.NotNil(t, connections)
	})
}

func TestDetector_MaliciousInputs(t *testing.T) {
	detector := NewDetector(Config{})
	tempDir := t.TempDir()

	t.Run("Null bytes in filename", func(t *testing.T) {
		// Filenames with null bytes can be used for attacks
		nullPath := "test\x00.png"
		_, _, err := detector.Detect(nullPath)
		assert.Error(t, err)
	})

	t.Run("Extremely long filename", func(t *testing.T) {
		// Create a filename with 1000 characters
		longName := filepath.Join(tempDir, string(bytes.Repeat([]byte("a"), 1000))+".png")
		_, _, err := detector.Detect(longName)
		assert.Error(t, err)
	})

	t.Run("Unicode tricks in filename", func(t *testing.T) {
		// Some unicode characters can be used to hide file extensions
		trickyPath := filepath.Join(tempDir, "test\u202e.png.exe")
		_, _, err := detector.Detect(trickyPath)
		assert.Error(t, err)
	})
}

func TestDetector_ResourceExhaustion(t *testing.T) {
	detector := NewDetector(Config{})
	tempDir := t.TempDir()

	t.Run("ZIP bomb prevention", func(t *testing.T) {
		// A PNG that expands to huge size when decoded
		// This is simulated by checking memory limits
		t.Skip("ZIP bomb test requires special crafted image")
	})

	t.Run("Multiple concurrent requests", func(t *testing.T) {
		// Create a valid small PNG
		validPath := filepath.Join(tempDir, "concurrent.png")
		createTestPNGFile(t, validPath, 50, 50)

		// Run multiple detections concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_, _, _ = detector.Detect(validPath)
				// Ignore error for concurrent test - just checking for crashes
				done <- true
			}()
		}

		// Wait for all to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
