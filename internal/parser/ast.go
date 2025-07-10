package parser

// DataType represents the data types in Grimoire
type DataType string

const (
	Integer DataType = "integer"
	Float   DataType = "float"
	String  DataType = "string"
	Boolean DataType = "boolean"
	Array   DataType = "array"
	Map     DataType = "map"
	Void    DataType = "void"
)

// OperatorType represents operator types
type OperatorType string

const (
	// Arithmetic
	Add      OperatorType = "add"
	Subtract OperatorType = "subtract"
	Multiply OperatorType = "multiply"
	Divide   OperatorType = "divide"
	
	// Comparison
	Equal        OperatorType = "equal"
	NotEqual     OperatorType = "not_equal"
	LessThan     OperatorType = "less_than"
	GreaterThan  OperatorType = "greater_than"
	LessEqual    OperatorType = "less_equal"
	GreaterEqual OperatorType = "greater_equal"
	
	// Logical
	And OperatorType = "and"
	Or  OperatorType = "or"
	Not OperatorType = "not"
	Xor OperatorType = "xor"
	
	// Assignment
	Assign OperatorType = "assign"
)

// ASTNode is the base interface for all AST nodes
type ASTNode interface {
	node()
}

// Program is the root node
type Program struct {
	HasOuterCircle bool
	MainEntry      *FunctionDef
	Functions      []*FunctionDef
	Globals        []Statement
}

func (*Program) node() {}

// FunctionDef represents a function definition
type FunctionDef struct {
	Name       string
	Parameters []*Parameter
	Body       []Statement
	ReturnType DataType
	IsMain     bool
}

func (*FunctionDef) node() {}

// Parameter represents a function parameter
type Parameter struct {
	Name         string
	DataType     DataType
	DefaultValue Expression
}

// Statement nodes
type Statement interface {
	ASTNode
	statement()
}

// Expression nodes
type Expression interface {
	ASTNode
	expression()
	Type() DataType
}

// Assignment statement
type Assignment struct {
	Target *Identifier
	Value  Expression
}

func (*Assignment) node()      {}
func (*Assignment) statement() {}

// OutputStatement represents output (star symbol)
type OutputStatement struct {
	Value Expression
}

func (*OutputStatement) node()      {}
func (*OutputStatement) statement() {}

// IfStatement represents conditional
type IfStatement struct {
	Condition  Expression
	ThenBranch []Statement
	ElseBranch []Statement
}

func (*IfStatement) node()      {}
func (*IfStatement) statement() {}

// ForLoop represents a for loop
type ForLoop struct {
	Counter *Identifier
	Start   Expression
	End     Expression
	Step    Expression
	Body    []Statement
}

func (*ForLoop) node()      {}
func (*ForLoop) statement() {}

// WhileLoop represents a while loop
type WhileLoop struct {
	Condition Expression
	Body      []Statement
}

func (*WhileLoop) node()      {}
func (*WhileLoop) statement() {}

// ParallelBlock represents parallel execution
type ParallelBlock struct {
	Branches [][]Statement
}

func (*ParallelBlock) node()      {}
func (*ParallelBlock) statement() {}

// ReturnStatement represents a return statement
type ReturnStatement struct {
	Value Expression
}

func (*ReturnStatement) node()      {}
func (*ReturnStatement) statement() {}

// ExpressionStatement wraps an expression as a statement
type ExpressionStatement struct {
	Expression Expression
}

func (*ExpressionStatement) node()      {}
func (*ExpressionStatement) statement() {}

// BinaryOp represents binary operations
type BinaryOp struct {
	Left     Expression
	Operator OperatorType
	Right    Expression
	DataType DataType
}

func (*BinaryOp) node()       {}
func (*BinaryOp) expression() {}
func (b *BinaryOp) Type() DataType { return b.DataType }

// UnaryOp represents unary operations
type UnaryOp struct {
	Operator OperatorType
	Operand  Expression
	DataType DataType
}

func (*UnaryOp) node()       {}
func (*UnaryOp) expression() {}
func (u *UnaryOp) Type() DataType { return u.DataType }

// Literal represents literal values
type Literal struct {
	Value       interface{}
	LiteralType DataType
}

func (*Literal) node()       {}
func (*Literal) expression() {}
func (l *Literal) Type() DataType { return l.LiteralType }

// Identifier represents variables
type Identifier struct {
	Name     string
	DataType DataType
}

func (*Identifier) node()       {}
func (*Identifier) expression() {}
func (i *Identifier) Type() DataType { return i.DataType }

// FunctionCall represents function calls
type FunctionCall struct {
	Function  *Identifier
	Arguments []Expression
	DataType  DataType
}

func (*FunctionCall) node()       {}
func (*FunctionCall) expression() {}
func (f *FunctionCall) Type() DataType { return f.DataType }

// ArrayLiteral represents array literals
type ArrayLiteral struct {
	Elements []Expression
}

func (*ArrayLiteral) node()       {}
func (*ArrayLiteral) expression() {}
func (*ArrayLiteral) Type() DataType { return Array }

// MapLiteral represents map literals
type MapLiteral struct {
	Pairs [][2]Expression // [key, value] pairs
}

func (*MapLiteral) node()       {}
func (*MapLiteral) expression() {}
func (*MapLiteral) Type() DataType { return Map }