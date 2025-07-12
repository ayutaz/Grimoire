package security

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

func createTestPNG(width, height int) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 0, 255})
		}
	}

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func TestSafeImageDecoder_DecodeImage(t *testing.T) {
	validator := NewImageValidator()
	validator.MaxImageWidth = 100
	validator.MaxImageHeight = 100
	validator.MaxFileSize = 10 * 1024 // 10KB

	decoder := NewSafeImageDecoder(validator)
	tempDir := t.TempDir()
	validator.WorkingDirectory = tempDir

	// Create test images
	smallImage, err := createTestPNG(50, 50)
	require.NoError(t, err)

	largeImage, err := createTestPNG(200, 200)
	require.NoError(t, err)

	tests := []struct {
		name        string
		setupFile   func() string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid small image",
			setupFile: func() string {
				path := filepath.Join(tempDir, "small.png")
				err := os.WriteFile(path, smallImage, 0644)
				require.NoError(t, err)
				return path
			},
			expectError: false,
		},
		{
			name: "Image exceeds width limit",
			setupFile: func() string {
				path := filepath.Join(tempDir, "wide.png")
				err := os.WriteFile(path, largeImage, 0644)
				require.NoError(t, err)
				return path
			},
			expectError: true,
			errorMsg:    "exceeds maximum allowed width",
		},
		{
			name: "Non-existent file",
			setupFile: func() string {
				return filepath.Join(tempDir, "nonexistent.png")
			},
			expectError: true,
			errorMsg:    "file not found",
		},
		{
			name: "Invalid image data",
			setupFile: func() string {
				path := filepath.Join(tempDir, "invalid.png")
				// PNG header but invalid content
				invalidData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
				invalidData = append(invalidData, []byte("invalid image data")...)
				err := os.WriteFile(path, invalidData, 0644)
				require.NoError(t, err)
				return path
			},
			expectError: true,
			errorMsg:    "failed to decode",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filePath := tc.setupFile()
			img, err := decoder.DecodeImage(filePath)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, img)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, img)
			}
		})
	}
}

func TestSafeImageDecoder_MemoryLimitCheck(t *testing.T) {
	validator := NewImageValidator()
	// Set limits that would allow a 100x100 image but not much larger
	validator.MaxImageWidth = 200
	validator.MaxImageHeight = 200
	validator.MaxFileSize = 1024 // 1KB - very small to trigger memory check

	decoder := NewSafeImageDecoder(validator)
	tempDir := t.TempDir()
	validator.WorkingDirectory = tempDir

	// Create a small but valid PNG that will be corrupted
	img, err := createTestPNG(10, 10)
	require.NoError(t, err)

	path := filepath.Join(tempDir, "memory_test.png")
	if len(img) > 1024 {
		// Truncate to create corrupted file
		err = os.WriteFile(path, img[:1024], 0644)
	} else {
		// If image is smaller than 1KB, corrupt it differently
		corrupted := append(img[:len(img)/2], []byte("corrupted")...)
		err = os.WriteFile(path, corrupted, 0644)
	}
	require.NoError(t, err)

	// This should fail due to corrupted data
	_, err = decoder.DecodeImage(path)
	assert.Error(t, err)
}
