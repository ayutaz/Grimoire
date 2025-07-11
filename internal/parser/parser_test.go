package parser

import (
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
	symbols := []*detector.Symbol{
		{
			Type:     detector.Circle,
			Position: detector.Position{X: 50, Y: 50},
			Size:     20,
		},
	}

	ast, err := Parse(symbols, []detector.Connection{})

	assert.Error(t, err)
	assert.Nil(t, ast)
	assert.Contains(t, err.Error(), "outer circle")
}

// TestParse_MinimalProgram tests parsing minimal valid program
func TestParse_MinimalProgram(t *testing.T) {
	symbols := []*detector.Symbol{
		{
			Type:       detector.OuterCircle,
			Position:   detector.Position{X: 100, Y: 100},
			Size:       180,
			Confidence: 0.9,
		},
		{
			Type:     detector.Star,
			Position: detector.Position{X: 100, Y: 150},
		},
	}

	ast, err := Parse(symbols, []detector.Connection{})

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
	symbols := []*detector.Symbol{
		{
			Type:     detector.OuterCircle,
			Position: detector.Position{X: 100, Y: 100},
			Size:     180,
		},
		{
			Type:     detector.DoubleCircle,
			Position: detector.Position{X: 100, Y: 100},
			Size:     40,
		},
	}

	ast, err := Parse(symbols, []detector.Connection{})

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

			node := &symbolNode{symbol: symbol}
			literal := parser.parseLiteral(node)
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

			node := &symbolNode{symbol: symbol}
			binOp := parser.parseBinaryOp(node)
			assert.Equal(t, tc.expected, binOp.Operator)
		})
	}
}

// TestParse_ErrorHandling tests parser error handling
func TestParse_ErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		symbols []*detector.Symbol
		wantErr string
	}{
		{
			name:    "Empty symbols",
			symbols: []*detector.Symbol{},
			wantErr: "No symbols to parse",
		},
		{
			name: "No outer circle",
			symbols: []*detector.Symbol{
				{Type: detector.Circle, Position: detector.Position{X: 50, Y: 50}},
			},
			wantErr: "outer circle",
		},
		{
			name: "Invalid program structure",
			symbols: []*detector.Symbol{
				{Type: detector.Circle, Position: detector.Position{X: 100, Y: 100}},
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 150}},
			},
			wantErr: "outer circle",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := Parse(tc.symbols, []detector.Connection{})
			if tc.wantErr != "" {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tc.wantErr)
				}
				assert.Nil(t, ast)
			}
		})
	}
}

// TestParse_WithConnections tests parsing with explicit connections
func TestParse_WithConnections(t *testing.T) {
	square1 := &detector.Symbol{
		Type:     detector.Square,
		Position: detector.Position{X: 50, Y: 50},
		Pattern:  "dot",
	}
	square2 := &detector.Symbol{
		Type:     detector.Square,
		Position: detector.Position{X: 150, Y: 50},
		Pattern:  "double_dot",
	}
	convergence := &detector.Symbol{
		Type:     detector.Convergence,
		Position: detector.Position{X: 100, Y: 100},
	}
	star := &detector.Symbol{
		Type:     detector.Star,
		Position: detector.Position{X: 100, Y: 150},
	}

	symbols := []*detector.Symbol{
		{Type: detector.OuterCircle, Position: detector.Position{X: 100, Y: 100}, Size: 200},
		square1,
		square2,
		convergence,
		star,
	}

	connections := []detector.Connection{
		{From: square1, To: convergence, ConnectionType: "solid"},
		{From: square2, To: convergence, ConnectionType: "solid"},
		{From: convergence, To: star, ConnectionType: "solid"},
	}

	ast, err := Parse(symbols, connections)

	require.NoError(t, err)
	require.NotNil(t, ast)
	assert.True(t, ast.HasOuterCircle)

	// Should have implicit main with output statement
	require.NotNil(t, ast.MainEntry)
	assert.True(t, ast.MainEntry.IsMain)
	require.Len(t, ast.MainEntry.Body, 1)

	outputStmt, ok := ast.MainEntry.Body[0].(*OutputStatement)
	require.True(t, ok)
	assert.NotNil(t, outputStmt.Value)
}

// TestParse_ComplexProgram tests parsing a more complex program
func TestParse_ComplexProgram(t *testing.T) {
	symbols := []*detector.Symbol{
		{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}, Size: 380},
		{Type: detector.DoubleCircle, Position: detector.Position{X: 200, Y: 100}, Size: 40},
		{Type: detector.Square, Position: detector.Position{X: 150, Y: 150}, Pattern: "triple_dot"},
		{Type: detector.Pentagon, Position: detector.Position{X: 200, Y: 200}},
		{Type: detector.Star, Position: detector.Position{X: 200, Y: 250}},
		{Type: detector.Triangle, Position: detector.Position{X: 200, Y: 300}},
		{Type: detector.Square, Position: detector.Position{X: 150, Y: 350}, Pattern: "cross"},
		{Type: detector.Star, Position: detector.Position{X: 150, Y: 380}},
		{Type: detector.Square, Position: detector.Position{X: 250, Y: 350}, Pattern: "half_circle"},
		{Type: detector.Star, Position: detector.Position{X: 250, Y: 380}},
	}

	ast, err := Parse(symbols, []detector.Connection{})

	require.NoError(t, err)
	require.NotNil(t, ast)
	assert.True(t, ast.HasOuterCircle)
	assert.NotNil(t, ast.MainEntry)
	assert.True(t, ast.MainEntry.IsMain)
}

// TestParse_ParallelBlock tests parsing parallel execution blocks
func TestParse_ParallelBlock(t *testing.T) {
	symbols := []*detector.Symbol{
		{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}, Size: 380},
		{Type: detector.Hexagon, Position: detector.Position{X: 200, Y: 200}},
		{Type: detector.Star, Position: detector.Position{X: 150, Y: 150}},
		{Type: detector.Star, Position: detector.Position{X: 250, Y: 150}},
		{Type: detector.Star, Position: detector.Position{X: 150, Y: 250}},
		{Type: detector.Star, Position: detector.Position{X: 250, Y: 250}},
	}

	ast, err := Parse(symbols, []detector.Connection{})

	require.NoError(t, err)
	require.NotNil(t, ast)

	// Should create implicit main with parallel block
	require.NotNil(t, ast.MainEntry)
	require.NotEmpty(t, ast.MainEntry.Body)
}

// TestParse_EdgeCases tests edge cases
func TestParse_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		symbols []*detector.Symbol
		check   func(t *testing.T, ast *Program)
	}{
		{
			name: "Only outer circle - minimal valid program",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 100, Y: 100}, Size: 180},
			},
			check: func(t *testing.T, ast *Program) {
				// With only outer circle and no other symbols, mainEntry should be nil
				assert.Nil(t, ast.MainEntry)
				assert.Empty(t, ast.Functions)
				assert.Empty(t, ast.Globals)
			},
		},
		{
			name: "Multiple functions",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}, Size: 300},
				{Type: detector.Circle, Position: detector.Position{X: 150, Y: 150}, Size: 30},
				{Type: detector.Circle, Position: detector.Position{X: 250, Y: 150}, Size: 30},
				{Type: detector.Circle, Position: detector.Position{X: 200, Y: 250}, Size: 30},
			},
			check: func(t *testing.T, ast *Program) {
				assert.Len(t, ast.Functions, 3)
			},
		},
		{
			name: "Nested control structures",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}, Size: 380},
				{Type: detector.Pentagon, Position: detector.Position{X: 200, Y: 150}},
				{Type: detector.Triangle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.Star, Position: detector.Position{X: 180, Y: 250}},
				{Type: detector.Star, Position: detector.Position{X: 220, Y: 250}},
			},
			check: func(t *testing.T, ast *Program) {
				assert.NotNil(t, ast.MainEntry)
				assert.NotEmpty(t, ast.MainEntry.Body)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := Parse(tc.symbols, []detector.Connection{})
			require.NoError(t, err)
			require.NotNil(t, ast)
			tc.check(t, ast)
		})
	}
}
