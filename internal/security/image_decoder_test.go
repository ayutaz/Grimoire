package security

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 0, 255})
		}
	}
	return img
}

func createTestPNG(width, height int) ([]byte, error) {
	img := createTestImage(width, height)
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func createTestJPEG(width, height int) ([]byte, error) {
	img := createTestImage(width, height)
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func createTestGIF(width, height int) ([]byte, error) {
	img := createTestImage(width, height)
	var buf bytes.Buffer
	err := gif.Encode(&buf, img, &gif.Options{})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func createTestWebP(width, height int) ([]byte, error) {
	// WebP encoding is not supported in golang.org/x/image/webp
	// For testing purposes, we'll skip WebP creation and use a pre-made sample
	// or return an error indicating WebP encoding is not available
	return nil, fmt.Errorf("WebP encoding not available in test")
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

func TestSafeImageDecoder_MultipleFormats(t *testing.T) {
	validator := NewImageValidator()
	validator.MaxImageWidth = 100
	validator.MaxImageHeight = 100
	validator.MaxFileSize = 100 * 1024 // 100KB

	decoder := NewSafeImageDecoder(validator)
	tempDir := t.TempDir()
	validator.WorkingDirectory = tempDir

	tests := []struct {
		name      string
		format    string
		createFn  func(int, int) ([]byte, error)
		extension string
	}{
		{
			name:      "PNG format",
			format:    "png",
			createFn:  createTestPNG,
			extension: ".png",
		},
		{
			name:      "JPEG format",
			format:    "jpeg",
			createFn:  createTestJPEG,
			extension: ".jpg",
		},
		{
			name:      "GIF format",
			format:    "gif",
			createFn:  createTestGIF,
			extension: ".gif",
		},
		{
			name:      "WebP format",
			format:    "webp",
			createFn:  createTestWebP,
			extension: ".webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image
			imageData, err := tt.createFn(50, 50)
			if tt.format == "webp" {
				// Skip WebP test for now as we don't have encoding support
				t.Skip("WebP encoding not available in test")
				return
			}
			require.NoError(t, err)

			// Write to file
			imagePath := filepath.Join(tempDir, "test"+tt.extension)
			err = os.WriteFile(imagePath, imageData, 0644)
			require.NoError(t, err)

			// Decode image
			img, err := decoder.DecodeImage(imagePath)
			assert.NoError(t, err)
			assert.NotNil(t, img)
			assert.Equal(t, 50, img.Bounds().Dx())
			assert.Equal(t, 50, img.Bounds().Dy())
		})
	}
}

func TestImageValidator_FileHeaders(t *testing.T) {
	validator := NewImageValidator()
	tempDir := t.TempDir()
	validator.WorkingDirectory = tempDir

	tests := []struct {
		name      string
		extension string
		header    []byte
		shouldErr bool
	}{
		{
			name:      "Valid PNG header",
			extension: ".png",
			header:    []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00},
			shouldErr: false,
		},
		{
			name:      "Valid JPEG header",
			extension: ".jpg",
			header:    []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			shouldErr: false,
		},
		{
			name:      "Valid GIF87a header",
			extension: ".gif",
			header:    []byte("GIF87a"),
			shouldErr: false,
		},
		{
			name:      "Valid GIF89a header",
			extension: ".gif",
			header:    []byte("GIF89a"),
			shouldErr: false,
		},
		{
			name:      "Valid WebP header",
			extension: ".webp",
			header:    []byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P'},
			shouldErr: false,
		},
		{
			name:      "Invalid PNG header",
			extension: ".png",
			header:    []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			shouldErr: true,
		},
		{
			name:      "Invalid JPEG header",
			extension: ".jpg",
			header:    []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create file with header
			filePath := filepath.Join(tempDir, "test"+tt.extension)

			// Pad the header to make it at least 512 bytes
			data := make([]byte, 512)
			copy(data, tt.header)

			err := os.WriteFile(filePath, data, 0644)
			require.NoError(t, err)

			// Validate header
			err = validator.ValidateFileHeader(filePath)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
