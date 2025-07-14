package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecuteCommand tests the main Execute function
func TestExecuteCommand(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test help display (version flag causes os.Exit which we can't easily test)
	os.Args = []string{"grimoire", "--help"}
	err := Execute("1.0.0", "abc123", "2024-01-01")
	// Help should not error
	assert.NoError(t, err)
}

// TestRunCommandWithInvalidFile tests run command with non-existent file
func TestRunCommandWithInvalidFile(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"grimoire", "run", "/nonexistent/file.png"}
	err := Execute("1.0.0", "abc123", "2024-01-01")
	assert.Error(t, err)
	// Check for either Japanese or English error message
	japaneseError := strings.Contains(err.Error(), "ファイルが見つかりません")
	englishError := strings.Contains(err.Error(), "FILE_NOT_FOUND")
	assert.True(t, japaneseError || englishError,
		"Should contain Japanese or English error message. Got: %s", err.Error())
}

// TestCompileCommandWithInvalidFile tests compile command with non-existent file
func TestCompileCommandWithInvalidFile(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"grimoire", "compile", "/nonexistent/file.png"}
	err := Execute("1.0.0", "abc123", "2024-01-01")
	assert.Error(t, err)
}

// TestDebugCommandWithInvalidFile tests debug command with non-existent file
func TestDebugCommandWithInvalidFile(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"grimoire", "debug", "/nonexistent/file.png"}
	err := Execute("1.0.0", "abc123", "2024-01-01")
	assert.Error(t, err)
}

// TestLanguageFlag tests language flag functionality
func TestLanguageFlag(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name string
		lang string
	}{
		{"Japanese short", "ja"},
		{"Japanese long", "japanese"},
		{"English short", "en"},
		{"English long", "english"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = []string{"grimoire", "--lang", tt.lang, "debug", "/nonexistent/file.png"}
			err := Execute("1.0.0", "abc123", "2024-01-01")
			assert.Error(t, err) // File doesn't exist, but language flag should be processed
		})
	}
}

// TestDebugFlag tests debug flag functionality
func TestDebugFlag(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldDebug := os.Getenv("GRIMOIRE_DEBUG")
	defer func() {
		os.Args = oldArgs
		os.Setenv("GRIMOIRE_DEBUG", oldDebug)
	}()

	// Test with flag
	os.Args = []string{"grimoire", "--debug", "debug", "/nonexistent/file.png"}
	err := Execute("1.0.0", "abc123", "2024-01-01")
	assert.Error(t, err)

	// Test with environment variable
	os.Setenv("GRIMOIRE_DEBUG", "1")
	os.Args = []string{"grimoire", "debug", "/nonexistent/file.png"}
	err = Execute("1.0.0", "abc123", "2024-01-01")
	assert.Error(t, err)
}

// TestCompileCommandWithOutput tests compile command with output flag
func TestCompileCommandWithOutput(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.py")

	os.Args = []string{"grimoire", "compile", "/nonexistent/file.png", "-o", outputPath}
	err := Execute("1.0.0", "abc123", "2024-01-01")
	assert.Error(t, err) // File doesn't exist, but output flag should be parsed
}

// TestInvalidCommand tests invalid command
func TestInvalidCommand(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"grimoire", "invalid-command"}
	err := Execute("1.0.0", "abc123", "2024-01-01")
	assert.Error(t, err)
}

// TestNoArguments tests running without arguments
func TestNoArguments(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"grimoire"}
	err := Execute("1.0.0", "abc123", "2024-01-01")
	// Should show help, not error
	assert.NoError(t, err)
}

// TestCommandsWithMissingArgs tests commands without required arguments
func TestCommandsWithMissingArgs(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	commands := []string{"run", "compile", "debug", "validate", "format", "optimize"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			os.Args = []string{"grimoire", cmd}
			err := Execute("1.0.0", "abc123", "2024-01-01")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "accepts 1 arg(s)")
		})
	}
}

// TestFormatCommandWithOutput tests format command with output flag
func TestFormatCommandWithOutput(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "formatted.png")

	os.Args = []string{"grimoire", "format", "/nonexistent/file.png", "-o", outputPath}
	err := Execute("1.0.0", "abc123", "2024-01-01")
	assert.Error(t, err) // File doesn't exist, but output flag should be parsed
}

// TestOptimizeCommandWithOutput tests optimize command with output flag
func TestOptimizeCommandWithOutput(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "optimized.py")

	os.Args = []string{"grimoire", "optimize", "/nonexistent/file.png", "-o", outputPath}
	err := Execute("1.0.0", "abc123", "2024-01-01")
	assert.Error(t, err) // File doesn't exist, but output flag should be parsed

	// Test with stdout output
	os.Args = []string{"grimoire", "optimize", "/nonexistent/file.png", "-o", "-"}
	err = Execute("1.0.0", "abc123", "2024-01-01")
	assert.Error(t, err)
}

// TestExecutePythonErrorCases tests executePython error handling
func TestExecutePythonErrorCases(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "syntax error",
			code: "print('unclosed string",
		},
		{
			name: "runtime error",
			code: "raise ValueError('test error')",
		},
		{
			name: "import error",
			code: "import nonexistent_module_12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executePython(tt.code)
			assert.Error(t, err)
		})
	}
}

// TestProcessImageInvalidPath tests processImage with invalid path
func TestProcessImageInvalidPath(t *testing.T) {
	_, err := processImage("/nonexistent/file.png")
	assert.Error(t, err)
}

// TestFormatErrorWithPermissionDenied tests formatError with permission denied
func TestFormatErrorWithPermissionDenied(t *testing.T) {
	// Create a file with no read permissions
	tmpDir := t.TempDir()
	protectedFile := filepath.Join(tmpDir, "protected.png")

	err := os.WriteFile(protectedFile, []byte("test"), 0000)
	require.NoError(t, err)

	// Try to read it (should fail with permission denied)
	_, readErr := os.ReadFile(protectedFile)
	if readErr != nil && strings.Contains(readErr.Error(), "permission denied") {
		formattedErr := formatError(readErr, protectedFile)
		assert.Error(t, formattedErr)
		// Check for either Japanese or English error message
		japaneseError := strings.Contains(formattedErr.Error(), "ファイル読み込みエラー")
		englishError := strings.Contains(formattedErr.Error(), "FILE_READ_ERROR")
		assert.True(t, japaneseError || englishError,
			"Should contain Japanese or English error message. Got: %s", formattedErr.Error())
	}
}
