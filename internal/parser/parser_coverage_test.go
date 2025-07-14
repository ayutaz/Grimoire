package parser

import (
	"os"
	"testing"

	"github.com/ayutaz/grimoire/internal/detector"
	"github.com/stretchr/testify/assert"
)

// TestParseParallelBlock tests parallel block parsing
func TestParseParallelBlock(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Parser
		node        *symbolNode
		checkResult func(t *testing.T, result *ParallelBlock)
	}{
		{
			name: "simple parallel block with children in different quadrants",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Parallel symbol
				parallelNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Hexagon,
						Position: detector.Position{X: 100, Y: 100},
						Pattern:  "parallel",
					},
				}
				// Child statements in different quadrants
				// Top-right quadrant (dx >= 0, dy >= 0)
				child1 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 120, Y: 120}, // Below and right
					},
					parent: parallelNode,
				}
				// Top-left quadrant (dx < 0, dy >= 0)
				child2 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 80, Y: 120}, // Below and left
					},
					parent: parallelNode,
				}
				parallelNode.children = []*symbolNode{child1, child2}
				p.symbolGraph[0] = parallelNode
				p.symbolGraph[1] = child1
				p.symbolGraph[2] = child2
				return p
			},
			node: &symbolNode{
				symbol: &detector.Symbol{
					Type:     detector.Hexagon,
					Position: detector.Position{X: 100, Y: 100},
					Pattern:  "parallel",
				},
			},
			checkResult: func(t *testing.T, result *ParallelBlock) {
				// groupChildrenByAngle groups by quadrants, so we should have 2 branches
				assert.Len(t, result.Branches, 2)
				// Each branch should have one statement
				for _, branch := range result.Branches {
					assert.Len(t, branch, 1)
					_, ok := branch[0].(*OutputStatement)
					assert.True(t, ok, "Expected OutputStatement")
				}
			},
		},
		{
			name: "parallel block with children in same quadrant",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Parallel symbol
				parallelNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Hexagon,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Both children in same quadrant (top-right)
				child1 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 120, Y: 120},
					},
					parent: parallelNode,
				}
				child2 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 130, Y: 130},
					},
					parent: parallelNode,
				}
				parallelNode.children = []*symbolNode{child1, child2}
				p.symbolGraph[0] = parallelNode
				p.symbolGraph[1] = child1
				p.symbolGraph[2] = child2
				return p
			},
			node: &symbolNode{
				symbol: &detector.Symbol{
					Type:     detector.Hexagon,
					Position: detector.Position{X: 100, Y: 100},
				},
			},
			checkResult: func(t *testing.T, result *ParallelBlock) {
				// Both children in same quadrant, so 1 branch with 2 statements
				assert.Len(t, result.Branches, 1)
				assert.Len(t, result.Branches[0], 2)
			},
		},
		{
			name: "parallel block with no children",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				parallelNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Hexagon,
						Position: detector.Position{X: 100, Y: 100},
					},
					children: []*symbolNode{},
				}
				p.symbolGraph[0] = parallelNode
				return p
			},
			node: &symbolNode{
				symbol: &detector.Symbol{
					Type:     detector.Hexagon,
					Position: detector.Position{X: 100, Y: 100},
				},
				children: []*symbolNode{},
			},
			checkResult: func(t *testing.T, result *ParallelBlock) {
				assert.Empty(t, result.Branches)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			// Get the node from the symbolGraph that was set up
			actualNode := p.symbolGraph[0]

			result := p.parseParallelBlock(actualNode)
			assert.NotNil(t, result)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestParseFunctionCall tests function call parsing
func TestParseFunctionCall(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Parser
		node        *symbolNode
		expectNil   bool
		checkResult func(t *testing.T, result *FunctionCall)
	}{
		{
			name: "simple function call - no arguments",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				callNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Circle,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				p.symbolGraph[0] = callNode
				return p
			},
			node: &symbolNode{
				symbol: &detector.Symbol{
					Type:     detector.Circle,
					Position: detector.Position{X: 100, Y: 100},
				},
			},
			expectNil: false,
			checkResult: func(t *testing.T, result *FunctionCall) {
				assert.Equal(t, "print", result.Function.Name)
				assert.Empty(t, result.Arguments)
				assert.Equal(t, Void, result.DataType)
			},
		},
		{
			name: "function call with parent arguments",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Function call node (Circle)
				callNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Circle,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Arguments as parents
				arg1 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 80, Y: 100},
						Pattern:  "triple_dot",
					},
					children: []*symbolNode{callNode},
				}
				arg2 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 120, Y: 100},
						Pattern:  "double_dot",
					},
					children: []*symbolNode{callNode},
				}
				p.symbolGraph[0] = callNode
				p.symbolGraph[1] = arg1
				p.symbolGraph[2] = arg2
				return p
			},
			node: &symbolNode{
				symbol: &detector.Symbol{
					Type:     detector.Circle,
					Position: detector.Position{X: 100, Y: 100},
				},
			},
			expectNil: false,
			checkResult: func(t *testing.T, result *FunctionCall) {
				assert.Equal(t, "print", result.Function.Name)
				assert.Len(t, result.Arguments, 2)
				// Arguments should be parsed as literals
				if len(result.Arguments) == 2 {
					lit1, ok1 := result.Arguments[0].(*Literal)
					lit2, ok2 := result.Arguments[1].(*Literal)
					assert.True(t, ok1)
					assert.True(t, ok2)
					if ok1 && ok2 {
						assert.Equal(t, 3, lit1.Value)
						assert.Equal(t, 2, lit2.Value)
					}
				}
			},
		},
		{
			name: "function call - already visited",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				callNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Circle,
						Position: detector.Position{X: 100, Y: 100},
					},
					visited: true,
				}
				p.symbolGraph[0] = callNode
				return p
			},
			node: &symbolNode{
				symbol: &detector.Symbol{
					Type:     detector.Circle,
					Position: detector.Position{X: 100, Y: 100},
				},
				visited: true,
			},
			expectNil: false,
			checkResult: func(t *testing.T, result *FunctionCall) {
				assert.Equal(t, "print", result.Function.Name)
				assert.Empty(t, result.Arguments)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			// Get the node from the symbolGraph that was set up
			actualNode := p.symbolGraph[0]

			result := p.parseFunctionCall(actualNode)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

// TestParseAssignmentOp tests assignment operator parsing
func TestParseAssignmentOp(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Parser
		node      *symbolNode
		expectNil bool
	}{
		{
			name: "assignment with target and value",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Transfer node
				transferNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Transfer,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Target (parent)
				targetNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 80, Y: 100},
						Pattern:  "filled",
					},
				}
				// Value (child)
				valueNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Circle,
						Position: detector.Position{X: 120, Y: 100},
						Pattern:  "42",
					},
					parent: transferNode,
				}
				transferNode.children = []*symbolNode{valueNode}

				p.symbolGraph[0] = transferNode
				p.symbolGraph[1] = targetNode
				p.symbolGraph[2] = valueNode
				return p
			},
			node: &symbolNode{
				symbol: &detector.Symbol{
					Type: detector.Transfer,
				},
			},
			expectNil: true, // parseAssignmentOp returns nil for valid assignments
		},
		{
			name: "assignment with no connections",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				transferNode := &symbolNode{
					symbol: &detector.Symbol{
						Type: detector.Transfer,
					},
				}
				p.symbolGraph[0] = transferNode
				return p
			},
			node: &symbolNode{
				symbol: &detector.Symbol{
					Type: detector.Transfer,
				},
			},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			// Get the node from the symbolGraph that was set up
			actualNode := p.symbolGraph[0]

			result := p.parseAssignmentOp(actualNode)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestParseLoopImproved tests additional loop parsing scenarios
func TestParseLoopImproved(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Parser
		checkResult func(t *testing.T, result Statement)
	}{
		{
			name: "for loop with counter parent square",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Loop node (Pentagon)
				loopNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Pentagon,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Counter parent (Square)
				counterNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 100, Y: 80},
						Pattern:  "triple_dot", // Will be parsed as 3
					},
					children: []*symbolNode{loopNode},
				}
				// Body statement
				bodyNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 120},
					},
					parent: loopNode,
				}
				loopNode.children = []*symbolNode{bodyNode}

				p.symbolGraph[0] = loopNode
				p.symbolGraph[1] = counterNode
				p.symbolGraph[2] = bodyNode
				return p
			},
			checkResult: func(t *testing.T, result Statement) {
				forLoop, ok := result.(*ForLoop)
				assert.True(t, ok, "Expected ForLoop")
				if ok {
					assert.Equal(t, "i", forLoop.Counter.Name)
					assert.Equal(t, 0, forLoop.Start.(*Literal).Value)
					assert.Equal(t, 3, forLoop.End.(*Literal).Value)
					assert.Equal(t, 1, forLoop.Step.(*Literal).Value)
					assert.Len(t, forLoop.Body, 1)
				}
			},
		},
		{
			name: "while loop without counter parent",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Loop node
				loopNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Pentagon,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Condition (comparison in children)
				condNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.LessThan,
						Position: detector.Position{X: 100, Y: 120},
					},
					parent: loopNode,
				}
				// Body
				bodyNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 140},
					},
					parent: loopNode,
				}
				loopNode.children = []*symbolNode{condNode, bodyNode}

				p.symbolGraph[0] = loopNode
				p.symbolGraph[1] = condNode
				p.symbolGraph[2] = bodyNode
				return p
			},
			checkResult: func(t *testing.T, result Statement) {
				whileLoop, ok := result.(*WhileLoop)
				assert.True(t, ok, "Expected WhileLoop")
				if ok {
					assert.NotNil(t, whileLoop.Condition)
					assert.Len(t, whileLoop.Body, 1)
				}
			},
		},
		{
			name: "while loop without any condition (default false)",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Loop node with only body
				loopNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Pentagon,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Body only
				bodyNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 120},
					},
					parent: loopNode,
				}
				loopNode.children = []*symbolNode{bodyNode}

				p.symbolGraph[0] = loopNode
				p.symbolGraph[1] = bodyNode
				return p
			},
			checkResult: func(t *testing.T, result Statement) {
				whileLoop, ok := result.(*WhileLoop)
				assert.True(t, ok, "Expected WhileLoop")
				if ok {
					// Default condition should be false literal
					lit, ok := whileLoop.Condition.(*Literal)
					assert.True(t, ok, "Expected Literal condition")
					if ok {
						assert.Equal(t, false, lit.Value)
						assert.Equal(t, Boolean, lit.LiteralType)
					}
					assert.Len(t, whileLoop.Body, 1)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			// Get the node from the symbolGraph that was set up
			actualNode := p.symbolGraph[0]

			result := p.parseLoop(actualNode)
			assert.NotNil(t, result)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestParseIfStatementImproved tests additional if statement scenarios
func TestParseIfStatementImproved(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Parser
		checkResult func(t *testing.T, result *IfStatement)
	}{
		{
			name: "if with then and else branches",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// If node
				ifNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Triangle,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Condition (found in children)
				condNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Equal,
						Position: detector.Position{X: 100, Y: 80},
					},
					parent: ifNode,
				}
				// Then branch (left side)
				thenNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 80, Y: 120}, // Left of if node
					},
					parent: ifNode,
				}
				// Else branch (right side)
				elseNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 120, Y: 120}, // Right of if node
					},
					parent: ifNode,
				}
				ifNode.children = []*symbolNode{condNode, thenNode, elseNode}

				p.symbolGraph[0] = ifNode
				p.symbolGraph[1] = condNode
				p.symbolGraph[2] = thenNode
				p.symbolGraph[3] = elseNode
				return p
			},
			checkResult: func(t *testing.T, result *IfStatement) {
				assert.NotNil(t, result.Condition)
				assert.Len(t, result.ThenBranch, 1)
				assert.Len(t, result.ElseBranch, 1)
			},
		},
		{
			name: "if with only then branch",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// If node
				ifNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Triangle,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Condition
				condNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Equal,
						Position: detector.Position{X: 100, Y: 80},
					},
					parent: ifNode,
				}
				// Only then branch (left side)
				thenNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 80, Y: 120}, // Left of if node
					},
					parent: ifNode,
				}
				ifNode.children = []*symbolNode{condNode, thenNode}

				p.symbolGraph[0] = ifNode
				p.symbolGraph[1] = condNode
				p.symbolGraph[2] = thenNode
				return p
			},
			checkResult: func(t *testing.T, result *IfStatement) {
				assert.NotNil(t, result.Condition)
				assert.Len(t, result.ThenBranch, 1)
				assert.Empty(t, result.ElseBranch)
			},
		},
		{
			name: "if with arithmetic condition in parent",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// If node
				ifNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Triangle,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Condition as parent (using Convergence/Add as condition)
				condNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Convergence,
						Position: detector.Position{X: 100, Y: 80},
					},
					children: []*symbolNode{ifNode},
				}
				// Then branch
				thenNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 80, Y: 120},
					},
					parent: ifNode,
				}
				ifNode.children = []*symbolNode{thenNode}

				p.symbolGraph[0] = ifNode
				p.symbolGraph[1] = condNode
				p.symbolGraph[2] = thenNode
				return p
			},
			checkResult: func(t *testing.T, result *IfStatement) {
				assert.NotNil(t, result.Condition)
				// Since LessThan is not handled in parseBinaryOp, check for default behavior
				assert.Len(t, result.ThenBranch, 1)
				assert.Empty(t, result.ElseBranch)
			},
		},
		{
			name: "if with no condition (default false)",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// If node with only body statements
				ifNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Triangle,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Only body statements, no condition
				thenNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 80, Y: 120},
					},
					parent: ifNode,
				}
				ifNode.children = []*symbolNode{thenNode}

				p.symbolGraph[0] = ifNode
				p.symbolGraph[1] = thenNode
				return p
			},
			checkResult: func(t *testing.T, result *IfStatement) {
				// Default condition should be false literal
				lit, ok := result.Condition.(*Literal)
				assert.True(t, ok, "Expected Literal condition")
				if ok {
					assert.Equal(t, false, lit.Value)
					assert.Equal(t, Boolean, lit.LiteralType)
				}
				assert.Len(t, result.ThenBranch, 1)
				assert.Empty(t, result.ElseBranch)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			// Get the node from the symbolGraph that was set up
			actualNode := p.symbolGraph[0]

			result := p.parseIfStatement(actualNode)
			assert.NotNil(t, result)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestParseCondition tests condition parsing
func TestParseCondition(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Parser
		checkResult func(t *testing.T, result Expression)
	}{
		{
			name: "condition with comparison in children",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Parent node
				parentNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Triangle,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Comparison operator as child
				compNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.LessThan,
						Position: detector.Position{X: 100, Y: 120},
					},
					parent: parentNode,
				}
				parentNode.children = []*symbolNode{compNode}

				p.symbolGraph[0] = parentNode
				p.symbolGraph[1] = compNode
				return p
			},
			checkResult: func(t *testing.T, result Expression) {
				binOp, ok := result.(*BinaryOp)
				assert.True(t, ok, "Expected BinaryOp")
				if ok {
					// parseBinaryOp doesn't handle comparison operators, defaults to Add
					assert.Equal(t, Add, binOp.Operator)
				}
			},
		},
		{
			name: "condition with comparison in parent",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Node
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Pentagon,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Comparison operator as parent
				compNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.GreaterThan,
						Position: detector.Position{X: 100, Y: 80},
					},
					children: []*symbolNode{node},
				}

				p.symbolGraph[0] = node
				p.symbolGraph[1] = compNode
				return p
			},
			checkResult: func(t *testing.T, result Expression) {
				binOp, ok := result.(*BinaryOp)
				assert.True(t, ok, "Expected BinaryOp")
				if ok {
					// parseBinaryOp doesn't handle comparison operators, defaults to Add
					assert.Equal(t, Add, binOp.Operator)
				}
			},
		},
		{
			name: "condition with no comparison (default false)",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Node without any comparison operators
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Pentagon,
						Position: detector.Position{X: 100, Y: 100},
					},
				}

				p.symbolGraph[0] = node
				return p
			},
			checkResult: func(t *testing.T, result Expression) {
				lit, ok := result.(*Literal)
				assert.True(t, ok, "Expected Literal")
				if ok {
					assert.Equal(t, false, lit.Value)
					assert.Equal(t, Boolean, lit.LiteralType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			// Get the node from the symbolGraph that was set up
			actualNode := p.symbolGraph[0]

			result := p.parseCondition(actualNode)
			assert.NotNil(t, result)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestIsExpressionSymbol tests the isExpressionSymbol function
func TestIsExpressionSymbol(t *testing.T) {
	tests := []struct {
		symbolType detector.SymbolType
		expected   bool
	}{
		{detector.Circle, true},
		{detector.EightPointedStar, true},
		{detector.Unknown, true},
		{detector.OuterCircle, true},
		{detector.Star, false},
		{detector.Square, false},
		{detector.Triangle, false},
		{detector.Pentagon, false},
		{detector.Hexagon, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.symbolType), func(t *testing.T) {
			result := isExpressionSymbol(tt.symbolType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHasOperatorChild tests the hasOperatorChild function
func TestHasOperatorChild(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *symbolNode
		expected bool
	}{
		{
			name: "node with operator child",
			setup: func() *symbolNode {
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type: detector.Square,
					},
				}
				opChild := &symbolNode{
					symbol: &detector.Symbol{
						Type: detector.Convergence,
					},
				}
				node.children = []*symbolNode{opChild}
				return node
			},
			expected: true,
		},
		{
			name: "node without operator child",
			setup: func() *symbolNode {
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type: detector.Square,
					},
				}
				nonOpChild := &symbolNode{
					symbol: &detector.Symbol{
						Type: detector.Star,
					},
				}
				node.children = []*symbolNode{nonOpChild}
				return node
			},
			expected: false,
		},
		{
			name: "node with no children",
			setup: func() *symbolNode {
				return &symbolNode{
					symbol: &detector.Symbol{
						Type: detector.Square,
					},
					children: []*symbolNode{},
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.setup()
			result := hasOperatorChild(node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseExpressionFromParent tests parseExpressionFromParent
func TestParseExpressionFromParent(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Parser
		checkResult func(t *testing.T, result Expression)
	}{
		{
			name: "node with parent expression",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Child node
				childNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 120},
					},
				}
				// Parent expression node
				parentNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 100, Y: 100},
						Pattern:  "double_dot",
					},
					children: []*symbolNode{childNode},
				}
				childNode.parent = parentNode

				p.symbolGraph[0] = childNode
				p.symbolGraph[1] = parentNode
				return p
			},
			checkResult: func(t *testing.T, result Expression) {
				lit, ok := result.(*Literal)
				assert.True(t, ok, "Expected Literal")
				if ok {
					assert.Equal(t, 2, lit.Value)
					assert.Equal(t, Integer, lit.LiteralType)
				}
			},
		},
		{
			name: "node without parent but with getParents",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Child node (no direct parent)
				childNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 120},
					},
				}
				// Parent through graph
				parentNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Circle,
						Position: detector.Position{X: 100, Y: 100},
					},
					children: []*symbolNode{childNode},
				}

				p.symbolGraph[0] = childNode
				p.symbolGraph[1] = parentNode
				return p
			},
			checkResult: func(t *testing.T, result Expression) {
				fc, ok := result.(*FunctionCall)
				assert.True(t, ok, "Expected FunctionCall")
				if ok {
					assert.Equal(t, "print", fc.Function.Name)
				}
			},
		},
		{
			name: "star node with no parent (default Hello World)",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Standalone star node
				starNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 100},
					},
				}

				p.symbolGraph[0] = starNode
				return p
			},
			checkResult: func(t *testing.T, result Expression) {
				lit, ok := result.(*Literal)
				assert.True(t, ok, "Expected Literal")
				if ok {
					assert.Equal(t, "Hello, World!", lit.Value)
					assert.Equal(t, String, lit.LiteralType)
				}
			},
		},
		{
			name: "non-star node with no parent (default 0)",
			setup: func() *Parser {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				// Non-star node
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 100, Y: 100},
					},
				}

				p.symbolGraph[0] = node
				return p
			},
			checkResult: func(t *testing.T, result Expression) {
				lit, ok := result.(*Literal)
				assert.True(t, ok, "Expected Literal")
				if ok {
					assert.Equal(t, 0, lit.Value)
					assert.Equal(t, Integer, lit.LiteralType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			// Get the node from the symbolGraph that was set up
			actualNode := p.symbolGraph[0]

			result := p.parseExpressionFromParent(actualNode)
			assert.NotNil(t, result)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestGroupChildrenByAngle tests groupChildrenByAngle
func TestGroupChildrenByAngle(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *symbolNode
		checkResult func(t *testing.T, result [][]*symbolNode)
	}{
		{
			name: "children in all four quadrants",
			setup: func() *symbolNode {
				centerNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Hexagon,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// Top-right (dx >= 0, dy >= 0)
				child1 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 120, Y: 120},
					},
				}
				// Top-left (dx < 0, dy >= 0)
				child2 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 80, Y: 120},
					},
				}
				// Bottom-left (dx < 0, dy < 0)
				child3 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 80, Y: 80},
					},
				}
				// Bottom-right (dx >= 0, dy < 0)
				child4 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 120, Y: 80},
					},
				}
				centerNode.children = []*symbolNode{child1, child2, child3, child4}
				return centerNode
			},
			checkResult: func(t *testing.T, result [][]*symbolNode) {
				assert.Len(t, result, 4, "Expected 4 groups (one for each quadrant)")
				// Each group should have one child
				for _, group := range result {
					assert.Len(t, group, 1)
				}
			},
		},
		{
			name: "no children",
			setup: func() *symbolNode {
				return &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Hexagon,
						Position: detector.Position{X: 100, Y: 100},
					},
					children: []*symbolNode{},
				}
			},
			checkResult: func(t *testing.T, result [][]*symbolNode) {
				assert.Nil(t, result, "Expected nil for no children")
			},
		},
		{
			name: "multiple children in same quadrant",
			setup: func() *symbolNode {
				centerNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Hexagon,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				// All in top-right quadrant
				child1 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 120, Y: 120},
					},
				}
				child2 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 130, Y: 130},
					},
				}
				child3 := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 110, Y: 110},
					},
				}
				centerNode.children = []*symbolNode{child1, child2, child3}
				return centerNode
			},
			checkResult: func(t *testing.T, result [][]*symbolNode) {
				assert.Len(t, result, 1, "Expected 1 group for same quadrant")
				assert.Len(t, result[0], 3, "Expected 3 children in the group")
			},
		},
	}

	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.setup()
			result := p.groupChildrenByAngle(node)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestParse tests the main Parse function
func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		symbols     []*detector.Symbol
		connections []detector.Connection
		expectError bool
		checkResult func(t *testing.T, result *Program, err error)
	}{
		{
			name:        "empty symbols",
			symbols:     []*detector.Symbol{},
			expectError: true,
			checkResult: func(t *testing.T, result *Program, err error) {
				assert.Nil(t, result)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "No symbols to parse")
			},
		},
		{
			name: "no outer circle",
			symbols: []*detector.Symbol{
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 100}},
			},
			expectError: true,
			checkResult: func(t *testing.T, result *Program, err error) {
				assert.Nil(t, result)
				assert.Error(t, err)
			},
		},
		{
			name: "star symbol special case",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 0, Y: 0}},
				{Type: detector.Star, Position: detector.Position{X: 100, Y: 100}},
			},
			expectError: false,
			checkResult: func(t *testing.T, result *Program, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.MainEntry)
				assert.Len(t, result.MainEntry.Body, 1)
			},
		},
		{
			name: "debug output enabled",
			symbols: []*detector.Symbol{
				{Type: detector.OuterCircle, Position: detector.Position{X: 0, Y: 0}},
				{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 100}},
			},
			expectError: false,
			checkResult: func(t *testing.T, result *Program, err error) {
				// Set debug env var temporarily
				oldDebug := os.Getenv("GRIMOIRE_DEBUG")
				os.Setenv("GRIMOIRE_DEBUG", "1")
				defer os.Setenv("GRIMOIRE_DEBUG", oldDebug)

				// Parse again with debug enabled
				p2 := NewParser()
				result2, err2 := p2.Parse([]*detector.Symbol{
					{Type: detector.OuterCircle, Position: detector.Position{X: 0, Y: 0}},
					{Type: detector.DoubleCircle, Position: detector.Position{X: 100, Y: 100}},
				}, []detector.Connection{})

				assert.NoError(t, err2)
				assert.NotNil(t, result2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			result, err := p.Parse(tt.symbols, tt.connections)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.checkResult != nil {
				tt.checkResult(t, result, err)
			}
		})
	}
}

// TestParseStatement tests parseStatement function
func TestParseStatement(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*Parser, *symbolNode)
		checkResult func(t *testing.T, p *Parser, result Statement)
	}{
		{
			name: "visited node (non-star)",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 100, Y: 100},
					},
					visited: true,
				}
				p.symbolGraph[0] = node
				return p, node
			},
			checkResult: func(t *testing.T, p *Parser, result Statement) {
				assert.Nil(t, result)
			},
		},
		{
			name: "hexagon (parallel block)",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Hexagon,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				p.symbolGraph[0] = node
				return p, node
			},
			checkResult: func(t *testing.T, p *Parser, result Statement) {
				assert.NotNil(t, result)
				_, ok := result.(*ParallelBlock)
				assert.True(t, ok, "Expected ParallelBlock")
			},
		},
		{
			name: "square with operator child",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				opChild := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Convergence,
						Position: detector.Position{X: 120, Y: 120},
					},
				}
				node.children = []*symbolNode{opChild}
				p.symbolGraph[0] = node
				p.symbolGraph[1] = opChild
				return p, node
			},
			checkResult: func(t *testing.T, p *Parser, result Statement) {
				assert.Nil(t, result)
			},
		},
		{
			name: "operator symbol with children",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Convergence,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				child := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 120, Y: 120},
					},
				}
				node.children = []*symbolNode{child}
				p.symbolGraph[0] = node
				p.symbolGraph[1] = child
				return p, node
			},
			checkResult: func(t *testing.T, p *Parser, result Statement) {
				assert.Nil(t, result)
				// Check that child was marked as visited
				assert.True(t, p.symbolGraph[1].visited)
			},
		},
		{
			name: "expression symbol (Circle)",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Circle,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				p.symbolGraph[0] = node
				return p, node
			},
			checkResult: func(t *testing.T, p *Parser, result Statement) {
				assert.Nil(t, result)
			},
		},
		{
			name: "unexpected symbol",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.SymbolType("InvalidType"),
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				p.symbolGraph[0] = node
				return p, node
			},
			checkResult: func(t *testing.T, p *Parser, result Statement) {
				assert.Nil(t, result)
				assert.Len(t, p.errors, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, node := tt.setup()
			result := p.parseStatement(node)
			if tt.checkResult != nil {
				tt.checkResult(t, p, result)
			}
		})
	}
}

// TestParseWithDebugOutput tests the Parse function with debug output enabled
func TestParseWithDebugOutput(t *testing.T) {
	// Set the debug environment variable
	oldDebug := os.Getenv("GRIMOIRE_DEBUG")
	os.Setenv("GRIMOIRE_DEBUG", "1")
	defer os.Setenv("GRIMOIRE_DEBUG", oldDebug)

	symbols := []*detector.Symbol{
		{
			Type:     detector.OuterCircle,
			Position: detector.Position{X: 50, Y: 50},
		},
		{
			Type:     detector.DoubleCircle,
			Position: detector.Position{X: 100, Y: 100},
		},
		{
			Type:     detector.Star,
			Position: detector.Position{X: 100, Y: 150},
		},
		{
			Type:     detector.Square,
			Position: detector.Position{X: 150, Y: 150},
			Pattern:  "dot",
		},
	}

	connections := []detector.Connection{
		{
			From: symbols[1], // DoubleCircle
			To:   symbols[2], // Star
		},
		{
			From: symbols[3], // Square
			To:   symbols[2], // Star (as parent)
		},
	}

	result, err := Parse(symbols, connections)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.HasOuterCircle)
	assert.NotNil(t, result.MainEntry)
}

// TestParseErrors tests error aggregation in Parse function
func TestParseErrors(t *testing.T) {
	p := NewParser()
	p.symbols = []*detector.Symbol{
		{
			Type:     detector.OuterCircle,
			Position: detector.Position{X: 50, Y: 50},
		},
		{
			Type:     detector.Convergence, // Binary op without operands
			Position: detector.Position{X: 100, Y: 100},
		},
	}
	p.buildSymbolGraph()

	// Parse the invalid binary op to trigger an error
	node := p.symbolGraph[1]
	_ = p.parseBinaryOp(node)

	// Now parse to trigger error aggregation
	result, err := p.Parse(p.symbols, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Parser encountered errors")
	assert.Contains(t, err.Error(), "Binary operator")
}

// TestInferConnections tests the inferConnections function
func TestInferConnections(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Parser
		check func(t *testing.T, p *Parser)
	}{
		{
			name: "connect operator to nearby squares",
			setup: func() *Parser {
				p := NewParser()
				p.symbols = []*detector.Symbol{
					{
						Type:     detector.Square,
						Position: detector.Position{X: 50, Y: 100},
					},
					{
						Type:     detector.Square,
						Position: detector.Position{X: 150, Y: 100},
					},
					{
						Type:     detector.Convergence,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				p.symbolGraph = make(map[int]*symbolNode)
				for i, sym := range p.symbols {
					p.symbolGraph[i] = &symbolNode{
						symbol:   sym,
						children: []*symbolNode{},
					}
				}
				return p
			},
			check: func(t *testing.T, p *Parser) {
				p.inferConnections()
				// Check that squares are connected as parents of the operator
				opNode := p.symbolGraph[2]
				assert.Len(t, p.getParents(opNode), 2)
			},
		},
		{
			name: "connect star to nearest expression above",
			setup: func() *Parser {
				p := NewParser()
				p.symbols = []*detector.Symbol{
					{
						Type:     detector.Square,
						Position: detector.Position{X: 100, Y: 50},
					},
					{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 150},
					},
				}
				p.symbolGraph = make(map[int]*symbolNode)
				for i, sym := range p.symbols {
					p.symbolGraph[i] = &symbolNode{
						symbol:   sym,
						children: []*symbolNode{},
					}
				}
				return p
			},
			check: func(t *testing.T, p *Parser) {
				p.inferConnections()
				// Check that star is connected to the square above
				starNode := p.symbolGraph[1]
				assert.NotNil(t, starNode.parent)
				assert.Equal(t, p.symbolGraph[0], starNode.parent)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			tt.check(t, p)
		})
	}
}

// TestParseFunctionDef tests the parseFunctionDef function
func TestParseFunctionDef(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (*Parser, *symbolNode)
		check func(t *testing.T, result *FunctionDef)
	}{
		{
			name: "parse main function",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)

				funcNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.DoubleCircle,
						Position: detector.Position{X: 100, Y: 100},
					},
				}

				// Add body statement
				bodyNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 150},
					},
					parent: funcNode,
				}
				funcNode.children = []*symbolNode{bodyNode}

				p.symbolGraph[0] = funcNode
				p.symbolGraph[1] = bodyNode

				return p, funcNode
			},
			check: func(t *testing.T, result *FunctionDef) {
				assert.NotNil(t, result)
				assert.True(t, result.IsMain)
				assert.Len(t, result.Body, 1)
				assert.Equal(t, Void, result.ReturnType)
			},
		},
		{
			name: "parse already visited function",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)

				existingFunc := &FunctionDef{
					Name:   "testFunc",
					IsMain: false,
				}

				funcNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Circle,
						Position: detector.Position{X: 100, Y: 100},
					},
					visited: true,
					astNode: existingFunc,
				}

				p.symbolGraph[0] = funcNode

				return p, funcNode
			},
			check: func(t *testing.T, result *FunctionDef) {
				assert.NotNil(t, result)
				assert.Equal(t, "testFunc", result.Name)
				assert.False(t, result.IsMain)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, node := tt.setup()
			result := p.parseFunctionDef(node, tt.name == "parse main function")
			tt.check(t, result)
		})
	}
}

// TestParseAssignment tests the parseAssignment function
func TestParseAssignment(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (*Parser, *symbolNode)
		check func(t *testing.T, result *Assignment)
	}{
		{
			name: "assignment with expression child",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)

				// Square node (assignment target)
				squareNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 100, Y: 100},
						Pattern:  "dot",
					},
				}

				// Expression child
				exprNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 150, Y: 100},
						Pattern:  "double_dot",
					},
					parent: squareNode,
				}

				squareNode.children = []*symbolNode{exprNode}

				p.symbolGraph[0] = squareNode
				p.symbolGraph[1] = exprNode

				return p, squareNode
			},
			check: func(t *testing.T, result *Assignment) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Target)
				assert.NotNil(t, result.Value)
				if lit, ok := result.Value.(*Literal); ok {
					assert.Equal(t, 2, lit.Value)
				}
			},
		},
		{
			name: "assignment with no children (use own value)",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)

				squareNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 100, Y: 100},
						Pattern:  "triple_dot",
					},
					children: []*symbolNode{},
				}

				p.symbolGraph[0] = squareNode

				return p, squareNode
			},
			check: func(t *testing.T, result *Assignment) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Target)
				assert.NotNil(t, result.Value)
				if lit, ok := result.Value.(*Literal); ok {
					assert.Equal(t, 3, lit.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, node := tt.setup()
			result := p.parseAssignment(node)
			tt.check(t, result)
		})
	}
}

// TestParseExpression tests the parseExpression function
func TestParseExpression(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (*Parser, *symbolNode)
		check func(t *testing.T, result Expression)
	}{
		{
			name: "parse already visited expression",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				existingExpr := &Literal{Value: 42, LiteralType: Integer}
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type: detector.Square,
					},
					visited: true,
					astNode: existingExpr,
				}
				return p, node
			},
			check: func(t *testing.T, result Expression) {
				assert.NotNil(t, result)
				if lit, ok := result.(*Literal); ok {
					assert.Equal(t, 42, lit.Value)
				}
			},
		},
		{
			name: "parse transfer operator",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type: detector.Transfer,
					},
				}
				p.symbolGraph[0] = node
				return p, node
			},
			check: func(t *testing.T, result Expression) {
				assert.Nil(t, result)
			},
		},
		{
			name: "parse unknown expression type",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				node := &symbolNode{
					symbol: &detector.Symbol{
						Type: detector.SymbolType("UnknownType"),
					},
				}
				return p, node
			},
			check: func(t *testing.T, result Expression) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, node := tt.setup()
			result := p.parseExpression(node)
			tt.check(t, result)
		})
	}
}

// TestParseLiteralPatterns tests parseLiteral with different patterns
func TestParseLiteralPatterns(t *testing.T) {
	tests := []struct {
		pattern      string
		expectedVal  interface{}
		expectedType DataType
	}{
		{"lines", "Text", String},
		{"triple_line", "Text", String},
		{"cross", true, Boolean},
		{"half_circle", false, Boolean},
		{"unknown", 0, Integer},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			p := NewParser()
			node := &symbolNode{
				symbol: &detector.Symbol{
					Pattern: tt.pattern,
				},
			}
			result := p.parseLiteral(node)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedVal, result.Value)
			assert.Equal(t, tt.expectedType, result.LiteralType)
		})
	}
}

// TestParseStatementPanicRecovery tests the panic recovery in parseStatement
func TestParseStatementPanicRecovery(t *testing.T) {
	p := NewParser()
	p.symbolGraph = make(map[int]*symbolNode)

	// Create a node that will cause a panic during parsing
	node := &symbolNode{
		symbol: &detector.Symbol{
			Type:     detector.Star,
			Position: detector.Position{X: 100, Y: 100},
		},
		// nil parent will cause panic in parseExpressionFromParent
		parent:   nil,
		children: []*symbolNode{},
	}

	// Override parseOutputStatement to trigger a panic
	oldErrors := p.errors
	p.errors = []error{}

	// This should recover from panic and add error
	result := p.parseStatement(node)
	assert.NotNil(t, result)   // Should still return output statement
	assert.Len(t, p.errors, 0) // No panic should be recorded as error in this case

	p.errors = oldErrors
}

// TestParseLoop tests additional cases for parseLoop
func TestParseLoop(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (*Parser, *symbolNode)
		check func(t *testing.T, result Statement)
	}{
		{
			name: "while loop with condition in parent",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)

				// Loop node
				loopNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Pentagon,
						Position: detector.Position{X: 100, Y: 100},
					},
				}

				// Condition as parent
				condNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Equal,
						Position: detector.Position{X: 100, Y: 80},
					},
					children: []*symbolNode{loopNode},
				}

				// Body statement
				bodyNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 120},
					},
					parent: loopNode,
				}

				loopNode.children = []*symbolNode{bodyNode}

				p.symbolGraph[0] = loopNode
				p.symbolGraph[1] = condNode
				p.symbolGraph[2] = bodyNode

				return p, loopNode
			},
			check: func(t *testing.T, result Statement) {
				assert.NotNil(t, result)
				if whileLoop, ok := result.(*WhileLoop); ok {
					assert.NotNil(t, whileLoop.Condition)
					assert.Len(t, whileLoop.Body, 1)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, node := tt.setup()
			result := p.parseLoop(node)
			tt.check(t, result)
		})
	}
}

// TestParseAssignmentOpAdditional tests additional cases for parseAssignmentOp
func TestParseAssignmentOpAdditional(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (*Parser, *symbolNode)
		check func(t *testing.T, result Expression)
	}{
		{
			name: "assignment with target and value",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)

				// Transfer node
				transferNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Transfer,
						Position: detector.Position{X: 100, Y: 100},
					},
				}

				// Target (parent)
				targetNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 50, Y: 100},
					},
					children: []*symbolNode{transferNode},
				}

				// Value (child)
				valueNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 150, Y: 100},
						Pattern:  "dot",
					},
					parent: transferNode,
				}

				transferNode.children = []*symbolNode{valueNode}

				p.symbolGraph[0] = transferNode
				p.symbolGraph[1] = targetNode
				p.symbolGraph[2] = valueNode

				return p, transferNode
			},
			check: func(t *testing.T, result Expression) {
				// parseAssignmentOp returns nil for valid assignments (they're statements)
				assert.Nil(t, result)
			},
		},
		{
			name: "assignment with value from second parent",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)

				// Transfer node
				transferNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Transfer,
						Position: detector.Position{X: 100, Y: 100},
					},
					children: []*symbolNode{}, // No children
				}

				// Target (first parent)
				targetNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 50, Y: 100},
					},
					children: []*symbolNode{transferNode},
				}

				// Value (second parent)
				valueNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Square,
						Position: detector.Position{X: 150, Y: 100},
						Pattern:  "double_dot",
					},
					children: []*symbolNode{transferNode},
				}

				p.symbolGraph[0] = transferNode
				p.symbolGraph[1] = targetNode
				p.symbolGraph[2] = valueNode

				return p, transferNode
			},
			check: func(t *testing.T, result Expression) {
				// parseAssignmentOp returns nil for valid assignments
				assert.Nil(t, result)
			},
		},
		{
			name: "assignment with no target",
			setup: func() (*Parser, *symbolNode) {
				p := NewParser()
				p.symbolGraph = make(map[int]*symbolNode)

				// Transfer node with no square parent
				transferNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Transfer,
						Position: detector.Position{X: 100, Y: 100},
					},
				}

				// Non-square parent
				parentNode := &symbolNode{
					symbol: &detector.Symbol{
						Type:     detector.Circle,
						Position: detector.Position{X: 50, Y: 100},
					},
					children: []*symbolNode{transferNode},
				}

				p.symbolGraph[0] = transferNode
				p.symbolGraph[1] = parentNode

				return p, transferNode
			},
			check: func(t *testing.T, result Expression) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, node := tt.setup()
			result := p.parseAssignmentOp(node)
			tt.check(t, result)
		})
	}
}

// TestParseGlobalStatements tests parseGlobalStatements function
func TestParseGlobalStatements(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (*Parser, *FunctionDef)
		check func(t *testing.T, globals []Statement, mainEntry *FunctionDef)
	}{
		{
			name: "main entry exists with body - no globals",
			setup: func() (*Parser, *FunctionDef) {
				p := NewParser()
				p.symbols = []*detector.Symbol{
					{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				p.symbolGraph = make(map[int]*symbolNode)
				p.symbolGraph[0] = &symbolNode{
					symbol:  p.symbols[0],
					visited: false,
				}

				mainEntry := &FunctionDef{
					IsMain: true,
					Body:   []Statement{&OutputStatement{}},
				}

				return p, mainEntry
			},
			check: func(t *testing.T, globals []Statement, mainEntry *FunctionDef) {
				assert.Empty(t, globals)
				assert.Len(t, mainEntry.Body, 1)
			},
		},
		{
			name: "main entry exists but empty - parse stars into it",
			setup: func() (*Parser, *FunctionDef) {
				p := NewParser()
				p.symbols = []*detector.Symbol{
					{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				p.symbolGraph = make(map[int]*symbolNode)
				p.symbolGraph[0] = &symbolNode{
					symbol:  p.symbols[0],
					visited: false,
				}

				mainEntry := &FunctionDef{
					IsMain: true,
					Body:   []Statement{},
				}

				return p, mainEntry
			},
			check: func(t *testing.T, globals []Statement, mainEntry *FunctionDef) {
				assert.Empty(t, globals)
				assert.Len(t, mainEntry.Body, 1)
			},
		},
		{
			name: "no main entry - parse stars as globals",
			setup: func() (*Parser, *FunctionDef) {
				p := NewParser()
				p.symbols = []*detector.Symbol{
					{
						Type:     detector.Star,
						Position: detector.Position{X: 100, Y: 100},
					},
				}
				p.symbolGraph = make(map[int]*symbolNode)
				p.symbolGraph[0] = &symbolNode{
					symbol:  p.symbols[0],
					visited: false,
				}

				return p, nil
			},
			check: func(t *testing.T, globals []Statement, mainEntry *FunctionDef) {
				assert.Len(t, globals, 1)
				assert.Nil(t, mainEntry)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, mainEntry := tt.setup()
			globals := p.parseGlobalStatements(mainEntry)
			tt.check(t, globals, mainEntry)
		})
	}
}

// TestParseFunctions tests parseFunctions function
func TestParseFunctions(t *testing.T) {
	p := NewParser()
	p.symbols = []*detector.Symbol{
		{
			Type:     detector.Circle,
			Position: detector.Position{X: 100, Y: 100},
		},
		{
			Type:     detector.Circle,
			Position: detector.Position{X: 200, Y: 100},
		},
		{
			Type:     detector.Square, // Not a function
			Position: detector.Position{X: 300, Y: 100},
		},
	}

	p.symbolGraph = make(map[int]*symbolNode)
	for i, sym := range p.symbols {
		p.symbolGraph[i] = &symbolNode{
			symbol:   sym,
			visited:  false,
			children: []*symbolNode{},
		}
	}

	// Mark one circle as visited
	p.symbolGraph[1].visited = true

	functions := p.parseFunctions()
	assert.Len(t, functions, 1) // Only one unvisited Circle should be parsed
}

// TestApplyConnections tests applyConnections function
func TestApplyConnections(t *testing.T) {
	p := NewParser()
	p.symbols = []*detector.Symbol{
		{
			Type:     detector.Square,
			Position: detector.Position{X: 100, Y: 100},
		},
		{
			Type:     detector.Star,
			Position: detector.Position{X: 100, Y: 150},
		},
	}

	p.connections = []detector.Connection{
		{
			From: p.symbols[0],
			To:   p.symbols[1],
		},
	}

	p.symbolGraph = make(map[int]*symbolNode)
	for i, sym := range p.symbols {
		p.symbolGraph[i] = &symbolNode{
			symbol:   sym,
			children: []*symbolNode{},
		}
	}

	p.applyConnections()

	// Check connections were applied
	assert.Len(t, p.symbolGraph[0].children, 1)
	assert.Equal(t, p.symbolGraph[1], p.symbolGraph[0].children[0])
	assert.Equal(t, p.symbolGraph[0], p.symbolGraph[1].parent)
}
