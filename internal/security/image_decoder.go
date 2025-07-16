package security

import (
	"fmt"
	"image"
	_ "image/gif"  // Register GIF decoder
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"os"

	_ "golang.org/x/image/webp" // Register WebP decoder
)

// SafeImageDecoder provides secure image decoding with size validation
type SafeImageDecoder struct {
	validator *ImageValidator
}

// NewSafeImageDecoder creates a new SafeImageDecoder
func NewSafeImageDecoder(validator *ImageValidator) *SafeImageDecoder {
	return &SafeImageDecoder{
		validator: validator,
	}
}

// DecodeImage safely decodes an image with size validation
func (d *SafeImageDecoder) DecodeImage(filePath string) (image.Image, error) {
	// First, perform all validation checks
	sanitizedPath, err := d.validator.ValidateImage(filePath)
	if err != nil {
		return nil, fmt.Errorf("image validation failed: %w", err)
	}

	// Open the validated file
	file, err := os.Open(sanitizedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Get image configuration without decoding the entire image
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	// Validate image dimensions
	if config.Width > d.validator.MaxImageWidth {
		return nil, fmt.Errorf("image width (%d) exceeds maximum allowed width (%d)",
			config.Width, d.validator.MaxImageWidth)
	}

	if config.Height > d.validator.MaxImageHeight {
		return nil, fmt.Errorf("image height (%d) exceeds maximum allowed height (%d)",
			config.Height, d.validator.MaxImageHeight)
	}

	// Check for potential memory usage (rough estimate)
	// 4 bytes per pixel (RGBA) + overhead
	estimatedMemory := int64(config.Width) * int64(config.Height) * 4
	if estimatedMemory > d.validator.MaxFileSize*2 { // Use 2x file size as memory limit
		return nil, fmt.Errorf("estimated memory usage (%d bytes) exceeds safe limits", estimatedMemory)
	}

	// Reset file position for actual decoding
	if _, seekErr := file.Seek(0, 0); seekErr != nil {
		return nil, fmt.Errorf("failed to reset file position: %w", seekErr)
	}

	// Decode the image
	img, decodedFormat, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Verify format consistency
	if format != decodedFormat {
		return nil, fmt.Errorf("format mismatch: config returned %s, decode returned %s",
			format, decodedFormat)
	}

	return img, nil
}
