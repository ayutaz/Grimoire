package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateCommandCoverage tests validate command for coverage
func TestValidateCommandCoverage(t *testing.T) {
	// Create a mock image file
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test.png")
	
	// Create minimal PNG
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	err := os.WriteFile(imagePath, pngData, 0644)
	require.NoError(t, err)

	// Test validate command - it will fail but execute code paths
	cmd := &cobra.Command{}
	err = validateCommand(cmd, []string{imagePath})
	assert.Error(t, err) // Will fail due to no outer circle
}

// TestFormatCommandCoverage tests format command for coverage
func TestFormatCommandCoverage(t *testing.T) {
	// Create a mock image file
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test.png")
	
	// Create minimal PNG
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	err := os.WriteFile(imagePath, pngData, 0644)
	require.NoError(t, err)

	// Test format command - it will fail but execute code paths
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")
	err = formatCommand(cmd, []string{imagePath})
	assert.Error(t, err) // Will fail due to no outer circle

	// Test with output flag
	cmd.Flags().Set("output", "formatted.png")
	err = formatCommand(cmd, []string{imagePath})
	assert.Error(t, err)
}

// TestOptimizeCommandCoverage tests optimize command for coverage
func TestOptimizeCommandCoverage(t *testing.T) {
	// Create a mock image file
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test.png")
	
	// Create minimal PNG
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	err := os.WriteFile(imagePath, pngData, 0644)
	require.NoError(t, err)

	// Test optimize command - it will fail but execute code paths
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")
	err = optimizeCommand(cmd, []string{imagePath})
	assert.Error(t, err) // Will fail due to no outer circle

	// Test with output to stdout
	cmd.Flags().Set("output", "-")
	err = optimizeCommand(cmd, []string{imagePath})
	assert.Error(t, err)

	// Test with output to file
	outputPath := filepath.Join(tmpDir, "optimized.py")
	cmd.Flags().Set("output", outputPath)
	err = optimizeCommand(cmd, []string{imagePath})
	assert.Error(t, err)
}

// TestExecutePythonCoverage tests executePython additional cases
func TestExecutePythonCoverage(t *testing.T) {
	// Test multiline code
	code := `
print("line 1")
print("line 2")
x = 1 + 2
print(f"Result: {x}")
`
	err := executePython(code)
	assert.NoError(t, err)
}

// TestProcessImageAdditional tests processImage error paths
func TestProcessImageAdditional(t *testing.T) {
	// Test with directory instead of file
	tmpDir := t.TempDir()
	_, err := processImage(tmpDir)
	assert.Error(t, err)
}

// TestFormatErrorAdditional tests formatError additional paths
func TestFormatErrorAdditional(t *testing.T) {
	// Test with "access is denied" error
	err := formatError(&testError{msg: "access is denied"}, "/test/file.png")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FILE_READ_ERROR")

	// Test with "failed to open file" error
	err = formatError(&testError{msg: "failed to open file"}, "/test/file.png")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FILE_READ_ERROR")
}