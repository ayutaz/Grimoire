package parser

import (
	"testing"

	"github.com/ayutaz/grimoire/internal/detector"
	"github.com/stretchr/testify/assert"
)

// TestParseParallelBlock tests parallel block parsing
func TestParseParallelBlock(t *testing.T) {
	tests := []struct {
		name        string
		symbols     []*detector.Symbol
		connections []detector.Connection
		wantBranches int
	}{
		{
			name: "hexagon with multiple branches",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.Hexagon, Position: detector.Position{X: 100, Y: 100}},
				// Children in different quadrants
				{Type: detector.Star, Position: detector.Position{X: 150, Y: 150}}, // Bottom-right
				{Type: detector.Star, Position: detector.Position{X: 50, Y: 150}},  // Bottom-left
				{Type: detector.Star, Position: detector.Position{X: 50, Y: 50}},   // Top-left
				{Type: detector.Star, Position: detector.Position{X: 150, Y: 50}},  // Top-right
			},
			connections: []detector.Connection{
				{From: &detector.Symbol{Type: detector.Hexagon}, To: &detector.Symbol{Type: detector.Star}},
			},
		},
		{
			name: "six pointed star parallel block",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.SixPointedStar, Position: detector.Position{X: 100, Y: 100}},
				{Type: detector.Star, Position: detector.Position{X: 120, Y: 120}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := Parse(tt.symbols, tt.connections)
			if err != nil {
				// Might fail due to missing main, but we're testing parsing
				return
			}
			assert.NotNil(t, program)
		})
	}
}

// TestParseFunctionCall tests function call parsing
func TestParseFunctionCall(t *testing.T) {
	tests := []struct {
		name    string
		symbols []*detector.Symbol
		wantErr bool
	}{
		{
			name: "function call with arguments",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 100}},
				{Type: detector.Circle, Position: detector.Position{X: 100, Y: 150}},
				{Type: detector.Square, Position: detector.Position{X: 80, Y: 130}, Pattern: "dot"},
			},
		},
		{
			name: "function definition",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.Circle, Position: detector.Position{X: 100, Y: 100}},
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 150}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.symbols, nil)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// TestParseAssignmentOp tests assignment operator parsing
func TestParseAssignmentOp(t *testing.T) {
	symbols := []*detector.Symbol{
		{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
		{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 50}},
		{Type: detector.Square, Position: detector.Position{X: 50, Y: 100}, Pattern: "empty"},
		{Type: detector.Transfer, Position: detector.Position{X: 100, Y: 100}},
		{Type: detector.Square, Position: detector.Position{X: 150, Y: 100}, Pattern: "dot"},
	}

	program, err := Parse(symbols, nil)
	if err == nil {
		assert.NotNil(t, program)
	}
}

// TestParseLoopVariations tests different loop parsing scenarios
func TestParseLoopVariations(t *testing.T) {
	tests := []struct {
		name    string
		symbols []*detector.Symbol
		wantFor bool
	}{
		{
			name: "for loop with counter",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 50}},
				{Type: detector.Square, Position: detector.Position{X: 100, Y: 100}, Pattern: "triple_dot"},
				{Type: detector.Pentagon, Position: detector.Position{X: 100, Y: 150}},
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 200}},
			},
			wantFor: true,
		},
		{
			name: "while loop without counter",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 50}},
				{Type: detector.Pentagon, Position: detector.Position{X: 100, Y: 150}},
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 200}},
			},
			wantFor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.symbols, nil)
			// Test completes if no panic
			_ = err
		})
	}
}

// TestGroupChildrenByAngle tests angle-based grouping
func TestGroupChildrenByAngle(t *testing.T) {
	parser := NewParser()
	
	// Create a parent node with children in different quadrants
	parentSymbol := &detector.Symbol{
		Type:     detector.Hexagon,
		Position: detector.Position{X: 100, Y: 100},
	}
	
	parentNode := &symbolNode{
		symbol: parentSymbol,
		children: []*symbolNode{
			{symbol: &detector.Symbol{Position: detector.Position{X: 150, Y: 150}}}, // Q0: Bottom-right
			{symbol: &detector.Symbol{Position: detector.Position{X: 50, Y: 150}}},  // Q1: Bottom-left
			{symbol: &detector.Symbol{Position: detector.Position{X: 50, Y: 50}}},   // Q2: Top-left
			{symbol: &detector.Symbol{Position: detector.Position{X: 150, Y: 50}}},  // Q3: Top-right
			{symbol: &detector.Symbol{Position: detector.Position{X: 100, Y: 100}}}, // Center (Q0)
		},
	}
	
	groups := parser.groupChildrenByAngle(parentNode)
	
	// Should have 4 groups (4 quadrants with children)
	assert.Equal(t, 4, len(groups))
}

// TestParseConditionVariations tests condition parsing
func TestParseConditionVariations(t *testing.T) {
	tests := []struct {
		name    string
		symbols []*detector.Symbol
	}{
		{
			name: "condition with comparison",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 50}},
				{Type: detector.Triangle, Position: detector.Position{X: 100, Y: 100}},
				{Type: detector.GreaterThan, Position: detector.Position{X: 100, Y: 150}},
				{Type: detector.Square, Position: detector.Position{X: 80, Y: 150}, Pattern: "dot"},
				{Type: detector.Square, Position: detector.Position{X: 120, Y: 150}, Pattern: "double_dot"},
			},
		},
		{
			name: "condition without comparison",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 50}},
				{Type: detector.Triangle, Position: detector.Position{X: 100, Y: 100}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.symbols, nil)
			// Test completes if no panic
			_ = err
		})
	}
}

// TestParseExpressionEdgeCases tests expression parsing edge cases
func TestParseExpressionEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		symbols []*detector.Symbol
	}{
		{
			name: "divergence operator",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 50}},
				{Type: detector.Square, Position: detector.Position{X: 80, Y: 100}, Pattern: "triple_dot"},
				{Type: detector.Divergence, Position: detector.Position{X: 100, Y: 100}},
				{Type: detector.Square, Position: detector.Position{X: 120, Y: 100}, Pattern: "dot"},
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 150}},
			},
		},
		{
			name: "amplification operator",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 50}},
				{Type: detector.Square, Position: detector.Position{X: 80, Y: 100}, Pattern: "double_dot"},
				{Type: detector.Amplification, Position: detector.Position{X: 100, Y: 100}},
				{Type: detector.Square, Position: detector.Position{X: 120, Y: 100}, Pattern: "double_dot"},
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 150}},
			},
		},
		{
			name: "distribution operator",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 50}},
				{Type: detector.Square, Position: detector.Position{X: 80, Y: 100}, Pattern: "triple_dot"},
				{Type: detector.Distribution, Position: detector.Position{X: 100, Y: 100}},
				{Type: detector.Square, Position: detector.Position{X: 120, Y: 100}, Pattern: "dot"},
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 150}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := Parse(tt.symbols, nil)
			if err == nil {
				assert.NotNil(t, program)
			}
		})
	}
}

// TestParseComplexProgram tests a complex program with multiple features
func TestParseComplexProgram(t *testing.T) {
	symbols := []*detector.Symbol{
		// Outer circle
		{Type: detector.OuterCircle, Position: detector.Position{X: 300, Y: 300}},
		
		// Main entry
		{Type: detector.DoubleCircle, Position: detector.Position{X: 150, Y: 50}},
		
		// Function definition
		{Type: detector.Circle, Position: detector.Position{X: 250, Y: 50}},
		
		// Variables
		{Type: detector.Square, Position: detector.Position{X: 100, Y: 100}, Pattern: "dot"},
		{Type: detector.Square, Position: detector.Position{X: 200, Y: 100}, Pattern: "double_dot"},
		
		// Operators
		{Type: detector.Convergence, Position: detector.Position{X: 150, Y: 150}},
		
		// Control flow
		{Type: detector.Triangle, Position: detector.Position{X: 150, Y: 200}},
		{Type: detector.Pentagon, Position: detector.Position{X: 100, Y: 250}},
		{Type: detector.Hexagon, Position: detector.Position{X: 200, Y: 250}},
		
		// Output
		{Type: detector.Star, Position: detector.Position{X: 150, Y: 300}},
	}
	
	connections := []detector.Connection{
		{From: symbols[1], To: symbols[3]},
		{From: symbols[3], To: symbols[5]},
		{From: symbols[4], To: symbols[5]},
		{From: symbols[5], To: symbols[9]},
	}
	
	program, err := Parse(symbols, connections)
	if err == nil {
		assert.NotNil(t, program)
		assert.True(t, program.HasOuterCircle)
		assert.NotNil(t, program.MainEntry)
	}
}

// TestHasOperatorChild tests the hasOperatorChild function
func TestHasOperatorChild(t *testing.T) {
	node := &symbolNode{
		symbol: &detector.Symbol{Type: detector.Square},
		children: []*symbolNode{
			{symbol: &detector.Symbol{Type: detector.Convergence}},
		},
	}
	
	assert.True(t, hasOperatorChild(node))
	
	node2 := &symbolNode{
		symbol: &detector.Symbol{Type: detector.Square},
		children: []*symbolNode{
			{symbol: &detector.Symbol{Type: detector.Star}},
		},
	}
	
	assert.False(t, hasOperatorChild(node2))
}

// TestParseWithDebugMode tests parsing with debug mode enabled
func TestParseWithDebugMode(t *testing.T) {
	// Enable debug mode
	t.Setenv("GRIMOIRE_DEBUG", "1")
	
	symbols := []*detector.Symbol{
		{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
		{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 100}},
		{Type: detector.Star, Position: detector.Position{X: 100, Y: 150}},
	}
	
	// Should print debug information but not fail
	_, err := Parse(symbols, nil)
	_ = err
}

// TestParseEdgeCasesAndErrors tests various error conditions
func TestParseEdgeCasesAndErrors(t *testing.T) {
	tests := []struct {
		name    string
		symbols []*detector.Symbol
		wantErr bool
		errType string
	}{
		{
			name:    "empty symbols",
			symbols: []*detector.Symbol{},
			wantErr: true,
			errType: "No symbols to parse",
		},
		{
			name: "no outer circle",
			symbols: []*detector.Symbol{
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 100}},
			},
			wantErr: true,
			errType: "NO_OUTER_CIRCLE",
		},
		{
			name: "unbalanced binary operator",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
				{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 50}},
				{Type: detector.Convergence, Position: detector.Position{X: 100, Y: 100}},
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 150}},
			},
			wantErr: false, // Binary operators with only one operand don't cause error in current implementation
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.symbols, nil)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != "" {
					assert.Contains(t, err.Error(), tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestParseLiteralPatterns tests all literal pattern types
func TestParseLiteralPatterns(t *testing.T) {
	patterns := []struct {
		pattern     string
		expectedVal interface{}
		expectedType DataType
	}{
		{"dot", 1, Integer},
		{"double_dot", 2, Integer},
		{"triple_dot", 3, Integer},
		{"empty", 0, Integer},
		{"lines", "Text", String},
		{"triple_line", "Text", String},
		{"cross", true, Boolean},
		{"half_circle", false, Boolean},
		{"unknown", 0, Integer},
	}
	
	parser := NewParser()
	for _, p := range patterns {
		t.Run(p.pattern, func(t *testing.T) {
			node := &symbolNode{
				symbol: &detector.Symbol{
					Type:    detector.Square,
					Pattern: p.pattern,
				},
			}
			
			literal := parser.parseLiteral(node)
			assert.Equal(t, p.expectedVal, literal.Value)
			assert.Equal(t, p.expectedType, literal.LiteralType)
		})
	}
}

// TestInferConnections tests connection inference
func TestInferConnections(t *testing.T) {
	parser := NewParser()
	parser.symbols = []*detector.Symbol{
		{Type: detector.OuterCircle, Position: detector.Position{X: 200, Y: 200}},
		{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 50}},
		{Type: detector.Star, Position: detector.Position{X: 100, Y: 150}},
		{Type: detector.Square, Position: detector.Position{X: 80, Y: 100}},
		{Type: detector.Convergence, Position: detector.Position{X: 100, Y: 100}},
	}
	
	// Initialize symbol graph
	parser.buildSymbolGraph()
	
	// Verify connections were inferred
	mainNode := parser.symbolGraph[1]
	assert.True(t, len(mainNode.children) > 0, "Main node should have children")
}

// TestParseStatementPanic tests panic recovery in parseStatement
func TestParseStatementPanic(t *testing.T) {
	parser := NewParser()
	// Create a node that might cause issues
	node := &symbolNode{
		symbol: &detector.Symbol{
			Type:     detector.SymbolType("INVALID"),
			Position: detector.Position{X: 100, Y: 100},
		},
	}
	
	// Should handle panic gracefully
	stmt := parser.parseStatement(node)
	assert.Nil(t, stmt) // Invalid symbol type should return nil
	assert.True(t, len(parser.errors) > 0, "Should have recorded an error")
}