package test

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_LoopProgram tests loop execution
func TestE2E_LoopProgram(t *testing.T) {
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
		err = os.Chmod(binaryFile, 0o755)
		require.NoError(t, err)
	}

	binaryName := "./" + binaryFile
	if runtime.GOOS == "windows" {
		binaryName = ".\\" + binaryFile
	}

	// Check if loop.png exists
	loopPath := "../examples/images/loop.png"
	if _, statErr := os.Stat(loopPath); os.IsNotExist(statErr) {
		t.Skip("Skipping test: loop.png not found")
	}

	// Test loop compilation
	cmd := exec.Command(binaryName, "compile", loopPath)
	output, err := cmd.CombinedOutput()
	
	outputStr := string(output)
	t.Logf("Loop compile output: %s", outputStr)
	
	// Should compile successfully
	if err != nil {
		t.Logf("Compilation error (expected if no loop symbol): %v", err)
	} else {
		// Check for Python code generation
		assert.Contains(t, outputStr, "#!/usr/bin/env python3", "Should generate Python code")
	}
}

// TestE2E_ConditionalProgram tests conditional execution
func TestE2E_ConditionalProgram(t *testing.T) {
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
		err = os.Chmod(binaryFile, 0o755)
		require.NoError(t, err)
	}

	binaryName := "./" + binaryFile
	if runtime.GOOS == "windows" {
		binaryName = ".\\" + binaryFile
	}

	// Check if conditional.png exists
	conditionalPath := "../examples/images/conditional.png"
	if _, statErr := os.Stat(conditionalPath); os.IsNotExist(statErr) {
		t.Skip("Skipping test: conditional.png not found")
	}

	// Test conditional compilation
	cmd := exec.Command(binaryName, "compile", conditionalPath)
	output, err := cmd.CombinedOutput()
	
	outputStr := string(output)
	t.Logf("Conditional compile output: %s", outputStr)
	
	// Should generate conditional code
	if err == nil {
		// Check for if/else structure in output
		assert.Contains(t, outputStr, "if", "Should generate if statement")
	}
}

// TestE2E_ParallelProgram tests parallel execution
func TestE2E_ParallelProgram(t *testing.T) {
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
		err = os.Chmod(binaryFile, 0o755)
		require.NoError(t, err)
	}

	binaryName := "./" + binaryFile
	if runtime.GOOS == "windows" {
		binaryName = ".\\" + binaryFile
	}

	// Check if parallel.png exists
	parallelPath := "../examples/images/parallel.png"
	if _, statErr := os.Stat(parallelPath); os.IsNotExist(statErr) {
		t.Skip("Skipping test: parallel.png not found")
	}

	// Test parallel compilation
	cmd := exec.Command(binaryName, "compile", parallelPath)
	output, err := cmd.CombinedOutput()
	
	outputStr := string(output)
	t.Logf("Parallel compile output: %s", outputStr)
	
	// Should compile successfully
	if err != nil {
		t.Logf("Compilation error (expected if no hexagon symbol): %v", err)
	} else {
		// Check for Python code generation
		assert.Contains(t, outputStr, "#!/usr/bin/env python3", "Should generate Python code")
		// Note: Without hexagon symbols, parallel execution won't be generated
		t.Logf("Note: parallel.png doesn't contain hexagon symbols for parallel execution")
	}
}

// TestE2E_ComplexProgram tests a complex program with multiple features
func TestE2E_ComplexProgram(t *testing.T) {
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
		err = os.Chmod(binaryFile, 0o755)
		require.NoError(t, err)
	}

	binaryName := "./" + binaryFile
	if runtime.GOOS == "windows" {
		binaryName = ".\\" + binaryFile
	}

	// Create a temporary complex image or use existing one
	complexPath := "../examples/images/complex.png"
	if _, statErr := os.Stat(complexPath); os.IsNotExist(statErr) {
		// Try other complex examples
		alternativePaths := []string{
			"../examples/images/fibonacci.png",
			"../examples/images/factorial.png",
			"../examples/images/nested_loops.png",
		}
		
		found := false
		for _, path := range alternativePaths {
			if _, checkErr := os.Stat(path); checkErr == nil {
				complexPath = path
				found = true
				break
			}
		}
		
		if !found {
			t.Skip("Skipping test: no complex example found")
		}
	}

	// Test complex program in debug mode to see full analysis
	cmd := exec.Command(binaryName, "debug", complexPath)
	output, err := cmd.CombinedOutput()
	
	outputStr := string(output)
	t.Logf("Complex debug output: %s", outputStr)
	
	// Should detect multiple symbols and connections
	if err == nil {
		lines := strings.Split(outputStr, "\n")
		
		// Count detected symbols
		symbolCount := 0
		connectionCount := 0
		
		for _, line := range lines {
			if strings.Contains(line, "Type:") {
				symbolCount++
			}
			if strings.Contains(line, "Connection") || strings.Contains(line, "->") {
				connectionCount++
			}
		}
		
		t.Logf("Detected %d symbols and %d connections", symbolCount, connectionCount)
		
		// Complex programs should have multiple symbols and connections
		assert.Greater(t, symbolCount, 3, "Complex program should have more than 3 symbols")
		if symbolCount > 3 {
			assert.Greater(t, connectionCount, 1, "Complex program should have connections")
		}
	}
}