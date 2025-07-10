package parser

import (
	"fmt"

	"github.com/ayutaz/grimoire/internal/detector"
)

// Parser converts detected symbols to AST
type Parser struct {
	symbols     []detector.Symbol
	connections []detector.Connection
	errors      []error
}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse converts symbols to AST
func Parse(symbols []detector.Symbol) (*Program, error) {
	parser := NewParser()
	return parser.Parse(symbols)
}

// Parse performs the parsing
func (p *Parser) Parse(symbols []detector.Symbol) (*Program, error) {
	p.symbols = symbols

	// Find outer circle
	var outerCircle *detector.Symbol
	for i := range symbols {
		if symbols[i].Type == detector.OuterCircle {
			outerCircle = &symbols[i]
			break
		}
	}

	if outerCircle == nil {
		return nil, fmt.Errorf("no outer circle detected: all Grimoire programs must be enclosed in a magic circle")
	}

	// Create program node
	program := &Program{
		HasOuterCircle: true,
		Functions:      []*FunctionDef{},
		Globals:        []Statement{},
	}

	// Find main entry (double circle)
	for i := range symbols {
		if symbols[i].Type == detector.DoubleCircle {
			program.MainEntry = p.parseFunctionDef(&symbols[i], true)
			break
		}
	}

	// Parse other symbols
	// TODO: Implement full parsing logic

	// For now, create a simple Hello World program
	if program.MainEntry == nil {
		program.MainEntry = &FunctionDef{
			IsMain: true,
			Body: []Statement{
				&OutputStatement{
					Value: &Literal{
						Value:       "Hello, World!",
						LiteralType: String,
					},
				},
			},
		}
	}

	return program, nil
}

// parseFunctionDef parses a function definition
func (p *Parser) parseFunctionDef(symbol *detector.Symbol, isMain bool) *FunctionDef {
	return &FunctionDef{
		Name:       "",
		Parameters: []*Parameter{},
		Body:       []Statement{},
		ReturnType: Void,
		IsMain:     isMain,
	}
}

// parseStatement parses a statement from a symbol
func (p *Parser) parseStatement(symbol *detector.Symbol) Statement {
	switch symbol.Type {
	case detector.Star:
		return p.parseOutputStatement(symbol)
	case detector.Triangle:
		return p.parseIfStatement(symbol)
	case detector.Pentagon:
		return p.parseLoop(symbol)
	case detector.Hexagon:
		return p.parseParallelBlock(symbol)
	default:
		return nil
	}
}

// parseOutputStatement parses an output statement (star)
func (p *Parser) parseOutputStatement(symbol *detector.Symbol) *OutputStatement {
	// TODO: Find connected expression
	return &OutputStatement{
		Value: &Literal{
			Value:       "Output",
			LiteralType: String,
		},
	}
}

// parseIfStatement parses a conditional (triangle)
func (p *Parser) parseIfStatement(symbol *detector.Symbol) *IfStatement {
	return &IfStatement{
		Condition: &Literal{
			Value:       true,
			LiteralType: Boolean,
		},
		ThenBranch: []Statement{},
		ElseBranch: []Statement{},
	}
}

// parseLoop parses a loop (pentagon)
func (p *Parser) parseLoop(symbol *detector.Symbol) Statement {
	// TODO: Determine if for or while loop
	return &WhileLoop{
		Condition: &Literal{
			Value:       false,
			LiteralType: Boolean,
		},
		Body: []Statement{},
	}
}

// parseParallelBlock parses a parallel block (hexagon)
func (p *Parser) parseParallelBlock(symbol *detector.Symbol) *ParallelBlock {
	return &ParallelBlock{
		Branches: [][]Statement{},
	}
}

// parseExpression parses an expression from a symbol
func (p *Parser) parseExpression(symbol *detector.Symbol) Expression {
	switch symbol.Type {
	case detector.Square:
		return p.parseLiteral(symbol)
	case detector.Circle:
		return p.parseFunctionCall(symbol)
	case detector.Convergence, detector.Divergence, detector.Amplification, detector.Distribution:
		return p.parseBinaryOp(symbol)
	default:
		return nil
	}
}

// parseLiteral parses a literal from symbol properties
func (p *Parser) parseLiteral(symbol *detector.Symbol) *Literal {
	pattern := symbol.Pattern
	
	switch pattern {
	case "dot":
		return &Literal{Value: 1, LiteralType: Integer}
	case "double_dot":
		return &Literal{Value: 2, LiteralType: Integer}
	case "triple_dot":
		return &Literal{Value: 3, LiteralType: Integer}
	case "empty":
		return &Literal{Value: 0, LiteralType: Integer}
	case "lines", "triple_line":
		return &Literal{Value: "Text", LiteralType: String}
	case "cross":
		return &Literal{Value: true, LiteralType: Boolean}
	case "half_circle":
		return &Literal{Value: false, LiteralType: Boolean}
	default:
		return &Literal{Value: 0, LiteralType: Integer}
	}
}

// parseFunctionCall parses a function call
func (p *Parser) parseFunctionCall(symbol *detector.Symbol) *FunctionCall {
	return &FunctionCall{
		Function:  &Identifier{Name: "print"},
		Arguments: []Expression{},
		DataType:  Void,
	}
}

// parseBinaryOp parses a binary operation
func (p *Parser) parseBinaryOp(symbol *detector.Symbol) *BinaryOp {
	var op OperatorType
	switch symbol.Type {
	case detector.Convergence:
		op = Add
	case detector.Divergence:
		op = Subtract
	case detector.Amplification:
		op = Multiply
	case detector.Distribution:
		op = Divide
	default:
		op = Add
	}

	return &BinaryOp{
		Left:     &Literal{Value: 0, LiteralType: Integer},
		Operator: op,
		Right:    &Literal{Value: 0, LiteralType: Integer},
		DataType: Integer,
	}
}