package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ayutaz/grimoire/internal/detector"
	"github.com/ayutaz/grimoire/internal/parser"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDetector provides mock symbol detection for testing
type MockDetector struct {
	Symbols     []*detector.Symbol
	Connections []detector.Connection
	Error       error
}

// createMockImageFile creates a minimal valid PNG file for testing
func createMockImageFile(t *testing.T, name string) string {
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, name)
	
	// Create a minimal valid PNG (1x1 pixel)
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	
	err := os.WriteFile(imagePath, pngData, 0644)
	require.NoError(t, err)
	
	return imagePath
}

// TestValidateCommandMocked tests validate command with mocked detector
func TestValidateCommandMocked(t *testing.T) {
	// This test would require refactoring the actual code to accept a detector interface
	// For now, we'll skip it but document the approach
	t.Skip("Validate command requires detector interface refactoring")
}

// TestFormatCommandBasic tests format command basic functionality
func TestFormatCommandBasic(t *testing.T) {
	// Test with invalid file path (no detector needed)
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")
	
	err := formatCommand(cmd, []string{"/nonexistent/file.png"})
	assert.Error(t, err)
}

// TestOptimizeCommandBasic tests optimize command basic functionality
func TestOptimizeCommandBasic(t *testing.T) {
	// Test with invalid file path (no detector needed)
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "")
	
	err := optimizeCommand(cmd, []string{"/nonexistent/file.png"})
	assert.Error(t, err)
}

// TestAnalyzeDefinedComplex tests analyzeDefined with complex nested structures
func TestAnalyzeDefinedComplex(t *testing.T) {
	defined := make(map[string]bool)

	// Test ForLoop with nested assignments
	forLoop := &parser.ForLoop{
		Counter: &parser.Identifier{Name: "i"},
		Start:   &parser.Literal{Value: "0"},
		End:     &parser.Literal{Value: "10"},
		Body: []parser.Statement{
			&parser.Assignment{
				Target: &parser.Identifier{Name: "sum"},
				Value:  &parser.Literal{Value: "0"},
			},
			&parser.IfStatement{
				Condition: &parser.BinaryOp{
					Left:     &parser.Identifier{Name: "i"},
					Operator: ">",
					Right:    &parser.Literal{Value: "5"},
				},
				ThenBranch: []parser.Statement{
					&parser.Assignment{
						Target: &parser.Identifier{Name: "result"},
						Value:  &parser.Identifier{Name: "sum"},
					},
				},
			},
		},
	}

	analyzeDefined(forLoop, defined)
	assert.True(t, defined["i"])
	assert.True(t, defined["sum"])
	assert.True(t, defined["result"])
}

// TestAnalyzeUsedComplex tests analyzeUsed with complex nested structures
func TestAnalyzeUsedComplex(t *testing.T) {
	used := make(map[string]bool)

	// Test nested statements
	ifStmt := &parser.IfStatement{
		Condition: &parser.BinaryOp{
			Left:     &parser.Identifier{Name: "x"},
			Operator: "==",
			Right:    &parser.Identifier{Name: "y"},
		},
		ThenBranch: []parser.Statement{
			&parser.ForLoop{
				Counter: &parser.Identifier{Name: "i"},
				Start:   &parser.Identifier{Name: "start"},
				End:     &parser.Identifier{Name: "end"},
				Body: []parser.Statement{
					&parser.OutputStatement{
						Value: &parser.BinaryOp{
							Left:     &parser.Identifier{Name: "array"},
							Operator: "+",
							Right:    &parser.Identifier{Name: "i"},
						},
					},
				},
			},
		},
		ElseBranch: []parser.Statement{
			&parser.WhileLoop{
				Condition: &parser.Identifier{Name: "flag"},
				Body: []parser.Statement{
					&parser.Assignment{
						Target: &parser.Identifier{Name: "result"},
						Value:  &parser.Identifier{Name: "default"},
					},
				},
			},
		},
	}

	analyzeUsed(ifStmt, used)
	assert.True(t, used["x"])
	assert.True(t, used["y"])
	assert.True(t, used["start"])
	assert.True(t, used["end"])
	assert.True(t, used["array"])
	assert.True(t, used["flag"])
	assert.True(t, used["default"])
}

// TestOptimizationAnalysis tests the optimization analysis functions
func TestOptimizationAnalysis(t *testing.T) {
	// Create a simple AST
	ast := &parser.Program{
		HasOuterCircle: true,
		MainEntry: &parser.FunctionDef{
			Name:   "main",
			IsMain: true,
			Body: []parser.Statement{
				&parser.Assignment{
					Target: &parser.Identifier{Name: "unused"},
					Value:  &parser.Literal{Value: "42"},
				},
				&parser.Assignment{
					Target: &parser.Identifier{Name: "used"},
					Value:  &parser.Literal{Value: "10"},
				},
				&parser.OutputStatement{
					Value: &parser.Identifier{Name: "used"},
				},
			},
		},
	}

	// Analyze defined and used variables
	defined := make(map[string]bool)
	used := make(map[string]bool)

	for _, stmt := range ast.MainEntry.Body {
		analyzeDefined(stmt, defined)
		analyzeUsed(stmt, used)
	}

	assert.True(t, defined["unused"])
	assert.True(t, defined["used"])
	assert.False(t, used["unused"])
	assert.True(t, used["used"])
}

// TestStatementsEqualTypes tests statementsEqual with different statement types
func TestStatementsEqualTypes(t *testing.T) {
	assign1 := &parser.Assignment{
		Target: &parser.Identifier{Name: "x"},
		Value:  &parser.Literal{Value: "1"},
	}
	assign2 := &parser.Assignment{
		Target: &parser.Identifier{Name: "y"},
		Value:  &parser.Literal{Value: "2"},
	}
	output := &parser.OutputStatement{
		Value: &parser.Identifier{Name: "x"},
	}
	forLoop := &parser.ForLoop{
		Counter: &parser.Identifier{Name: "i"},
		Start:   &parser.Literal{Value: "0"},
		End:     &parser.Literal{Value: "10"},
		Body:    []parser.Statement{},
	}

	// Same type
	assert.True(t, statementsEqual(assign1, assign2))
	
	// Different types
	assert.False(t, statementsEqual(assign1, output))
	assert.False(t, statementsEqual(output, forLoop))
	assert.False(t, statementsEqual(forLoop, assign1))
}

// TestParallelBlockAnalysis tests analysis of parallel blocks
func TestParallelBlockAnalysis(t *testing.T) {
	defined := make(map[string]bool)
	used := make(map[string]bool)

	parallel := &parser.ParallelBlock{
		Branches: [][]parser.Statement{
			{
				&parser.Assignment{
					Target: &parser.Identifier{Name: "branch1_var"},
					Value:  &parser.Literal{Value: "1"},
				},
				&parser.OutputStatement{
					Value: &parser.Identifier{Name: "shared_var"},
				},
			},
			{
				&parser.Assignment{
					Target: &parser.Identifier{Name: "branch2_var"},
					Value:  &parser.Identifier{Name: "input_var"},
				},
			},
		},
	}

	analyzeDefined(parallel, defined)
	analyzeUsed(parallel, used)

	assert.True(t, defined["branch1_var"])
	assert.True(t, defined["branch2_var"])
	assert.True(t, used["shared_var"])
	assert.True(t, used["input_var"])
}

// TestCaptureOutput tests output capture functionality
func TestCaptureOutput(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Create pipe
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Write something
	fmt.Println("test output")

	// Close and restore
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "test output")
}