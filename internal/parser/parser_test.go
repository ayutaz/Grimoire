package parser

import (
	"image"
	"testing"

	"github.com/ayutaz/grimoire/internal/detector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParserCreation tests that we can create a new parser
func TestParserCreation(t *testing.T) {
	parser := NewParser()
	assert.NotNil(t, parser)
}

// TestParse_NoOuterCircle tests parsing fails without outer circle
func TestParse_NoOuterCircle(t *testing.T) {
	symbols := []detector.Symbol{
		{
			Type:     detector.Circle,
			Position: image.Point{X: 50, Y: 50},
			Size:     20,
		},
	}

	ast, err := Parse(symbols)
	
	assert.Error(t, err)
	assert.Nil(t, ast)
	assert.Contains(t, err.Error(), "outer circle")
}

// TestParse_MinimalProgram tests parsing minimal valid program
func TestParse_MinimalProgram(t *testing.T) {
	symbols := []detector.Symbol{
		{
			Type:       detector.OuterCircle,
			Position:   image.Point{X: 100, Y: 100},
			Size:       180,
			Confidence: 0.9,
		},
	}

	ast, err := Parse(symbols)
	
	require.NoError(t, err)
	require.NotNil(t, ast)
	
	assert.True(t, ast.HasOuterCircle)
	assert.NotNil(t, ast.MainEntry)
	assert.True(t, ast.MainEntry.IsMain)
	assert.Empty(t, ast.Functions)
	
	// Should have default Hello World output
	require.Len(t, ast.MainEntry.Body, 1)
	output, ok := ast.MainEntry.Body[0].(*OutputStatement)
	require.True(t, ok)
	
	literal, ok := output.Value.(*Literal)
	require.True(t, ok)
	assert.Equal(t, "Hello, World!", literal.Value)
}

// TestParse_WithMainEntry tests parsing with explicit main entry
func TestParse_WithMainEntry(t *testing.T) {
	symbols := []detector.Symbol{
		{
			Type:     detector.OuterCircle,
			Position: image.Point{X: 100, Y: 100},
			Size:     180,
		},
		{
			Type:     detector.DoubleCircle,
			Position: image.Point{X: 100, Y: 100},
			Size:     40,
		},
	}

	ast, err := Parse(symbols)
	
	require.NoError(t, err)
	require.NotNil(t, ast)
	
	assert.True(t, ast.HasOuterCircle)
	assert.NotNil(t, ast.MainEntry)
	assert.True(t, ast.MainEntry.IsMain)
}

// TestParseLiteral tests literal parsing from patterns
func TestParseLiteral(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		pattern  string
		expected interface{}
		dataType DataType
	}{
		{"empty", 0, Integer},
		{"dot", 1, Integer},
		{"double_dot", 2, Integer},
		{"triple_dot", 3, Integer},
		{"lines", "Text", String},
		{"triple_line", "Text", String},
		{"cross", true, Boolean},
		{"half_circle", false, Boolean},
	}
	
	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			symbol := &detector.Symbol{
				Type:    detector.Square,
				Pattern: tc.pattern,
			}
			
			literal := parser.parseLiteral(symbol)
			assert.Equal(t, tc.expected, literal.Value)
			assert.Equal(t, tc.dataType, literal.LiteralType)
		})
	}
}

// TestParseOperators tests operator parsing
func TestParseOperators(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		symbolType detector.SymbolType
		expected   OperatorType
	}{
		{detector.Convergence, Add},
		{detector.Divergence, Subtract},
		{detector.Amplification, Multiply},
		{detector.Distribution, Divide},
	}
	
	for _, tc := range tests {
		t.Run(string(tc.symbolType), func(t *testing.T) {
			symbol := &detector.Symbol{
				Type: tc.symbolType,
			}
			
			binOp := parser.parseBinaryOp(symbol)
			assert.Equal(t, tc.expected, binOp.Operator)
		})
	}
}