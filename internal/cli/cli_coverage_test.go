package cli

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	grimoireErrors "github.com/ayutaz/grimoire/internal/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDebugCommandSuccess tests successful debug command execution with symbols and connections
func TestDebugCommandSuccess(t *testing.T) {
	// Create a realistic test image with outer circle and symbols
	tmpDir := t.TempDir()
	testImage := filepath.Join(tmpDir, "debug_test.png")

	// Create a more realistic image
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw outer circle (thick)
	drawCircle(img, 250, 250, 200, 195, color.Black)

	// Draw a star in the center
	drawStar(img, 250, 250, 30, color.Black)

	// Draw a small circle
	drawCircle(img, 150, 150, 25, 20, color.Black)

	// Draw a square
	drawSquare(img, 350, 150, 40, color.Black)

	// Draw connections
	drawLine(img, 250, 250, 150, 150, color.Black)
	drawLine(img, 250, 250, 350, 150, color.Black)

	// Save the image
	f, err := os.Create(testImage)
	require.NoError(t, err)
	err = png.Encode(f, img)
	require.NoError(t, err)
	f.Close()

	// Run debug command
	cmd := &cobra.Command{}
	args := []string{testImage}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = debugCommand(cmd, args)

	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify success
	assert.NoError(t, err, "Debug command should succeed")
	assert.Contains(t, output, "=== ")
	assert.Contains(t, output, "のデバッグ情報 ===")
	assert.Contains(t, output, "個のシンボルと")

	// Should show symbols section
	assert.Contains(t, output, "シンボル:")

	// May or may not have connections, but should handle both cases
	if strings.Contains(output, "接続") && !strings.Contains(output, "0個の接続") {
		assert.Contains(t, output, "接続:")
	}
}

// TestDebugCommandWithConnections tests debug command with multiple connections
func TestDebugCommandWithConnections(t *testing.T) {
	// Mock the detector to return specific symbols and connections
	tmpDir := t.TempDir()
	testImage := filepath.Join(tmpDir, "connections_test.png")

	// Create a complex image with guaranteed connections
	img := createComplexImageWithConnections()

	f, err := os.Create(testImage)
	require.NoError(t, err)
	err = png.Encode(f, img)
	require.NoError(t, err)
	f.Close()

	// Run debug command
	cmd := &cobra.Command{}
	args := []string{testImage}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = debugCommand(cmd, args)

	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should execute without error
	assert.NoError(t, err, "Debug command should succeed")

	// Should contain debug output
	assert.Contains(t, output, "のデバッグ情報")
	assert.Contains(t, output, filepath.Base(testImage))
}

// TestFormatErrorWithGrimoireError tests formatError with already formatted GrimoireError
func TestFormatErrorWithGrimoireError(t *testing.T) {
	// Create a GrimoireError
	grimoireErr := grimoireErrors.NewError(
		grimoireErrors.NoOuterCircle,
		"No outer circle detected",
	).WithLocation("test.png", 10, 20)

	// Format it
	result := formatError(grimoireErr, "ignored.png")

	// Should return an EnhancedError
	enhancedErr, ok := result.(*grimoireErrors.EnhancedError)
	assert.True(t, ok, "Expected EnhancedError, got %T", result)
	assert.NotNil(t, enhancedErr)
	assert.Contains(t, result.Error(), "E2002") // Check for error code
	assert.Contains(t, result.Error(), "外周円が検出されません")
}

// TestFormatErrorWithGenericError tests formatError with generic errors
func TestFormatErrorWithGenericError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		imagePath     string
		expectedType  grimoireErrors.ErrorType
		expectedInMsg string
	}{
		{
			name:          "generic error without 'no such file'",
			err:           errors.New("internal processing error"),
			imagePath:     "/path/to/image.png",
			expectedType:  grimoireErrors.ExecutionError,
			expectedInMsg: "An error occurred",
		},
		{
			name:          "network error",
			err:           errors.New("connection refused"),
			imagePath:     "remote.png",
			expectedType:  grimoireErrors.ExecutionError,
			expectedInMsg: "An error occurred",
		},
		{
			name:          "timeout error",
			err:           errors.New("operation timed out"),
			imagePath:     "slow.png",
			expectedType:  grimoireErrors.ExecutionError,
			expectedInMsg: "operation timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatError(tt.err, tt.imagePath)

			// Should be a GrimoireError
			grimoireErr, ok := result.(*grimoireErrors.GrimoireError)
			assert.True(t, ok, "Should return a GrimoireError")

			// Check error type
			assert.Equal(t, tt.expectedType, grimoireErr.Type)

			// Check that inner error is preserved
			assert.Equal(t, tt.err, grimoireErr.InnerError)

			// Check location
			assert.Equal(t, tt.imagePath, grimoireErr.FileName)

			// Check error message contains expected text
			assert.Contains(t, result.Error(), "実行エラー")
		})
	}
}

// TestExecutePythonErrors tests executePython error handling
func TestExecutePythonErrors(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "syntax error in Python code",
			code:    "print('unclosed string",
			wantErr: true,
		},
		{
			name:    "runtime error in Python code",
			code:    "raise ValueError('test error')",
			wantErr: true,
		},
		{
			name:    "import error",
			code:    "import nonexistent_module_12345",
			wantErr: true,
		},
		{
			name:    "valid Python code",
			code:    "print('Hello from test')",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executePython(tt.code)
			if tt.wantErr {
				assert.Error(t, err, "Should return error for invalid Python code")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestExecutePythonFileCreationError tests executePython when temp file creation fails
func TestExecutePythonFileCreationError(t *testing.T) {
	// Handle different temp directory environment variables for different OS
	var tempEnvVar string
	var originalTemp string
	var invalidPath string

	switch runtime.GOOS {
	case "windows":
		// Windows uses TEMP or TMP
		tempEnvVar = "TEMP"
		originalTemp = os.Getenv(tempEnvVar)
		invalidPath = "Z:\\invalid\\path\\that\\does\\not\\exist"
	default:
		// Unix-like systems use TMPDIR
		tempEnvVar = "TMPDIR"
		originalTemp = os.Getenv(tempEnvVar)
		invalidPath = "/invalid/path/that/does/not/exist"
	}

	// Set temp directory to an invalid path
	os.Setenv(tempEnvVar, invalidPath)
	defer os.Setenv(tempEnvVar, originalTemp)

	// Also set TMP on Windows as a fallback
	if runtime.GOOS == "windows" {
		originalTMP := os.Getenv("TMP")
		os.Setenv("TMP", invalidPath)
		defer os.Setenv("TMP", originalTMP)
	}

	// This should fail to create temp file
	err := executePython("print('test')")
	assert.Error(t, err, "Should return error when temp file creation fails")
}

// TestExecutePythonWriteError simulates write error by testing with very large code
func TestExecutePythonWriteError(t *testing.T) {
	// Create a string that might cause issues
	// This is a best-effort test as it's hard to reliably trigger write errors
	largeCode := "# " + strings.Repeat("x", 1024*1024) + "\nprint('test')"

	err := executePython(largeCode)
	// This may or may not error depending on system limits, but we're testing the code path
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

// TestProcessImageErrors tests processImage error paths
func TestProcessImageErrors(t *testing.T) {
	// This function is already well tested through other test cases,
	// but we can add edge cases

	tests := []struct {
		name      string
		imagePath string
		wantErr   bool
	}{
		{
			name:      "empty path",
			imagePath: "",
			wantErr:   true,
		},
		{
			name:      "path with special characters",
			imagePath: "/tmp/test@#$.png",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processImage(tt.imagePath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions for drawing shapes

func drawCircle(img *image.RGBA, cx, cy, outerRadius, innerRadius int, c color.Color) {
	for y := cy - outerRadius; y <= cy+outerRadius; y++ {
		for x := cx - outerRadius; x <= cx+outerRadius; x++ {
			dx := x - cx
			dy := y - cy
			dist := dx*dx + dy*dy
			if dist >= innerRadius*innerRadius && dist <= outerRadius*outerRadius {
				img.Set(x, y, c)
			}
		}
	}
}

func drawStar(img *image.RGBA, cx, cy, size int, c color.Color) {
	// Draw a simple cross/star pattern
	for i := -size; i <= size; i++ {
		img.Set(cx+i, cy, c)
		img.Set(cx, cy+i, c)
		if i >= -size/2 && i <= size/2 {
			img.Set(cx+i, cy+i, c)
			img.Set(cx+i, cy-i, c)
		}
	}
}

func drawSquare(img *image.RGBA, x, y, size int, c color.Color) {
	for i := 0; i < size; i++ {
		// Top and bottom edges
		img.Set(x+i, y, c)
		img.Set(x+i, y+size-1, c)
		// Left and right edges
		img.Set(x, y+i, c)
		img.Set(x+size-1, y+i, c)
	}
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	// Simple line drawing using Bresenham's algorithm
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	sy := 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x1, y1, c)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func createComplexImageWithConnections() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 600, 600))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw outer circle
	drawCircle(img, 300, 300, 250, 245, color.Black)

	// Draw multiple symbols
	drawStar(img, 300, 300, 25, color.Black)       // Center star
	drawCircle(img, 200, 200, 30, 25, color.Black) // Top-left circle
	drawSquare(img, 380, 200, 50, color.Black)     // Top-right square
	drawCircle(img, 200, 400, 30, 25, color.Black) // Bottom-left circle
	drawSquare(img, 380, 380, 50, color.Black)     // Bottom-right square

	// Draw connections between all symbols
	drawLine(img, 300, 300, 200, 200, color.Black)
	drawLine(img, 300, 300, 405, 225, color.Black)
	drawLine(img, 300, 300, 200, 400, color.Black)
	drawLine(img, 300, 300, 405, 405, color.Black)
	drawLine(img, 200, 200, 405, 225, color.Black)
	drawLine(img, 200, 400, 405, 405, color.Black)

	return img
}

// TestDebugCommandDirectly tests the debugCommand function directly
func TestDebugCommandDirectly(t *testing.T) {
	// Create a mock implementation that returns various scenarios
	testCases := []struct {
		name                  string
		setupImage            func() string
		expectedInOutput      []string
		shouldHaveSymbols     bool
		shouldHaveConnections bool
	}{
		{
			name: "image with symbols and connections",
			setupImage: func() string {
				tmpDir := t.TempDir()
				testImage := filepath.Join(tmpDir, "full_test.png")
				img := createComplexImageWithConnections()
				f, _ := os.Create(testImage)
				if err := png.Encode(f, img); err != nil {
					t.Fatalf("Failed to encode PNG: %v", err)
				}
				f.Close()
				return testImage
			},
			expectedInOutput: []string{
				"のデバッグ情報",
				"個のシンボルと",
				"シンボル:",
			},
			shouldHaveSymbols:     true,
			shouldHaveConnections: true,
		},
		{
			name: "image with only outer circle",
			setupImage: func() string {
				tmpDir := t.TempDir()
				testImage := filepath.Join(tmpDir, "circle_only.png")
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
				drawCircle(img, 200, 200, 180, 175, color.Black)
				f, _ := os.Create(testImage)
				if err := png.Encode(f, img); err != nil {
					t.Fatalf("Failed to encode PNG: %v", err)
				}
				f.Close()
				return testImage
			},
			expectedInOutput: []string{
				"のデバッグ情報",
				"個のシンボルと",
			},
			shouldHaveSymbols:     false,
			shouldHaveConnections: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imagePath := tc.setupImage()

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call debugCommand directly
			cmd := &cobra.Command{}
			err := debugCommand(cmd, []string{imagePath})

			w.Close()
			os.Stdout = oldStdout

			// Read output
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Check for expected output
			for _, expected := range tc.expectedInOutput {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}

			// The command may succeed or fail depending on detection,
			// but it should handle both cases gracefully
			if err != nil {
				// If there's an error, it should be properly formatted
				assert.True(t, grimoireErrors.IsGrimoireError(err), "Error should be a GrimoireError")
			}
		})
	}
}

// TestDebugCommandWithMockData tests debug output formatting with known data
func TestDebugCommandWithMockData(t *testing.T) {
	// This test ensures the debug output formatting works correctly
	// by creating an image that should produce predictable results

	tmpDir := t.TempDir()
	testImage := filepath.Join(tmpDir, "mock_test.png")

	// Create image with precise shapes
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw a perfect outer circle
	center := 256
	for angle := 0.0; angle < 360.0; angle += 0.1 {
		rad := angle * 3.14159 / 180.0
		for r := 230; r <= 235; r++ {
			x := int(float64(center) + float64(r)*math.Cos(rad))
			y := int(float64(center) + float64(r)*math.Sin(rad))
			if x >= 0 && x < 512 && y >= 0 && y < 512 {
				img.Set(x, y, color.Black)
			}
		}
	}

	// Draw a clear star in the center
	for i := -20; i <= 20; i++ {
		for j := -2; j <= 2; j++ {
			img.Set(center+i, center+j, color.Black)
			img.Set(center+j, center+i, color.Black)
		}
	}

	f, err := os.Create(testImage)
	require.NoError(t, err)
	err = png.Encode(f, img)
	require.NoError(t, err)
	f.Close()

	// Run debug command
	cmd := &cobra.Command{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = debugCommand(cmd, []string{testImage})

	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify the output format
	assert.Contains(t, output, fmt.Sprintf("=== %s のデバッグ情報 ===", filepath.Base(testImage)))
	assert.Contains(t, output, "個のシンボルと")
	assert.Contains(t, output, "個のシンボルと")
	assert.Contains(t, output, "個の接続")

	// The actual detection may vary, but the format should be consistent
	t.Logf("Debug output:\n%s", output)
}

// TestCompileCommandWriteError tests compile command when output file write fails
func TestCompileCommandWriteError(t *testing.T) {
	// Create a valid test image
	tmpDir := t.TempDir()
	testImage := filepath.Join(tmpDir, "test.png")

	// Create image with outer circle and double circle (main entry)
	img := image.NewRGBA(image.Rect(0, 0, 400, 400))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	drawCircle(img, 200, 200, 180, 175, color.Black)
	// Draw double circle for main entry
	drawCircle(img, 200, 200, 30, 25, color.Black)
	drawCircle(img, 200, 200, 25, 20, color.Black)

	f, err := os.Create(testImage)
	require.NoError(t, err)
	err = png.Encode(f, img)
	require.NoError(t, err)
	f.Close()

	// Try to write to a directory that doesn't exist
	nonExistentDir := filepath.Join(tmpDir, "nonexistent", "deep", "path")
	outputFile := filepath.Join(nonExistentDir, "output.py")

	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "Output file path")
	err = cmd.ParseFlags([]string{"-o", outputFile})
	require.NoError(t, err)

	err = compileCommand(cmd, []string{testImage})

	// Should get a file write error
	assert.Error(t, err, "Should return error when output file can't be written")
	grimoireErr, ok := err.(*grimoireErrors.GrimoireError)
	if assert.True(t, ok, "Should be a GrimoireError") {
		assert.Equal(t, grimoireErrors.FileWriteError, grimoireErr.Type)
	}
}

// TestRunCommandExecutionError tests run command when Python execution fails
func TestRunCommandExecutionError(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")

	// Set PATH to empty to make python3 not found
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", originalPath)

	// Create a valid test image
	tmpDir := t.TempDir()
	testImage := filepath.Join(tmpDir, "test.png")

	// Create image with outer circle and double circle (main entry)
	img := image.NewRGBA(image.Rect(0, 0, 400, 400))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	drawCircle(img, 200, 200, 180, 175, color.Black)
	// Draw double circle for main entry
	drawCircle(img, 200, 200, 30, 25, color.Black)
	drawCircle(img, 200, 200, 25, 20, color.Black)

	f, err := os.Create(testImage)
	require.NoError(t, err)
	err = png.Encode(f, img)
	require.NoError(t, err)
	f.Close()

	cmd := &cobra.Command{}
	err = runCommand(cmd, []string{testImage})

	// Should get error (could be parser error or execution error)
	if assert.Error(t, err, "Should return error when processing image") {
		// Error could be either GrimoireError or a generic error depending on the environment
		grimoireErr, ok := err.(*grimoireErrors.GrimoireError)
		if ok {
			// Could be either ExecutionError (if Python not found) or SyntaxError (if parsing fails)
			assert.True(t, grimoireErr.Type == grimoireErrors.ExecutionError || grimoireErr.Type == grimoireErrors.SyntaxError,
				"Error type should be ExecutionError or SyntaxError, got: %v", grimoireErr.Type)
		}
		// Check that the error is meaningful (either about Python or parsing)
		errStr := err.Error()
		assert.True(t, strings.Contains(errStr, "Python") || strings.Contains(errStr, "python") ||
			strings.Contains(errStr, "executable file not found") || strings.Contains(errStr, "構文エラー") ||
			strings.Contains(errStr, "Parser"),
			"Error should be related to Python execution or parsing: %s", errStr)
	}
}
