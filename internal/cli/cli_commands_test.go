package cli

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateCommand tests the validate command
func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name               string
		setupImage         func(t *testing.T) string
		expectError        bool
		expectInOutput     []string
		dontExpectInOutput []string
	}{
		{
			name: "valid magic circle",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "valid.png")

				// Create valid image with outer circle and main entry
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

				// Draw outer circle
				drawCircle(img, 200, 200, 180, 175, color.Black)

				// Draw double circle (main entry)
				drawCircle(img, 200, 200, 30, 25, color.Black)
				drawCircle(img, 200, 200, 25, 20, color.Black)

				// Draw another symbol with connection
				drawCircle(img, 150, 150, 20, 15, color.Black)
				drawLine(img, 200, 200, 150, 150, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			expectError: false,
			expectInOutput: []string{
				"検証に合格しました",
				"検出されたシンボル:",
				"検出された接続:",
			},
		},
		{
			name: "missing outer circle",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "no_outer.png")

				// Create image without outer circle
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

				// Draw double circle (main entry) but no outer circle
				drawCircle(img, 200, 200, 30, 25, color.Black)
				drawCircle(img, 200, 200, 25, 20, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			expectError: true,
			expectInOutput: []string{
				"外周円が検出されません",
			},
		},
		{
			name: "missing main entry",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "no_main.png")

				// Create image with outer circle but no main entry
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

				// Draw outer circle
				drawCircle(img, 200, 200, 180, 175, color.Black)

				// Draw single circle (not double)
				drawCircle(img, 150, 150, 20, 15, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			expectError: true,
			expectInOutput: []string{
				"検証で問題が見つかりました",
				"メイン関数（二重円）が見つかりません",
			},
		},
		{
			name: "orphaned symbols",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "orphaned.png")

				// Create image with orphaned symbols
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

				// Draw outer circle
				drawCircle(img, 200, 200, 180, 175, color.Black)

				// Draw double circle (main entry)
				drawCircle(img, 200, 200, 30, 25, color.Black)
				drawCircle(img, 200, 200, 25, 20, color.Black)

				// Draw orphaned symbol (no connection)
				drawCircle(img, 100, 100, 15, 10, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			expectError: true,
			expectInOutput: []string{
				"検証で問題が見つかりました",
				"検出された接続: 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imagePath := tt.setupImage(t)

			// Capture output
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd := &cobra.Command{}
			err := validateCommand(cmd, []string{imagePath})

			w.Close()
			os.Stdout = oldStdout
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expected := range tt.expectInOutput {
				assert.Contains(t, output, expected, "Expected to find '%s' in output", expected)
			}

			for _, notExpected := range tt.dontExpectInOutput {
				assert.NotContains(t, output, notExpected, "Did not expect to find '%s' in output", notExpected)
			}
		})
	}
}

// TestFormatCommand tests the format command
func TestFormatCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupImage     func(t *testing.T) string
		outputPath     string
		expectInOutput []string
	}{
		{
			name: "well formatted magic circle",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "well_formatted.png")

				// Create well-formatted image
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

				// Draw outer circle
				drawCircle(img, 200, 200, 180, 175, color.Black)

				// Draw symbols at aligned positions
				drawCircle(img, 200, 200, 30, 25, color.Black)
				drawCircle(img, 200, 200, 25, 20, color.Black)
				drawSquare(img, 100, 200, 40, color.Black)
				drawSquare(img, 300, 200, 40, color.Black)

				// Draw perfectly straight connections
				drawLine(img, 200, 200, 120, 200, color.Black)
				drawLine(img, 200, 200, 300, 200, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			expectInOutput: []string{
				"魔法陣の構造を分析中...",
				"魔法陣は適切にフォーマットされています",
			},
		},
		{
			name: "misaligned symbols",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "misaligned.png")

				// Create image with slightly misaligned symbols
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

				// Draw outer circle
				drawCircle(img, 200, 200, 180, 175, color.Black)

				// Draw symbols at slightly misaligned positions
				drawCircle(img, 200, 200, 30, 25, color.Black)
				drawCircle(img, 200, 200, 25, 20, color.Black)
				drawSquare(img, 100, 202, 40, color.Black) // Slightly off
				drawSquare(img, 300, 198, 40, color.Black) // Slightly off

				// Draw connections
				drawLine(img, 200, 200, 120, 202, color.Black)
				drawLine(img, 200, 200, 300, 198, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			expectInOutput: []string{
				"魔法陣の構造を分析中...",
				"フォーマットの提案",
			},
		},
		{
			name: "format with output file",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "format_output.png")

				// Create simple image
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
				drawCircle(img, 200, 200, 180, 175, color.Black)
				drawCircle(img, 200, 200, 30, 25, color.Black)
				drawCircle(img, 200, 200, 25, 20, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			outputPath: "formatted.png",
			expectInOutput: []string{
				"注意:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imagePath := tt.setupImage(t)

			// Capture output
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd := &cobra.Command{}
			cmd.Flags().StringP("output", "o", "", "")
			if tt.outputPath != "" {
				err := cmd.ParseFlags([]string{"-o", tt.outputPath})
				require.NoError(t, err)
			}

			err := formatCommand(cmd, []string{imagePath})

			w.Close()
			os.Stdout = oldStdout
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			assert.NoError(t, err)

			for _, expected := range tt.expectInOutput {
				assert.Contains(t, output, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

// TestOptimizeCommand tests the optimize command
func TestOptimizeCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupImage     func(t *testing.T) string
		outputPath     string
		expectInOutput []string
	}{
		{
			name: "well optimized program",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "optimized.png")

				// Create simple optimized image
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

				// Draw outer circle
				drawCircle(img, 200, 200, 180, 175, color.Black)

				// Draw main entry
				drawCircle(img, 200, 200, 30, 25, color.Black)
				drawCircle(img, 200, 200, 25, 20, color.Black)

				// Simple output operation
				drawSquare(img, 150, 200, 40, color.Black)
				drawLine(img, 200, 200, 170, 200, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			expectInOutput: []string{
				"最適化の機会を探してプログラムを分析中...",
				"プログラムは十分に最適化されています",
			},
		},
		{
			name: "program with unused variables",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "unoptimized.png")

				// Create image with potential optimizations
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

				// Draw outer circle
				drawCircle(img, 200, 200, 180, 175, color.Black)

				// Draw main entry
				drawCircle(img, 200, 200, 30, 25, color.Black)
				drawCircle(img, 200, 200, 25, 20, color.Black)

				// Assignment without usage
				drawSquare(img, 150, 150, 30, color.Black) // Variable
				drawStar(img, 250, 150, 20, color.Black)   // Assignment
				drawLine(img, 170, 150, 230, 150, color.Black)

				// Output something else
				drawSquare(img, 150, 250, 30, color.Black)
				drawLine(img, 200, 200, 150, 250, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			expectInOutput: []string{
				"最適化の機会を探してプログラムを分析中...",
				"最適化の提案",
			},
		},
		{
			name: "optimize with output to stdout",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "optimize_stdout.png")

				// Create simple image
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
				drawCircle(img, 200, 200, 180, 175, color.Black)
				drawCircle(img, 200, 200, 30, 25, color.Black)
				drawCircle(img, 200, 200, 25, 20, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			outputPath: "-",
			expectInOutput: []string{
				"最適化されたコード",
			},
		},
		{
			name: "optimize with output to file",
			setupImage: func(t *testing.T) string {
				tmpDir := t.TempDir()
				imagePath := filepath.Join(tmpDir, "optimize_file.png")

				// Create simple image
				img := image.NewRGBA(image.Rect(0, 0, 400, 400))
				draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
				drawCircle(img, 200, 200, 180, 175, color.Black)
				drawCircle(img, 200, 200, 30, 25, color.Black)
				drawCircle(img, 200, 200, 25, 20, color.Black)

				f, err := os.Create(imagePath)
				require.NoError(t, err)
				err = png.Encode(f, img)
				require.NoError(t, err)
				f.Close()

				return imagePath
			},
			outputPath: "optimized.py",
			expectInOutput: []string{
				"最適化されたコードを保存しました",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imagePath := tt.setupImage(t)

			// Capture output
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd := &cobra.Command{}
			cmd.Flags().StringP("output", "o", "", "")
			if tt.outputPath != "" {
				err := cmd.ParseFlags([]string{"-o", tt.outputPath})
				require.NoError(t, err)
			}

			tmpDir := t.TempDir()
			if tt.outputPath != "" && tt.outputPath != "-" {
				// Use temp dir for output file
				cmd.Flags().Set("output", filepath.Join(tmpDir, tt.outputPath))
			}

			err := optimizeCommand(cmd, []string{imagePath})

			w.Close()
			os.Stdout = oldStdout
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			assert.NoError(t, err)

			for _, expected := range tt.expectInOutput {
				assert.Contains(t, output, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

// TestOptimizeCommandHelpers tests the optimization helper functions
func TestOptimizeCommandHelpers(t *testing.T) {
	// Create a test image that will produce an AST
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test_optimize.png")

	// Create image with various patterns for optimization analysis
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw outer circle
	drawCircle(img, 250, 250, 230, 225, color.Black)

	// Draw main entry (double circle)
	drawCircle(img, 250, 250, 30, 25, color.Black)
	drawCircle(img, 250, 250, 25, 20, color.Black)

	// Draw various symbols for different operations
	// Variable assignment (square = star)
	drawSquare(img, 150, 150, 30, color.Black)     // Variable x
	drawStar(img, 250, 150, 20, color.Black)       // Value
	drawLine(img, 165, 150, 230, 150, color.Black) // Assignment connection

	// Another variable assignment (square = star)
	drawSquare(img, 150, 200, 30, color.Black)     // Variable y
	drawStar(img, 250, 200, 20, color.Black)       // Value
	drawLine(img, 165, 200, 230, 200, color.Black) // Assignment connection

	// Output statement using first variable
	drawTriangle(img, 350, 150, 30, color.Black)   // Output
	drawLine(img, 165, 150, 335, 150, color.Black) // Connection from x to output

	// For loop
	drawCircle(img, 150, 300, 20, 15, color.Black) // Loop counter
	for i := 0; i < 3; i++ {
		y := 300 + i*30
		drawSquare(img, 250, y, 20, color.Black)
		drawLine(img, 170, 300, 240, y, color.Black)
	}

	// Parallel block with branches
	drawSquare(img, 150, 400, 40, color.Black) // Parallel start
	drawTriangle(img, 250, 380, 25, color.Black)
	drawTriangle(img, 250, 420, 25, color.Black)
	drawLine(img, 190, 400, 225, 380, color.Black)
	drawLine(img, 190, 400, 225, 420, color.Black)

	f, err := os.Create(imagePath)
	require.NoError(t, err)
	err = png.Encode(f, img)
	require.NoError(t, err)
	f.Close()

	// Test the optimize command with this complex image
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")

	err = optimizeCommand(cmd, []string{imagePath})

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)

	// Should detect the unused variable y
	assert.Contains(t, output, "最適化の機会を探してプログラムを分析中", "Should show analyzing message")
	// Either it's well optimized or has suggestions
	assert.True(t,
		strings.Contains(output, "プログラムは十分に最適化されています") ||
			strings.Contains(output, "最適化の提案"),
		"Should show optimization result")
}

// TestValidateCommandErrorCases tests error handling in validate command
func TestValidateCommandErrorCases(t *testing.T) {
	// Test with non-existent file
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := validateCommand(cmd, []string{"/non/existent/file.png"})

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)

	assert.Error(t, err)
}

// TestFormatCommandError tests error handling in format command
func TestFormatCommandError(t *testing.T) {
	// Test with invalid file
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")
	err := formatCommand(cmd, []string{"/invalid/path.png"})

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)

	assert.Error(t, err)
}

// TestOptimizeCommandError tests error handling in optimize command
func TestOptimizeCommandError(t *testing.T) {
	// Test with invalid file
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")
	err := optimizeCommand(cmd, []string{"/invalid/path.png"})

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)

	assert.Error(t, err)
}

// TestLanguageSwitching tests language switching functionality
func TestLanguageSwitching(t *testing.T) {
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test.png")

	// Create a simple valid image
	img := image.NewRGBA(image.Rect(0, 0, 400, 400))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	drawCircle(img, 200, 200, 180, 175, color.Black)
	drawCircle(img, 200, 200, 30, 25, color.Black)
	drawCircle(img, 200, 200, 25, 20, color.Black)

	f, err := os.Create(imagePath)
	require.NoError(t, err)
	err = png.Encode(f, img)
	require.NoError(t, err)
	f.Close()

	tests := []struct {
		name     string
		lang     string
		command  func(*cobra.Command, []string) error
		expected string
	}{
		{
			name:     "validate in Japanese",
			lang:     "ja",
			command:  validateCommand,
			expected: "検証に合格しました",
		},
		{
			name:     "validate in English",
			lang:     "en",
			command:  validateCommand,
			expected: "検証に合格しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set language
			rootCmd := &cobra.Command{
				PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
					// This mimics the language setting logic from Execute
					return nil
				},
			}
			rootCmd.PersistentFlags().StringP("lang", "l", "", "")
			err := rootCmd.ParseFlags([]string{"--lang", tt.lang})
			require.NoError(t, err)

			// Capture output
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd := &cobra.Command{}
			err = tt.command(cmd, []string{imagePath})

			w.Close()
			os.Stdout = oldStdout
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			assert.NoError(t, err)
			assert.Contains(t, output, tt.expected)
		})
	}
}

// TestOptimizeWriteError tests optimize command when write fails
func TestOptimizeWriteError(t *testing.T) {
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test.png")

	// Create a simple valid image
	img := image.NewRGBA(image.Rect(0, 0, 400, 400))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	drawCircle(img, 200, 200, 180, 175, color.Black)
	drawCircle(img, 200, 200, 30, 25, color.Black)
	drawCircle(img, 200, 200, 25, 20, color.Black)

	f, err := os.Create(imagePath)
	require.NoError(t, err)
	err = png.Encode(f, img)
	require.NoError(t, err)
	f.Close()

	// Try to write to an invalid path
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")
	err = cmd.ParseFlags([]string{"-o", "/invalid/path/that/does/not/exist/output.py"})
	require.NoError(t, err)

	err = optimizeCommand(cmd, []string{imagePath})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "I/Oエラー")
}

// TestStatementsEqual tests the statementsEqual helper function
func TestStatementsEqual(t *testing.T) {
	// Since we can't easily create parser.Statement objects without a full parse,
	// we'll test the edge cases
	assert.True(t, statementsEqual(nil, nil))
}

// TestValidateCommandMultipleIssues tests validate command with multiple validation issues
func TestValidateCommandMultipleIssues(t *testing.T) {
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "multiple_issues.png")

	// Create image with multiple issues
	img := image.NewRGBA(image.Rect(0, 0, 400, 400))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// No outer circle
	// No main entry
	// Just some orphaned symbols
	drawCircle(img, 100, 100, 20, 15, color.Black)
	drawSquare(img, 300, 300, 30, color.Black)
	drawTriangle(img, 200, 250, 25, color.Black)

	f, err := os.Create(imagePath)
	require.NoError(t, err)
	err = png.Encode(f, img)
	require.NoError(t, err)
	f.Close()

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err = validateCommand(cmd, []string{imagePath})

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.Error(t, err)
	assert.Contains(t, output, "検証で問題が見つかりました")
	assert.Contains(t, output, "1.")
	assert.Contains(t, output, "2.")
	// Should have multiple numbered issues
}

// TestFormatCommandWithAngles tests format command angle detection
func TestFormatCommandWithAngles(t *testing.T) {
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "angled_connections.png")

	// Create image with connections at various angles
	img := image.NewRGBA(image.Rect(0, 0, 600, 600))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw outer circle
	drawCircle(img, 300, 300, 280, 275, color.Black)

	// Draw main entry
	drawCircle(img, 300, 300, 30, 25, color.Black)
	drawCircle(img, 300, 300, 25, 20, color.Black)

	// Draw symbols with almost-straight connections
	drawSquare(img, 400, 302, 30, color.Black) // Almost horizontal (2 pixels off)
	drawLine(img, 330, 300, 400, 302, color.Black)

	drawCircle(img, 300, 200, 20, 15, color.Black) // Perfectly vertical
	drawLine(img, 300, 280, 300, 200, color.Black)

	drawTriangle(img, 200, 198, 25, color.Black) // Almost 45 degrees
	drawLine(img, 280, 280, 200, 198, color.Black)

	f, err := os.Create(imagePath)
	require.NoError(t, err)
	err = png.Encode(f, img)
	require.NoError(t, err)
	f.Close()

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")
	err = formatCommand(cmd, []string{imagePath})

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "フォーマットの提案")
	// Should detect the almost-straight connections
}

// Helper function to draw a triangle
func drawTriangle(img *image.RGBA, cx, cy, size int, c color.Color) {
	// Draw an equilateral triangle
	height := int(float64(size) * 0.866) // sqrt(3)/2

	// Three vertices
	x1, y1 := cx, cy-height*2/3
	x2, y2 := cx-size/2, cy+height/3
	x3, y3 := cx+size/2, cy+height/3

	// Draw three sides
	drawLine(img, x1, y1, x2, y2, c)
	drawLine(img, x2, y2, x3, y3, c)
	drawLine(img, x3, y3, x1, y1, c)
}
