package security

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageValidator_ValidateAndSanitizePath(t *testing.T) {
	validator := NewImageValidator()
	tempDir := t.TempDir()
	validator.WorkingDirectory = tempDir

	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid relative path",
			input:       "test.png",
			expectError: false,
		},
		{
			name:        "Valid nested path",
			input:       "images/test.png",
			expectError: false,
		},
		{
			name:        "Path traversal attempt with ..",
			input:       "../../../etc/passwd",
			expectError: true,
			errorMsg:    "path traversal",
		},
		{
			name:        "Path traversal with mixed separators",
			input:       "images/../../etc/passwd",
			expectError: true,
			errorMsg:    "path traversal",
		},
		{
			name:        "Absolute path (allowed)",
			input:       "/tmp/test.png",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.ValidateAndSanitizePath(tc.input)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				// Result should be absolute path
				assert.True(t, filepath.IsAbs(result))
			}
		})
	}
}

func TestImageValidator_ValidateFileExtension(t *testing.T) {
	validator := NewImageValidator()

	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "Valid PNG extension",
			filePath:    "test.png",
			expectError: false,
		},
		{
			name:        "Valid JPG extension",
			filePath:    "test.jpg",
			expectError: false,
		},
		{
			name:        "Valid JPEG extension",
			filePath:    "test.jpeg",
			expectError: false,
		},
		{
			name:        "Invalid GIF extension",
			filePath:    "test.gif",
			expectError: true,
		},
		{
			name:        "Invalid BMP extension",
			filePath:    "test.bmp",
			expectError: true,
		},
		{
			name:        "No extension",
			filePath:    "test",
			expectError: true,
		},
		{
			name:        "Case insensitive PNG",
			filePath:    "test.PNG",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateFileExtension(tc.filePath)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImageValidator_ValidateFileSize(t *testing.T) {
	validator := NewImageValidator()
	validator.MaxFileSize = 1024 // 1KB for testing

	// Create test files
	tempDir := t.TempDir()

	// Small file (under limit)
	smallFile := filepath.Join(tempDir, "small.png")
	err := os.WriteFile(smallFile, make([]byte, 512), 0644)
	require.NoError(t, err)

	// Large file (over limit)
	largeFile := filepath.Join(tempDir, "large.png")
	err = os.WriteFile(largeFile, make([]byte, 2048), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "File under size limit",
			filePath:    smallFile,
			expectError: false,
		},
		{
			name:        "File over size limit",
			filePath:    largeFile,
			expectError: true,
		},
		{
			name:        "Non-existent file",
			filePath:    filepath.Join(tempDir, "nonexistent.png"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateFileSize(tc.filePath)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImageValidator_ValidateFileHeader(t *testing.T) {
	validator := NewImageValidator()
	tempDir := t.TempDir()

	// PNG magic number
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	// JPEG magic number
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	// Invalid header
	invalidHeader := []byte{0x00, 0x00, 0x00, 0x00}

	tests := []struct {
		name        string
		fileName    string
		content     []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid PNG file",
			fileName:    "valid.png",
			content:     append(pngHeader, make([]byte, 100)...),
			expectError: false,
		},
		{
			name:        "Valid JPEG file",
			fileName:    "valid.jpg",
			content:     append(jpegHeader, make([]byte, 100)...),
			expectError: false,
		},
		{
			name:        "PNG extension with JPEG content",
			fileName:    "fake.png",
			content:     append(jpegHeader, make([]byte, 100)...),
			expectError: true,
			errorMsg:    "not a valid PNG",
		},
		{
			name:        "JPEG extension with PNG content",
			fileName:    "fake.jpg",
			content:     append(pngHeader, make([]byte, 100)...),
			expectError: true,
			errorMsg:    "not a valid JPEG",
		},
		{
			name:        "Invalid header",
			fileName:    "invalid.png",
			content:     append(invalidHeader, make([]byte, 100)...),
			expectError: true,
			errorMsg:    "not a valid",
		},
		{
			name:        "File too small",
			fileName:    "tiny.png",
			content:     []byte{0x89, 0x50},
			expectError: true,
			errorMsg:    "too small",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tc.fileName)
			err := os.WriteFile(filePath, tc.content, 0644)
			require.NoError(t, err)

			err = validator.ValidateFileHeader(filePath)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImageValidator_ValidateImage(t *testing.T) {
	validator := NewImageValidator()
	validator.MaxFileSize = 1024 // 1KB for testing
	tempDir := t.TempDir()
	validator.WorkingDirectory = tempDir

	// Create a valid small PNG file
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	validPNG := filepath.Join(tempDir, "valid.png")
	err := os.WriteFile(validPNG, append(pngHeader, make([]byte, 100)...), 0644)
	require.NoError(t, err)

	// Create a large PNG file
	largePNG := filepath.Join(tempDir, "large.png")
	err = os.WriteFile(largePNG, append(pngHeader, make([]byte, 2000)...), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		inputPath   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid image",
			inputPath:   validPNG,
			expectError: false,
		},
		{
			name:        "Non-existent file",
			inputPath:   filepath.Join(tempDir, "nonexistent.png"),
			expectError: true,
			errorMsg:    "file not found",
		},
		{
			name:        "Invalid extension",
			inputPath:   filepath.Join(tempDir, "test.gif"),
			expectError: true,
			errorMsg:    "file not found", // File doesn't exist, so it fails at existence check first
		},
		{
			name:        "File too large",
			inputPath:   largePNG,
			expectError: true,
			errorMsg:    "exceeds maximum",
		},
		{
			name:        "Path traversal",
			inputPath:   "../../../etc/passwd",
			expectError: true,
			errorMsg:    "path traversal",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.ValidateImage(tc.inputPath)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}
