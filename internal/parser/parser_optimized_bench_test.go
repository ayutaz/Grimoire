package parser

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/ayutaz/grimoire/internal/detector"
)

// createLargeBenchmarkSymbols creates a large set of symbols for benchmarking
func createLargeBenchmarkSymbols(numSymbols int) []*detector.Symbol {
	symbols := []*detector.Symbol{
		{Type: detector.OuterCircle, Position: detector.Position{X: 800, Y: 800}},
		{Type: detector.DoubleCircle, Position: detector.Position{X: 400, Y: 100}},
	}

	// Use a fixed seed for reproducible benchmarks
	rng := rand.New(rand.NewSource(42))

	// Add random symbols
	for i := 0; i < numSymbols-2; i++ {
		symbolTypes := []detector.SymbolType{
			detector.Square, detector.Circle, detector.Triangle,
			detector.Pentagon, detector.Hexagon, detector.Star,
			detector.Convergence, detector.Divergence,
			detector.Amplification, detector.Distribution,
		}

		symbolType := symbolTypes[rng.Intn(len(symbolTypes))]
		x := float64(100 + rng.Intn(600))
		y := float64(150 + rng.Intn(600))

		symbol := &detector.Symbol{
			Type:     symbolType,
			Position: detector.Position{X: x, Y: y},
		}

		// Add pattern for squares
		if symbolType == detector.Square {
			patterns := []string{"empty", "dot", "double_dot", "triple_dot"}
			symbol.Pattern = patterns[rng.Intn(len(patterns))]
		}

		symbols = append(symbols, symbol)
	}

	return symbols
}

// createDenseConnections creates connections between nearby symbols
func createDenseConnections(symbols []*detector.Symbol, density float64) []detector.Connection {
	var connections []detector.Connection

	for i := 0; i < len(symbols); i++ {
		for j := i + 1; j < len(symbols); j++ {
			if rand.Float64() < density {
				// Check distance
				dx := symbols[i].Position.X - symbols[j].Position.X
				dy := symbols[i].Position.Y - symbols[j].Position.Y
				dist := dx*dx + dy*dy

				if dist < 150*150 { // Within connection range
					connections = append(connections, detector.Connection{
						From: symbols[i],
						To:   symbols[j],
					})
				}
			}
		}
	}

	return connections
}

// Benchmark parser comparison
func BenchmarkParserComparison(b *testing.B) {
	testCases := []struct {
		numSymbols int
		density    float64
	}{
		{50, 0.1},
		{100, 0.1},
		{200, 0.05},
		{500, 0.02},
	}
	
	// Limit test cases in CI to avoid timeouts
	if os.Getenv("CI") != "" {
		testCases = testCases[:2] // Only test 50 and 100 symbols in CI
	}

	for _, tc := range testCases {
		symbols := createLargeBenchmarkSymbols(tc.numSymbols)
		connections := createDenseConnections(symbols, tc.density)

		b.Run(fmt.Sprintf("Standard_%dsymbols", tc.numSymbols), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser := NewParser()
				_, _ = parser.Parse(symbols, connections)
			}
		})

		b.Run(fmt.Sprintf("Optimized_%dsymbols", tc.numSymbols), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser := NewOptimizedParser()
				_, _ = parser.Parse(symbols, connections)
			}
		})

		// Add V2 parser benchmark
		b.Run(fmt.Sprintf("OptimizedV2_%dsymbols", tc.numSymbols), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser := NewOptimizedParserV2()
				_, _ = parser.Parse(symbols, connections)
				parser.Cleanup()
			}
		})
	}
}

// Benchmark symbol graph building
func BenchmarkSymbolGraphBuilding(b *testing.B) {
	testCases := []int{50, 100, 200, 500}

	for _, numSymbols := range testCases {
		symbols := createLargeBenchmarkSymbols(numSymbols)

		b.Run(fmt.Sprintf("Standard_%dsymbols", numSymbols), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser := NewParser()
				parser.symbols = symbols
				parser.buildSymbolGraph()
			}
		})

		b.Run(fmt.Sprintf("Optimized_%dsymbols", numSymbols), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser := NewOptimizedParser()
				parser.symbols = symbols
				parser.buildSpatialIndex()
				parser.buildOptimizedSymbolGraph()
			}
		})
	}
}

// Benchmark connection inference
func BenchmarkConnectionInference(b *testing.B) {
	testCases := []int{50, 100, 200}

	for _, numSymbols := range testCases {
		symbols := createLargeBenchmarkSymbols(numSymbols)

		b.Run(fmt.Sprintf("Standard_%dsymbols", numSymbols), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser := NewParser()
				parser.symbols = symbols
				parser.buildSymbolGraph()
				parser.inferConnections()
			}
		})

		b.Run(fmt.Sprintf("Optimized_%dsymbols", numSymbols), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser := NewOptimizedParser()
				parser.symbols = symbols
				parser.buildSpatialIndex()
				parser.buildSymbolGraph()
				parser.inferConnectionsOptimized()
			}
		})
	}
}

// Benchmark spatial index performance
func BenchmarkSpatialIndex(b *testing.B) {
	numSymbols := 500
	symbols := createLargeBenchmarkSymbols(numSymbols)

	parser := NewOptimizedParser()
	parser.symbols = symbols
	parser.buildSpatialIndex()

	b.Run("PointLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pos := detector.Position{
				X: float64(100 + (i*13)%600),
				Y: float64(100 + (i*17)%600),
			}
			_ = parser.spatialIndex.getNearbyNodes(pos, 150)
		}
	})

	b.Run("RangeLookup_Small", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pos := detector.Position{
				X: float64(100 + (i*13)%600),
				Y: float64(100 + (i*17)%600),
			}
			_ = parser.spatialIndex.getNearbyNodes(pos, 50)
		}
	})

	b.Run("RangeLookup_Large", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pos := detector.Position{
				X: float64(100 + (i*13)%600),
				Y: float64(100 + (i*17)%600),
			}
			_ = parser.spatialIndex.getNearbyNodes(pos, 300)
		}
	})
}

// Benchmark expression parsing with caching
func BenchmarkExpressionParsing(b *testing.B) {
	// Create a complex expression tree
	symbols := []*detector.Symbol{
		{Type: detector.Square, Pattern: "dot", Position: detector.Position{X: 100, Y: 100}},
		{Type: detector.Square, Pattern: "double_dot", Position: detector.Position{X: 200, Y: 100}},
		{Type: detector.Convergence, Position: detector.Position{X: 150, Y: 150}},
		{Type: detector.Square, Pattern: "triple_dot", Position: detector.Position{X: 300, Y: 100}},
		{Type: detector.Amplification, Position: detector.Position{X: 250, Y: 200}},
		{Type: detector.Star, Position: detector.Position{X: 250, Y: 300}},
	}

	parser := NewParser()
	parser.symbols = symbols
	parser.buildSymbolGraph()
	parser.inferConnections()

	cache := NewExpressionCache()

	b.Run("WithoutCache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for idx := range symbols {
				node := parser.symbolGraph[idx]
				_ = parser.parseExpression(node)
			}
		}
	})

	b.Run("WithCache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for idx := range symbols {
				node := parser.symbolGraph[idx]

				if expr, ok := cache.get(node); ok {
					_ = expr
				} else {
					expr := parser.parseExpression(node)
					cache.set(node, expr)
				}
			}
		}
	})
}

// Benchmark statement parsing
func BenchmarkStatementParsing(b *testing.B) {
	statementTypes := []struct {
		name   string
		symbol *detector.Symbol
		setup  func(*Parser)
	}{
		{
			"OutputStatement",
			&detector.Symbol{Type: detector.Star},
			nil,
		},
		{
			"IfStatement",
			&detector.Symbol{Type: detector.Triangle},
			func(p *Parser) {
				// Add condition and branches
				p.symbols = append(p.symbols,
					&detector.Symbol{Type: detector.Square, Pattern: "cross"},
					&detector.Symbol{Type: detector.Star},
				)
			},
		},
		{
			"ForLoop",
			&detector.Symbol{Type: detector.Pentagon},
			func(p *Parser) {
				// Add counter and body
				p.symbols = append(p.symbols,
					&detector.Symbol{Type: detector.Square, Pattern: "triple_dot"},
					&detector.Symbol{Type: detector.Star},
				)
			},
		},
		{
			"ParallelBlock",
			&detector.Symbol{Type: detector.Hexagon},
			func(p *Parser) {
				// Add parallel branches
				for i := 0; i < 3; i++ {
					p.symbols = append(p.symbols,
						&detector.Symbol{Type: detector.Star},
					)
				}
			},
		},
	}

	for _, tc := range statementTypes {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser := NewParser()
				parser.symbols = []*detector.Symbol{tc.symbol}

				if tc.setup != nil {
					tc.setup(parser)
				}

				parser.buildSymbolGraph()
				node := parser.symbolGraph[0]
				_ = parser.parseStatement(node)
			}
		})
	}
}
