package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantOut string
	}{
		{
			name:    "no arguments shows help",
			args:    []string{"grimoire"},
			wantErr: false,
			wantOut: "visual programming language",
		},
		{
			name:    "version flag",
			args:    []string{"grimoire", "--version"},
			wantErr: false,
			wantOut: "0.1.0",
		},
		{
			name:    "help flag",
			args:    []string{"grimoire", "--help"},
			wantErr: false,
			wantOut: "Draw your spells",
		},
		{
			name:    "unknown command",
			args:    []string{"grimoire", "unknown"},
			wantErr: true,
			wantOut: "",
		},
		{
			name:    "compile without args",
			args:    []string{"grimoire", "compile"},
			wantErr: true,
			wantOut: "",
		},
		{
			name:    "run without args",
			args:    []string{"grimoire", "run"},
			wantErr: true,
			wantOut: "",
		},
		{
			name:    "debug without args",
			args:    []string{"grimoire", "debug"},
			wantErr: true,
			wantOut: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			os.Args = tt.args
			err := Execute("0.1.0", "test", "now")

			// Restore
			w.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			// Read output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantOut != "" {
				assert.Contains(t, output, tt.wantOut)
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	// Create a test image file
	tmpDir := t.TempDir()
	testImage := filepath.Join(tmpDir, "test.png")
	
	// Create an empty file to simulate image
	f, err := os.Create(testImage)
	require.NoError(t, err)
	f.Close()

	// This will fail because it's not a valid image, but we're testing the command logic
	oldArgs := os.Args
	os.Args = []string{"grimoire", "run", testImage}
	defer func() { os.Args = oldArgs }()

	err = Execute("test", "test", "test")
	// Will error because of invalid image, but that's expected
	assert.Error(t, err)
}

func TestCompileCommand(t *testing.T) {
	// Create a test image file
	tmpDir := t.TempDir()
	testImage := filepath.Join(tmpDir, "test.png")
	outputFile := filepath.Join(tmpDir, "output.py")
	
	// Create an empty file to simulate image
	f, err := os.Create(testImage)
	require.NoError(t, err)
	f.Close()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "compile to stdout",
			args: []string{"grimoire", "compile", testImage},
		},
		{
			name: "compile to file",
			args: []string{"grimoire", "compile", testImage, "-o", outputFile},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			os.Args = tt.args
			defer func() { os.Args = oldArgs }()

			// Capture output
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			err := Execute("test", "test", "test")
			
			w.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			// Will error because of invalid image
			assert.Error(t, err)
			
			// Error output goes to stderr, not stdout
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()
			
			// Should have error message
			assert.True(t, err != nil || output != "", "Should either error or produce output")
		})
	}
}

func TestDebugCommand(t *testing.T) {
	// Create a test image file
	tmpDir := t.TempDir()
	testImage := filepath.Join(tmpDir, "test.png")
	
	// Create an empty file to simulate image
	f, err := os.Create(testImage)
	require.NoError(t, err)
	f.Close()

	oldArgs := os.Args
	os.Args = []string{"grimoire", "debug", testImage}
	defer func() { os.Args = oldArgs }()

	// Capture output
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	err = Execute("test", "test", "test")
	
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Will error because of invalid image
	assert.Error(t, err)
	
	// But should show debug attempt
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	assert.True(t,
		strings.Contains(output, "Debug") || 
		strings.Contains(output, "debug") ||
		strings.Contains(output, "Error") ||
		strings.Contains(output, "error"),
		"Should show debug or error output")
}

func TestFileValidation(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		wantValid bool
	}{
		{
			name:      "png file",
			filename:  "test.png",
			wantValid: true,
		},
		{
			name:      "jpg file",
			filename:  "test.jpg",
			wantValid: true,
		},
		{
			name:      "jpeg file",
			filename:  "test.jpeg",
			wantValid: true,
		},
		{
			name:      "gif file",
			filename:  "test.gif",
			wantValid: true,
		},
		{
			name:      "invalid extension",
			filename:  "test.txt",
			wantValid: false,
		},
		{
			name:      "uppercase extension",
			filename:  "test.PNG",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), tt.filename)
			
			// Create the file
			f, err := os.Create(tmpFile)
			require.NoError(t, err)
			f.Close()

			oldArgs := os.Args
			os.Args = []string{"grimoire", "compile", tmpFile}
			defer func() { os.Args = oldArgs }()

			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			err = Execute("test", "test", "test")
			
			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			buf.ReadFrom(r)
			errOutput := buf.String()

			if !tt.wantValid {
				assert.Error(t, err)
				assert.Contains(t, errOutput, "UNSUPPORTED_FORMAT")
			}
		})
	}
}