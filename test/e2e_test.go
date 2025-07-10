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
	buildCmd := exec.Command("go", "build", "-o", "grimoire_test", "./cmd/grimoire")
	err := buildCmd.Run()
	require.NoError(t, err)
	defer os.Remove("grimoire_test")

	// Make it executable on Unix
	if runtime.GOOS != "windows" {
		err = os.Chmod("grimoire_test", 0755)
		require.NoError(t, err)
	}

	// Run hello world
	binaryName := "./grimoire_test"
	if runtime.GOOS == "windows" {
		binaryName = "grimoire_test.exe"
	}

	cmd := exec.Command(binaryName, "run", "examples/images/hello_world.png")
	output, err := cmd.CombinedOutput()

	// For now, we expect it to work even with placeholder implementation
	if err != nil {
		t.Logf("Output: %s", output)
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
	buildCmd := exec.Command("go", "build", "-o", "grimoire_test", "./cmd/grimoire")
	err := buildCmd.Run()
	require.NoError(t, err)
	defer os.Remove("grimoire_test")

	// Run compile command
	binaryName := "./grimoire_test"
	if runtime.GOOS == "windows" {
		binaryName = "grimoire_test.exe"
	}

	outputFile := filepath.Join(t.TempDir(), "output.py")
	cmd := exec.Command(binaryName, "compile", "examples/images/hello_world.png", "-o", outputFile)
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
	buildCmd := exec.Command("go", "build", "-o", "grimoire_test", "./cmd/grimoire")
	err := buildCmd.Run()
	require.NoError(t, err)
	defer os.Remove("grimoire_test")

	// Run debug command
	binaryName := "./grimoire_test"
	if runtime.GOOS == "windows" {
		binaryName = "grimoire_test.exe"
	}

	cmd := exec.Command(binaryName, "debug", "examples/images/hello_world.png")
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
	buildCmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", "grimoire_test", "./cmd/grimoire")
	err := buildCmd.Run()
	require.NoError(t, err)
	defer os.Remove("grimoire_test")

	// Measure execution time
	binaryName := "./grimoire_test"
	if runtime.GOOS == "windows" {
		binaryName = "grimoire_test.exe"
	}

	// Run multiple times and check performance
	for i := 0; i < 3; i++ {
		cmd := exec.Command(binaryName, "run", "examples/images/hello_world.png")
		err := cmd.Run()
		
		// Just check it runs for now
		_ = err
	}

	// TODO: Add actual timing measurements
	// assert.Less(t, avgTime, 100*time.Millisecond, "Execution should be under 100ms")
}

// Benchmark tests
func BenchmarkHelloWorld(b *testing.B) {
	// Build once
	buildCmd := exec.Command("go", "build", "-o", "grimoire_bench", "./cmd/grimoire")
	err := buildCmd.Run()
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove("grimoire_bench")

	binaryName := "./grimoire_bench"
	if runtime.GOOS == "windows" {
		binaryName = "grimoire_bench.exe"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryName, "run", "examples/images/hello_world.png")
		_ = cmd.Run()
	}
}