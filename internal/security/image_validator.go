package security

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ImageValidator provides secure image validation
type ImageValidator struct {
	// MaxFileSize is the maximum allowed file size in bytes (default: 50MB)
	MaxFileSize int64
	// MaxImageWidth is the maximum allowed image width in pixels (default: 10000)
	MaxImageWidth int
	// MaxImageHeight is the maximum allowed image height in pixels (default: 10000)
	MaxImageHeight int
	// AllowedExtensions contains the allowed file extensions
	AllowedExtensions []string
	// WorkingDirectory is the base directory for file operations
	WorkingDirectory string
}

// NewImageValidator creates a new ImageValidator with default settings
func NewImageValidator() *ImageValidator {
	return &ImageValidator{
		MaxFileSize:       50 * 1024 * 1024, // 50MB
		MaxImageWidth:     10000,
		MaxImageHeight:    10000,
		AllowedExtensions: []string{".png", ".jpg", ".jpeg"},
		WorkingDirectory:  ".",
	}
}

// ValidateAndSanitizePath validates and sanitizes the input file path
func (v *ImageValidator) ValidateAndSanitizePath(inputPath string) (string, error) {
	// Clean the path to normalize it
	cleanPath := filepath.Clean(inputPath)

	// Convert to absolute path to check final destination
	var absPath string
	var err error

	if filepath.IsAbs(cleanPath) {
		absPath = cleanPath
	} else {
		// For relative paths, resolve relative to working directory
		absPath, err = filepath.Abs(filepath.Join(v.WorkingDirectory, cleanPath))
		if err != nil {
			return "", fmt.Errorf("failed to resolve absolute path: %w", err)
		}
	}

	// Now check if the final resolved path is attempting to access sensitive areas
	// This is a more intelligent check than just looking for ".."
	lowerPath := strings.ToLower(absPath)

	// Check for common sensitive directories (customize based on OS)
	sensitivePaths := []string{
		"/etc/",
		"/sys/",
		"/proc/",
		"\\windows\\system32",
		"\\windows\\system",
	}

	for _, sensitive := range sensitivePaths {
		if strings.Contains(lowerPath, sensitive) {
			return "", fmt.Errorf("path traversal attempt detected: %s", inputPath)
		}
	}

	// Additional checks for path traversal patterns
	if strings.Contains(inputPath, "../..") || strings.Contains(inputPath, "..\\..") {
		return "", fmt.Errorf("path traversal attempt detected: %s", inputPath)
	}

	// Check if the path attempts to go above the working directory
	if strings.HasPrefix(inputPath, "../../../") {
		return "", fmt.Errorf("path traversal attempt detected: %s", inputPath)
	}

	return absPath, nil
}

// ValidateFileExtension checks if the file has an allowed extension
func (v *ImageValidator) ValidateFileExtension(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	for _, allowedExt := range v.AllowedExtensions {
		if ext == allowedExt {
			return nil
		}
	}

	return fmt.Errorf("unsupported file extension: %s (allowed: %v)", ext, v.AllowedExtensions)
}

// ValidateFileSize checks if the file size is within limits
func (v *ImageValidator) ValidateFileSize(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if fileInfo.Size() > v.MaxFileSize {
		return fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes)",
			fileInfo.Size(), v.MaxFileSize)
	}

	return nil
}

// ValidateFileHeader checks the file's magic number to ensure it matches the expected format
func (v *ImageValidator) ValidateFileHeader(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 512 bytes for magic number detection
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	if n < 8 {
		return fmt.Errorf("file too small to be a valid image")
	}

	// Check magic numbers
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".png":
		// PNG magic number: 89 50 4E 47 0D 0A 1A 0A
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		if !bytes.HasPrefix(header, pngHeader) {
			return fmt.Errorf("file extension is .png but content is not a valid PNG")
		}
	case ".jpg", ".jpeg":
		// JPEG magic number: FF D8 FF
		jpegHeader := []byte{0xFF, 0xD8, 0xFF}
		if !bytes.HasPrefix(header, jpegHeader) {
			return fmt.Errorf("file extension is %s but content is not a valid JPEG", ext)
		}
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	return nil
}

// ValidateImage performs all validation checks on the image file
func (v *ImageValidator) ValidateImage(inputPath string) (string, error) {
	// Step 1: Validate and sanitize the path
	sanitizedPath, err := v.ValidateAndSanitizePath(inputPath)
	if err != nil {
		return "", fmt.Errorf("path validation failed: %w", err)
	}

	// Step 2: Check if file exists
	if _, err := os.Stat(sanitizedPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", sanitizedPath)
	}

	// Step 3: Validate file extension
	if err := v.ValidateFileExtension(sanitizedPath); err != nil {
		return "", err
	}

	// Step 4: Validate file size
	if err := v.ValidateFileSize(sanitizedPath); err != nil {
		return "", err
	}

	// Step 5: Validate file header (magic number)
	if err := v.ValidateFileHeader(sanitizedPath); err != nil {
		return "", err
	}

	return sanitizedPath, nil
}
