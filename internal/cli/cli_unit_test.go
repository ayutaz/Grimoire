package cli

import (
	"errors"
	"testing"

	"github.com/ayutaz/grimoire/internal/parser"
	"github.com/stretchr/testify/assert"
)

// TestAnalyzeDefinedBasic tests analyzeDefined with basic statements
func TestAnalyzeDefinedBasic(t *testing.T) {
	defined := make(map[string]bool)

	// Test simple assignment
	assignment := &parser.Assignment{
		Target: &parser.Identifier{Name: "x"},
		Value:  &parser.Literal{Value: "42"},
	}
	analyzeDefined(assignment, defined)
	assert.True(t, defined["x"])

	// Test nil statement
	analyzeDefined(nil, defined)
	assert.Len(t, defined, 1) // Should not change
}

// TestAnalyzeUsedBasic tests analyzeUsed with basic statements
func TestAnalyzeUsedBasic(t *testing.T) {
	used := make(map[string]bool)

	// Test assignment with variable reference
	assignment := &parser.Assignment{
		Target: &parser.Identifier{Name: "result"},
		Value:  &parser.Identifier{Name: "input"},
	}
	analyzeUsed(assignment, used)
	assert.True(t, used["input"])
	assert.False(t, used["result"]) // Target shouldn't be marked as used

	// Test nil statement
	analyzeUsed(nil, used)
	assert.Len(t, used, 1) // Should not change
}

// TestAnalyzeUsedExprBasic tests analyzeUsedExpr with basic expressions
func TestAnalyzeUsedExprBasic(t *testing.T) {
	used := make(map[string]bool)

	// Test nil expression
	analyzeUsedExpr(nil, used)
	assert.Empty(t, used)

	// Test literal (should not mark anything as used)
	literal := &parser.Literal{Value: "42"}
	analyzeUsedExpr(literal, used)
	assert.Empty(t, used)

	// Test identifier
	id := &parser.Identifier{Name: "var1"}
	analyzeUsedExpr(id, used)
	assert.True(t, used["var1"])
}

// TestStatementsEqualBasic tests statementsEqual function
func TestStatementsEqualBasic(t *testing.T) {
	// Test nil cases
	assert.True(t, statementsEqual(nil, nil))

	assign := &parser.Assignment{
		Target: &parser.Identifier{Name: "x"},
		Value:  &parser.Literal{Value: "1"},
	}
	assert.False(t, statementsEqual(assign, nil))
	assert.False(t, statementsEqual(nil, assign))
}

// TestExecutePythonBasic tests executePython with basic cases
func TestExecutePythonBasic(t *testing.T) {
	// Test empty code
	err := executePython("")
	assert.NoError(t, err, "Empty Python code should execute successfully")

	// Test simple print statement
	err = executePython("print('test')")
	assert.NoError(t, err, "Simple print should execute successfully")
}

// TestFormatErrorTypes tests formatError with different error types
func TestFormatErrorTypes(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		imagePath   string
		expectInMsg string
	}{
		{
			name:        "no such file error",
			err:         errors.New("no such file or directory"),
			imagePath:   "/test/missing.png",
			expectInMsg: "FILE_NOT_FOUND", // Error type is always in English
		},
		{
			name:        "permission denied",
			err:         errors.New("permission denied"),
			imagePath:   "/test/forbidden.png",
			expectInMsg: "FILE_READ_ERROR", // Error type is always in English
		},
		{
			name:        "generic error",
			err:         errors.New("unknown error"),
			imagePath:   "/test/file.png",
			expectInMsg: "EXECUTION_ERROR", // Error type is always in English
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatError(tt.err, tt.imagePath)
			assert.Error(t, result)
			assert.Contains(t, result.Error(), tt.expectInMsg)
		})
	}
}
