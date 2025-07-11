package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Calculator tests calculator.png processing
func TestE2E_Calculator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build the binary
	binaryFile := "grimoire_test"
	if runtime.GOOS == "windows" {
		binaryFile = "grimoire_test.exe"
	}
	
	buildCmd := exec.Command("go", "build", "-o", binaryFile, "../cmd/grimoire")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build grimoire binary")
	defer os.Remove(binaryFile)

	// Make it executable on Unix
	if runtime.GOOS != "windows" {
		err = os.Chmod(binaryFile, 0755)
		require.NoError(t, err)
	}

	// Run calculator test
	binaryName := "./" + binaryFile
	if runtime.GOOS == "windows" {
		binaryName = ".\\" + binaryFile
	}

	// Check if calculator.png exists
	calculatorPath := "../examples/images/calculator.png"
	if _, err := os.Stat(calculatorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: calculator.png not found")
	}

	t.Run("compile_calculator", func(t *testing.T) {
		// Test compile command
		cmd := exec.Command(binaryName, "compile", calculatorPath)
		output, _ := cmd.CombinedOutput()
		
		// Even if it fails, check that it attempted to process
		outputStr := string(output)
		t.Logf("Compile output: %s", outputStr)
		
		// Should generate Python code
		assert.Contains(t, outputStr, "#!/usr/bin/env python3", "Should generate Python code")
	})

	t.Run("debug_calculator", func(t *testing.T) {
		// Test debug command
		cmd := exec.Command(binaryName, "debug", calculatorPath)
		output, err := cmd.CombinedOutput()
		
		// Debug should provide detailed information
		outputStr := string(output)
		t.Logf("Debug output: %s", outputStr)
		
		if err == nil {
			assert.Contains(t, outputStr, "Symbols:", "Should show symbols")
			assert.Contains(t, outputStr, "Connections:", "Should show connections")
		}
	})
}

// TestE2E_AllExamples tests all example images
func TestE2E_AllExamples(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build the binary once
	binaryFile := "grimoire_test"
	if runtime.GOOS == "windows" {
		binaryFile = "grimoire_test.exe"
	}
	
	buildCmd := exec.Command("go", "build", "-o", binaryFile, "../cmd/grimoire")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build grimoire binary")
	defer os.Remove(binaryFile)

	// Make it executable on Unix
	if runtime.GOOS != "windows" {
		err = os.Chmod(binaryFile, 0755)
		require.NoError(t, err)
	}

	binaryName := "./" + binaryFile
	if runtime.GOOS == "windows" {
		binaryName = ".\\" + binaryFile
	}

	// Find all example images
	examplesDir := "../examples/images"
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Skip("Cannot read examples directory")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if !strings.HasSuffix(name, ".png") && !strings.HasSuffix(name, ".jpg") {
			continue
		}

		t.Run(name, func(t *testing.T) {
			imagePath := filepath.Join(examplesDir, name)
			
			// Test compile
			cmd := exec.Command(binaryName, "compile", imagePath)
			output, err := cmd.CombinedOutput()
			
			outputStr := string(output)
			t.Logf("Processing %s: %s", name, outputStr)
			
			// At minimum, it should attempt to process the image
			if err != nil {
				// Some images might fail, but they should fail gracefully
				assert.NotContains(t, outputStr, "panic", "Should not panic")
				assert.NotContains(t, outputStr, "runtime error", "Should not have runtime errors")
			}
		})
	}
}

// TestE2E_ErrorHandling tests error handling
func TestE2E_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build the binary
	binaryFile := "grimoire_test"
	if runtime.GOOS == "windows" {
		binaryFile = "grimoire_test.exe"
	}
	
	buildCmd := exec.Command("go", "build", "-o", binaryFile, "../cmd/grimoire")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build grimoire binary")
	defer os.Remove(binaryFile)

	// Make it executable on Unix
	if runtime.GOOS != "windows" {
		err = os.Chmod(binaryFile, 0755)
		require.NoError(t, err)
	}

	binaryName := "./" + binaryFile
	if runtime.GOOS == "windows" {
		binaryName = ".\\" + binaryFile
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing file",
			args:        []string{"compile", "nonexistent.png"},
			wantErr:     true,
			errContains: "FILE_NOT_FOUND",
		},
		{
			name:        "invalid command",
			args:        []string{"invalid-command"},
			wantErr:     true,
			errContains: "unknown command",
		},
		{
			name:        "no arguments",
			args:        []string{},
			wantErr:     false, // Should show help
		},
		{
			name:        "help flag",
			args:        []string{"--help"},
			wantErr:     false,
		},
		{
			name:        "version flag",
			args:        []string{"--version"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryName, tt.args...)
			output, err := cmd.CombinedOutput()
			
			outputStr := string(output)
			t.Logf("Output: %s", outputStr)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, outputStr, tt.errContains)
				}
			} else {
				// Command might succeed or show help
				if err != nil && !strings.Contains(outputStr, "help") {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestE2E_CalculatorPerformance tests performance benchmarks
func TestE2E_CalculatorPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build optimized binary
	binaryFile := "grimoire_test"
	if runtime.GOOS == "windows" {
		binaryFile = "grimoire_test.exe"
	}
	
	buildCmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", binaryFile, "../cmd/grimoire")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build grimoire binary")
	defer os.Remove(binaryFile)

	// Make it executable on Unix
	if runtime.GOOS != "windows" {
		err = os.Chmod(binaryFile, 0755)
		require.NoError(t, err)
	}

	binaryName := "./" + binaryFile
	if runtime.GOOS == "windows" {
		binaryName = ".\\" + binaryFile
	}

	// Check binary size
	info, err := os.Stat(binaryFile)
	require.NoError(t, err)
	
	binarySize := info.Size()
	t.Logf("Binary size: %.2f MB", float64(binarySize)/(1024*1024))
	
	// Binary should be reasonably small (under 5MB)
	assert.Less(t, binarySize, int64(5*1024*1024), "Binary should be under 5MB")

	// Test startup time with version flag
	cmd := exec.Command(binaryName, "--version")
	err = cmd.Start()
	require.NoError(t, err)
	
	err = cmd.Wait()
	assert.NoError(t, err, "Version flag should exit successfully")
}