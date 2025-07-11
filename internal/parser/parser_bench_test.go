package parser

import (
	"testing"

	"github.com/ayutaz/grimoire/internal/detector"
)

// createBenchmarkSymbols creates a set of symbols for benchmarking
func createBenchmarkSymbols(complexity string) []*detector.Symbol {
	symbols := []*detector.Symbol{
		{Type: detector.OuterCircle, Position: detector.Position{X: 400, Y: 400}},
	}

	switch complexity {
	case "simple":
		// Simple program with just main and output
		symbols = append(symbols,
			&detector.Symbol{Type: detector.DoubleCircle, Position: detector.Position{X: 200, Y: 100}},
			&detector.Symbol{Type: detector.Star, Position: detector.Position{X: 200, Y: 200}},
		)

	case "medium":
		// Program with variables, operators, and control flow
		symbols = append(symbols,
			&detector.Symbol{Type: detector.DoubleCircle, Position: detector.Position{X: 200, Y: 50}},
			&detector.Symbol{Type: detector.Square, Position: detector.Position{X: 100, Y: 100}, Pattern: "dot"},
			&detector.Symbol{Type: detector.Square, Position: detector.Position{X: 200, Y: 100}, Pattern: "double_dot"},
			&detector.Symbol{Type: detector.Convergence, Position: detector.Position{X: 150, Y: 150}},
			&detector.Symbol{Type: detector.Triangle, Position: detector.Position{X: 150, Y: 200}},
			&detector.Symbol{Type: detector.Star, Position: detector.Position{X: 100, Y: 250}},
			&detector.Symbol{Type: detector.Star, Position: detector.Position{X: 200, Y: 250}},
		)

	case "complex":
		// Complex program with functions, loops, and parallel blocks
		symbols = append(symbols,
			// Main function
			&detector.Symbol{Type: detector.DoubleCircle, Position: detector.Position{X: 200, Y: 50}},
			// Function definition
			&detector.Symbol{Type: detector.Circle, Position: detector.Position{X: 400, Y: 50}},
			// Variables
			&detector.Symbol{Type: detector.Square, Position: detector.Position{X: 100, Y: 100}, Pattern: "empty"},
			&detector.Symbol{Type: detector.Square, Position: detector.Position{X: 200, Y: 100}, Pattern: "dot"},
			&detector.Symbol{Type: detector.Square, Position: detector.Position{X: 300, Y: 100}, Pattern: "double_dot"},
			&detector.Symbol{Type: detector.Square, Position: detector.Position{X: 400, Y: 100}, Pattern: "triple_dot"},
			// Operators
			&detector.Symbol{Type: detector.Convergence, Position: detector.Position{X: 150, Y: 150}},
			&detector.Symbol{Type: detector.Divergence, Position: detector.Position{X: 250, Y: 150}},
			&detector.Symbol{Type: detector.Amplification, Position: detector.Position{X: 350, Y: 150}},
			// Control flow
			&detector.Symbol{Type: detector.Triangle, Position: detector.Position{X: 200, Y: 200}},
			&detector.Symbol{Type: detector.Pentagon, Position: detector.Position{X: 300, Y: 200}},
			&detector.Symbol{Type: detector.Hexagon, Position: detector.Position{X: 400, Y: 200}},
			// Outputs
			&detector.Symbol{Type: detector.Star, Position: detector.Position{X: 150, Y: 300}},
			&detector.Symbol{Type: detector.Star, Position: detector.Position{X: 250, Y: 300}},
			&detector.Symbol{Type: detector.Star, Position: detector.Position{X: 350, Y: 300}},
		)
	}

	return symbols
}

func BenchmarkParse(b *testing.B) {
	complexities := []string{"simple", "medium", "complex"}

	for _, complexity := range complexities {
		symbols := createBenchmarkSymbols(complexity)

		b.Run(complexity, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = Parse(symbols, nil)
			}
		})
	}
}

func BenchmarkBuildSymbolGraph(b *testing.B) {
	symbols := createBenchmarkSymbols("complex")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewParser()
		parser.symbols = symbols
		parser.buildSymbolGraph()
	}
}

func BenchmarkInferConnections(b *testing.B) {
	symbols := createBenchmarkSymbols("complex")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewParser()
		parser.symbols = symbols
		parser.buildSymbolGraph()
		parser.inferConnections()
	}
}

func BenchmarkParseStatement(b *testing.B) {
	testCases := []struct {
		name   string
		symbol *detector.Symbol
	}{
		{"star", &detector.Symbol{Type: detector.Star}},
		{"triangle", &detector.Symbol{Type: detector.Triangle}},
		{"pentagon", &detector.Symbol{Type: detector.Pentagon}},
		{"hexagon", &detector.Symbol{Type: detector.Hexagon}},
	}

	parser := NewParser()

	for _, tc := range testCases {
		node := &symbolNode{symbol: tc.symbol}

		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = parser.parseStatement(node)
			}
		})
	}
}

func BenchmarkParseExpression(b *testing.B) {
	testCases := []struct {
		name   string
		symbol *detector.Symbol
	}{
		{"square_empty", &detector.Symbol{Type: detector.Square, Pattern: "empty"}},
		{"square_dot", &detector.Symbol{Type: detector.Square, Pattern: "dot"}},
		{"convergence", &detector.Symbol{Type: detector.Convergence}},
		{"divergence", &detector.Symbol{Type: detector.Divergence}},
	}

	parser := NewParser()

	for _, tc := range testCases {
		node := &symbolNode{symbol: tc.symbol}

		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = parser.parseExpression(node)
			}
		})
	}
}
