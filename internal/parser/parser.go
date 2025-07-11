package parser

import (
	"fmt"
	"math"
	"os"

	"github.com/ayutaz/grimoire/internal/detector"
	grimoireErrors "github.com/ayutaz/grimoire/internal/errors"
)

// symbolNode wraps a symbol with parsing metadata
type symbolNode struct {
	symbol   *detector.Symbol
	visited  bool
	astNode  ASTNode
	parent   *symbolNode
	children []*symbolNode
}

// Parser converts detected symbols to AST
type Parser struct {
	symbols     []*detector.Symbol
	connections []detector.Connection
	symbolGraph map[int]*symbolNode
	errors      []error
}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{
		symbolGraph: make(map[int]*symbolNode),
	}
}

// Parse converts symbols to AST
func Parse(symbols []*detector.Symbol, connections []detector.Connection) (*Program, error) {
	parser := NewParser()
	return parser.Parse(symbols, connections)
}

// Parse performs the parsing
func (p *Parser) Parse(symbols []*detector.Symbol, connections []detector.Connection) (*Program, error) {
	p.symbols = symbols
	p.connections = connections

	// Validate input
	if len(symbols) == 0 {
		return nil, grimoireErrors.NewError(grimoireErrors.SyntaxError, "No symbols to parse").
			WithDetails("The input contains no detected symbols")
	}

	// Build symbol graph
	p.buildSymbolGraph()

	// Debug: print symbol graph
	if os.Getenv("GRIMOIRE_DEBUG") != "" {
		fmt.Printf("Symbol graph:\n")
		for i, node := range p.symbolGraph {
			fmt.Printf("[%d] %s: %d children\n", i, p.symbols[i].Type, len(node.children))
			for _, child := range node.children {
				childIdx := -1
				for k, n := range p.symbolGraph {
					if n == child {
						childIdx = k
						break
					}
				}
				if childIdx >= 0 {
					fmt.Printf("  -> [%d] %s\n", childIdx, p.symbols[childIdx].Type)
				}
			}
		}
	}

	// Find outer circle
	var outerCircle *detector.Symbol
	for _, symbol := range symbols {
		if symbol.Type == detector.OuterCircle {
			outerCircle = symbol
			break
		}
	}

	if outerCircle == nil {
		return nil, grimoireErrors.NoOuterCircleError()
	}

	// Find main entry (double circle)
	var mainEntry *FunctionDef
	for i, symbol := range symbols {
		if symbol.Type == detector.DoubleCircle {
			node := p.symbolGraph[i]
			mainEntry = p.parseFunctionDef(node, true)
			break
		}
	}

	// Parse functions (circles)
	functions := p.parseFunctions()

	// Parse global statements
	globals := p.parseGlobalStatements(mainEntry)

	// If we have global statements but no main entry, create implicit main
	if mainEntry == nil && len(globals) > 0 {
		mainEntry = &FunctionDef{
			IsMain: true,
			Body:   globals,
		}
		globals = []Statement{}
	}

	// Special case: check if we have a star symbol
	if mainEntry == nil || len(mainEntry.Body) == 0 {
		for i, symbol := range symbols {
			if symbol.Type == detector.Star {
				node := p.symbolGraph[i]
				stmt := p.parseStatement(node)
				if stmt != nil {
					if mainEntry == nil {
						mainEntry = &FunctionDef{
							IsMain: true,
							Body:   []Statement{stmt},
						}
					} else {
						mainEntry.Body = []Statement{stmt}
					}
					globals = []Statement{}
				}
				break
			}
		}
	}

	// Check if we have any errors
	if len(p.errors) > 0 {
		// Combine all errors into a single error message
		errorMsg := "Parser encountered errors:"
		for _, err := range p.errors {
			errorMsg += "\n  - " + err.Error()
		}
		return nil, grimoireErrors.NewError(grimoireErrors.SyntaxError, errorMsg)
	}

	return &Program{
		HasOuterCircle: true,
		MainEntry:      mainEntry,
		Functions:      functions,
		Globals:        globals,
	}, nil
}

// buildSymbolGraph builds a graph of symbols and their connections
func (p *Parser) buildSymbolGraph() {
	// Create nodes for all symbols
	for i, symbol := range p.symbols {
		p.symbolGraph[i] = &symbolNode{
			symbol:   symbol,
			children: []*symbolNode{},
		}
	}

	// Use explicit connections if available
	if len(p.connections) > 0 {
		p.applyConnections()
	} else {
		// Otherwise infer connections
		p.inferConnections()
	}
}

// inferConnections infers connections based on symbol positions
func (p *Parser) inferConnections() {
	// Skip outer circle
	symbolsToConnect := []*symbolNode{}
	for i, sym := range p.symbols {
		if sym.Type != detector.OuterCircle {
			symbolsToConnect = append(symbolsToConnect, p.symbolGraph[i])
		}
	}

	// Connect main entry to symbols below it
	for _, node := range symbolsToConnect {
		if node.symbol.Type == detector.DoubleCircle {
			mainY := node.symbol.Position.Y
			// Find symbols below main
			for _, other := range symbolsToConnect {
				if other != node && other.symbol.Position.Y > mainY {
					// Check horizontal alignment
					xDiff := abs(other.symbol.Position.X - node.symbol.Position.X)
					if xDiff < 150 {
						node.children = append(node.children, other)
						other.parent = node
					}
				}
			}
		}
	}

	// Connect operators to nearby operands
	for _, node := range symbolsToConnect {
		if isOperator(node.symbol.Type) {
			// Find nearby squares
			nearbyOperands := []*symbolNode{}
			for _, other := range symbolsToConnect {
				if other != node && other.symbol.Type == detector.Square {
					dist := distance(node.symbol.Position, other.symbol.Position)
					if dist < 150 {
						nearbyOperands = append(nearbyOperands, other)
					}
				}
			}
			// Connect operands as parents of operator
			for _, operand := range nearbyOperands {
				operand.children = append(operand.children, node)
			}
		}
	}

	// Connect stars to nearest expressions above
	for _, node := range symbolsToConnect {
		if node.symbol.Type == detector.Star {
			starPos := node.symbol.Position
			var nearest *symbolNode
			minDist := 999999.0

			for _, other := range symbolsToConnect {
				if other != node && other.symbol.Position.Y < starPos.Y {
					dist := distance(starPos, other.symbol.Position)
					if dist < minDist {
						minDist = dist
						nearest = other
					}
				}
			}

			if nearest != nil && minDist < 150 {
				nearest.children = append(nearest.children, node)
				node.parent = nearest
			}
		}
	}
}

// parseFunctionDef parses a function definition
func (p *Parser) parseFunctionDef(node *symbolNode, isMain bool) *FunctionDef {
	if node.visited {
		if fn, ok := node.astNode.(*FunctionDef); ok {
			return fn
		}
		return nil
	}

	node.visited = true

	// Parse function body
	body := p.parseStatementSequence(node.children)

	fn := &FunctionDef{
		Name:       "",
		Parameters: []*Parameter{},
		Body:       body,
		ReturnType: Void,
		IsMain:     isMain,
	}

	node.astNode = fn
	return fn
}

// parseFunctions parses all function definitions
func (p *Parser) parseFunctions() []*FunctionDef {
	functions := []*FunctionDef{}

	for i, symbol := range p.symbols {
		if symbol.Type == detector.Circle {
			node := p.symbolGraph[i]
			if !node.visited {
				fn := p.parseFunctionDef(node, false)
				if fn != nil {
					functions = append(functions, fn)
				}
			}
		}
	}

	return functions
}

// parseGlobalStatements parses statements not in functions
func (p *Parser) parseGlobalStatements(mainEntry *FunctionDef) []Statement {
	globals := []Statement{}

	// If main entry exists and is populated, don't parse globals
	if mainEntry != nil && len(mainEntry.Body) > 0 {
		return globals
	}

	// If main entry exists and is empty, parse star statements into it
	if mainEntry != nil && len(mainEntry.Body) == 0 {
		for i, symbol := range p.symbols {
			node := p.symbolGraph[i]
			if symbol.Type == detector.Star && !node.visited {
				stmt := p.parseStatement(node)
				if stmt != nil {
					mainEntry.Body = append(mainEntry.Body, stmt)
				}
			}
		}
		return globals
	}

	// Otherwise parse only star statements as globals
	for i, symbol := range p.symbols {
		node := p.symbolGraph[i]
		if symbol.Type == detector.Star && !node.visited {
			stmt := p.parseStatement(node)
			if stmt != nil {
				globals = append(globals, stmt)
			}
		}
	}

	return globals
}

// parseStatementSequence parses a sequence of statements
func (p *Parser) parseStatementSequence(nodes []*symbolNode) []Statement {
	stmts := []Statement{}

	for _, node := range nodes {
		if !node.visited {
			stmt := p.parseStatement(node)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		}
	}

	return stmts
}

// parseStatement parses a statement from a symbol
func (p *Parser) parseStatement(node *symbolNode) Statement {
	if node.visited && node.symbol.Type != detector.Star {
		return nil
	}

	node.visited = true
	symbol := node.symbol

	// Track parsing errors
	defer func() {
		if r := recover(); r != nil {
			err := grimoireErrors.NewError(grimoireErrors.SyntaxError, fmt.Sprintf("Panic during parsing: %v", r)).
				WithDetails(fmt.Sprintf("Symbol: %s at (%.0f, %.0f)", symbol.Type, symbol.Position.X, symbol.Position.Y))
			p.errors = append(p.errors, err)
		}
	}()

	switch symbol.Type {
	case detector.Star:
		return p.parseOutputStatement(node)
	case detector.Triangle:
		return p.parseIfStatement(node)
	case detector.Pentagon:
		return p.parseLoop(node)
	case detector.Hexagon, detector.SixPointedStar:
		return p.parseParallelBlock(node)
	case detector.Square:
		// Check if it's an assignment or part of expression
		if hasOperatorChild(node) {
			return nil // Part of expression
		}
		return p.parseAssignment(node)
	default:
		// Skip operators and other symbols
		if isOperator(symbol.Type) {
			// Mark children as visited
			for _, child := range node.children {
				child.visited = true
			}
			return nil
		}
		// Report unexpected symbol
		err := grimoireErrors.UnexpectedSymbolError(
			string(symbol.Type), "statement symbol",
			symbol.Position.X, symbol.Position.Y)
		p.errors = append(p.errors, err)
		return nil
	}
}

// parseOutputStatement parses an output statement (star)
func (p *Parser) parseOutputStatement(node *symbolNode) *OutputStatement {
	// Find expression from parent
	expr := p.parseExpressionFromParent(node)
	return &OutputStatement{
		Value: expr,
	}
}

// parseIfStatement parses a conditional (triangle)
func (p *Parser) parseIfStatement(node *symbolNode) *IfStatement {
	// Find condition
	condition := p.parseCondition(node)

	// Split children by position
	var leftChildren, rightChildren []*symbolNode
	for _, child := range node.children {
		if child.symbol.Position.X < node.symbol.Position.X {
			leftChildren = append(leftChildren, child)
		} else {
			rightChildren = append(rightChildren, child)
		}
	}

	thenBranch := p.parseStatementSequence(leftChildren)
	var elseBranch []Statement
	if len(rightChildren) > 0 {
		elseBranch = p.parseStatementSequence(rightChildren)
	}

	return &IfStatement{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}

// parseLoop parses a loop (pentagon)
func (p *Parser) parseLoop(node *symbolNode) Statement {
	// Look for counter in parent
	var counterNode *symbolNode
	for _, parent := range p.getParents(node) {
		if parent.symbol.Type == detector.Square {
			counterNode = parent
			break
		}
	}

	body := p.parseStatementSequence(node.children)

	if counterNode != nil {
		// For loop
		counter := &Identifier{Name: "i"}
		start := &Literal{Value: 0, LiteralType: Integer}
		end := p.parseLiteral(counterNode)
		step := &Literal{Value: 1, LiteralType: Integer}

		return &ForLoop{
			Counter: counter,
			Start:   start,
			End:     end,
			Step:    step,
			Body:    body,
		}
	}
	
	// While loop
	condition := p.parseCondition(node)
	return &WhileLoop{
		Condition: condition,
		Body:      body,
	}
}

// parseParallelBlock parses a parallel block (hexagon)
func (p *Parser) parseParallelBlock(node *symbolNode) *ParallelBlock {
	// Group children by angle
	groups := p.groupChildrenByAngle(node)

	branches := [][]Statement{}
	for _, group := range groups {
		branch := p.parseStatementSequence(group)
		if len(branch) > 0 {
			branches = append(branches, branch)
		}
	}

	return &ParallelBlock{
		Branches: branches,
	}
}

// parseAssignment parses an assignment statement
func (p *Parser) parseAssignment(node *symbolNode) *Assignment {
	varName := fmt.Sprintf("var_%p", node.symbol)
	target := &Identifier{Name: varName}

	// Look for value in children
	var value Expression
	for _, child := range node.children {
		value = p.parseExpression(child)
		if value != nil {
			break
		}
	}

	if value == nil {
		// Use literal from properties
		value = p.parseLiteral(node)
	}

	return &Assignment{
		Target: target,
		Value:  value,
	}
}

// parseExpression parses an expression from a symbol
func (p *Parser) parseExpression(node *symbolNode) Expression {
	if node.visited && node.astNode != nil {
		if expr, ok := node.astNode.(Expression); ok {
			return expr
		}
	}

	node.visited = true
	symbol := node.symbol

	switch symbol.Type {
	case detector.Square:
		return p.parseLiteral(node)
	case detector.Circle:
		return p.parseFunctionCall(node)
	case detector.Convergence, detector.Divergence, detector.Amplification, detector.Distribution:
		return p.parseBinaryOp(node)
	case detector.Transfer:
		return p.parseAssignmentOp(node)
	default:
		return nil
	}
}

// parseLiteral parses a literal from symbol properties
func (p *Parser) parseLiteral(node *symbolNode) *Literal {
	symbol := node.symbol
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
func (p *Parser) parseFunctionCall(node *symbolNode) *FunctionCall {
	if node.visited {
		return &FunctionCall{
			Function:  &Identifier{Name: "print"},
			Arguments: []Expression{},
			DataType:  Void,
		}
	}

	node.visited = true

	// Get arguments from parents
	arguments := []Expression{}
	for _, parent := range p.getParents(node) {
		if !parent.visited {
			arg := p.parseExpression(parent)
			if arg != nil {
				arguments = append(arguments, arg)
			}
		}
	}

	return &FunctionCall{
		Function:  &Identifier{Name: "print"},
		Arguments: arguments,
		DataType:  Void,
	}
}

// parseBinaryOp parses a binary operation
func (p *Parser) parseBinaryOp(node *symbolNode) *BinaryOp {
	symbol := node.symbol
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

	// Find operands from parents
	operands := []Expression{}
	for _, parent := range p.getParents(node) {
		if parent.symbol.Type == detector.Square {
			literal := p.parseLiteral(parent)
			if literal != nil {
				operands = append(operands, literal)
			}
		} else {
			expr := p.parseExpression(parent)
			if expr != nil {
				operands = append(operands, expr)
			}
		}
	}

	// Validate operands
	if len(operands) < 2 {
		err := grimoireErrors.NewError(grimoireErrors.UnbalancedExpression,
			fmt.Sprintf("Binary operator %s requires two operands, found %d", symbol.Type, len(operands))).
			WithDetails(fmt.Sprintf("At position (%.0f, %.0f)", symbol.Position.X, symbol.Position.Y)).
			WithSuggestion("Ensure the operator is connected to two operand symbols")
		p.errors = append(p.errors, err)
	}

	// Ensure we have two operands
	var left Expression = &Literal{Value: 0, LiteralType: Integer}
	var right Expression = &Literal{Value: 0, LiteralType: Integer}

	if len(operands) > 0 {
		left = operands[0]
	}
	if len(operands) > 1 {
		right = operands[1]
	}

	return &BinaryOp{
		Left:     left,
		Operator: op,
		Right:    right,
		DataType: Integer,
	}
}

// Helper functions

// getParents returns all parents of a node
func (p *Parser) getParents(node *symbolNode) []*symbolNode {
	parents := []*symbolNode{}
	for _, n := range p.symbolGraph {
		for _, child := range n.children {
			if child == node {
				parents = append(parents, n)
			}
		}
	}
	return parents
}

// parseExpressionFromParent parses expression from parent nodes
func (p *Parser) parseExpressionFromParent(node *symbolNode) Expression {
	// If node has a parent, parse that
	if node.parent != nil {
		expr := p.parseExpression(node.parent)
		if expr != nil {
			return expr
		}
	}

	// Otherwise look at all parents
	for _, parent := range p.getParents(node) {
		expr := p.parseExpression(parent)
		if expr != nil {
			return expr
		}
	}

	// For standalone stars, return "Hello, World!"
	if node.symbol.Type == detector.Star {
		return &Literal{
			Value:       "Hello, World!",
			LiteralType: String,
		}
	}

	return &Literal{Value: 0, LiteralType: Integer}
}

// parseCondition parses a condition expression
func (p *Parser) parseCondition(node *symbolNode) Expression {
	// Look for comparison operators in children
	for _, child := range node.children {
		if isComparisonOperator(child.symbol.Type) {
			return p.parseBinaryOp(child)
		}
	}

	// Look in parents too
	for _, parent := range p.getParents(node) {
		if isComparisonOperator(parent.symbol.Type) {
			return p.parseBinaryOp(parent)
		}
	}

	// Default to false
	return &Literal{Value: false, LiteralType: Boolean}
}

// parseAssignmentOp parses assignment using transfer operator
func (p *Parser) parseAssignmentOp(node *symbolNode) Expression {
	parents := p.getParents(node)
	children := node.children

	var target *Identifier
	var value Expression

	// Left side is usually a parent
	if len(parents) > 0 {
		parent := parents[0]
		if parent.symbol.Type == detector.Square {
			varName := fmt.Sprintf("var_%p", parent.symbol)
			target = &Identifier{Name: varName}
		}
	}

	// Right side is usually a child
	if len(children) > 0 {
		value = p.parseExpression(children[0])
	} else if len(parents) > 1 {
		value = p.parseExpression(parents[1])
	}

	if target != nil && value != nil {
		// Return assignment as expression statement
		return nil // Assignments are statements, not expressions
	}

	return nil
}

// groupChildrenByAngle groups children by their angular position
func (p *Parser) groupChildrenByAngle(node *symbolNode) [][]*symbolNode {
	if len(node.children) == 0 {
		return nil
	}

	// Simple grouping by quadrants
	groups := make([][]*symbolNode, 4)
	centerX := node.symbol.Position.X
	centerY := node.symbol.Position.Y

	for _, child := range node.children {
		dx := child.symbol.Position.X - centerX
		dy := child.symbol.Position.Y - centerY

		var group int
		if dx >= 0 && dy >= 0 {
			group = 0 // Top-right
		} else if dx < 0 && dy >= 0 {
			group = 1 // Top-left
		} else if dx < 0 && dy < 0 {
			group = 2 // Bottom-left
		} else {
			group = 3 // Bottom-right
		}

		groups[group] = append(groups[group], child)
	}

	// Remove empty groups
	result := [][]*symbolNode{}
	for _, g := range groups {
		if len(g) > 0 {
			result = append(result, g)
		}
	}

	return result
}

// Utility functions

func isOperator(t detector.SymbolType) bool {
	return t == detector.Convergence || t == detector.Divergence ||
		t == detector.Amplification || t == detector.Distribution ||
		t == detector.Transfer || t == detector.Seal ||
		t == detector.Circulation || isComparisonOperator(t) ||
		t == detector.LogicalAnd || t == detector.LogicalOr ||
		t == detector.LogicalNot || t == detector.LogicalXor
}

func isComparisonOperator(t detector.SymbolType) bool {
	return t == detector.Equal || t == detector.NotEqual ||
		t == detector.LessThan || t == detector.GreaterThan ||
		t == detector.LessEqual || t == detector.GreaterEqual
}

func hasOperatorChild(node *symbolNode) bool {
	for _, child := range node.children {
		if isOperator(child.symbol.Type) {
			return true
		}
	}
	return false
}

func distance(p1, p2 detector.Position) float64 {
	dx := float64(p1.X - p2.X)
	dy := float64(p1.Y - p2.Y)
	return dx*dx + dy*dy // Return squared distance for efficiency
}

func abs(x float64) float64 {
	return math.Abs(x)
}

// applyConnections applies explicit connections to the symbol graph
func (p *Parser) applyConnections() {
	for _, conn := range p.connections {
		// Find the indices of the connected symbols
		fromIdx := -1
		toIdx := -1

		for i, sym := range p.symbols {
			if sym == conn.From {
				fromIdx = i
			}
			if sym == conn.To {
				toIdx = i
			}
		}

		if fromIdx >= 0 && toIdx >= 0 {
			fromNode := p.symbolGraph[fromIdx]
			toNode := p.symbolGraph[toIdx]
			fromNode.children = append(fromNode.children, toNode)
			toNode.parent = fromNode
		}
	}
}
