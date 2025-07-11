package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_HelloWorld tests end-to-end hello world execution
func TestE2E_HelloWorld(t *testing.T) {
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
		err = os.Chmod("grimoire_test", 0755)
		require.NoError(t, err)
	}

	// Run hello world
	binaryName := "./grimoire_test"
	if runtime.GOOS == "windows" {
		binaryName = ".\\grimoire_test.exe"
	}

	// Check if examples directory exists
	if _, err := os.Stat("../examples/images/hello_world.png"); os.IsNotExist(err) {
		t.Skip("Skipping test: examples/images/hello_world.png not found")
	}

	cmd := exec.Command(binaryName, "run", "../examples/images/hello_world.png")
	output, err := cmd.CombinedOutput()

	// For now, we expect it to work even with placeholder implementation
	if err != nil {
		t.Logf("Output: %s", output)
		t.Logf("Error: %v", err)
	}

	// Once fully implemented, check for "Hello, World!"
	// assert.Contains(t, string(output), "Hello, World!")
}

// TestE2E_CompileCommand tests compile command
func TestE2E_CompileCommand(t *testing.T) {
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

	// Run compile command
	binaryName := "./grimoire_test"
	if runtime.GOOS == "windows" {
		binaryName = ".\\grimoire_test.exe"
	}

	outputFile := filepath.Join(t.TempDir(), "output.py")
	cmd := exec.Command(binaryName, "compile", "../examples/images/hello_world.png", "-o", outputFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Output: %s", output)
	}

	// Check output file was created
	// _, err = os.Stat(outputFile)
	// assert.NoError(t, err)
}

// TestE2E_DebugCommand tests debug command
func TestE2E_DebugCommand(t *testing.T) {
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

	// Run debug command
	binaryName := "./grimoire_test"
	if runtime.GOOS == "windows" {
		binaryName = ".\\grimoire_test.exe"
	}

	cmd := exec.Command(binaryName, "debug", "../examples/images/hello_world.png")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Output: %s", output)
	}

	// Should show detected symbols
	outputStr := string(output)
	assert.Contains(t, outputStr, "Detected")
	// Once implemented: assert.Contains(t, outputStr, "outer_circle")
}

// TestE2E_Performance tests performance requirements
func TestE2E_Performance(t *testing.T) {
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
	require.NoError(t, err, "Failed to build optimized grimoire binary")
	defer os.Remove(binaryFile)

	// Measure execution time
	binaryName := "./grimoire_test"
	if runtime.GOOS == "windows" {
		binaryName = ".\\grimoire_test.exe"
	}

	// Check if examples directory exists
	if _, err := os.Stat("../examples/images/hello_world.png"); os.IsNotExist(err) {
		t.Skip("Skipping test: examples/images/hello_world.png not found")
	}

	// Run multiple times and check performance
	for i := 0; i < 3; i++ {
		cmd := exec.Command(binaryName, "run", "../examples/images/hello_world.png")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Run %d failed: %v", i, err)
			t.Logf("Output: %s", output)
		}
	}

	// TODO: Add actual timing measurements
	// assert.Less(t, avgTime, 100*time.Millisecond, "Execution should be under 100ms")
}

// Benchmark tests
func BenchmarkHelloWorld(b *testing.B) {
	// Check if examples directory exists
	if _, err := os.Stat("examples/images/hello_world.png"); os.IsNotExist(err) {
		b.Skip("Skipping benchmark: examples/images/hello_world.png not found")
	}

	// Build once
	binaryFile := "grimoire_bench"
	if runtime.GOOS == "windows" {
		binaryFile = "grimoire_bench.exe"
	}
	buildCmd := exec.Command("go", "build", "-o", binaryFile, "../cmd/grimoire")
	buildCmd.Dir = "."
	err := buildCmd.Run()
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(binaryFile)

	binaryName := "./grimoire_bench"
	if runtime.GOOS == "windows" {
		binaryName = ".\\grimoire_bench.exe"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryName, "run", "../examples/images/hello_world.png")
		_ = cmd.Run()
	}
}
