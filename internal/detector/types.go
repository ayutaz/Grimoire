package detector

// SymbolType represents the type of detected symbol
type SymbolType string

const (
	// Structure elements
	OuterCircle    SymbolType = "outer_circle"
	Circle         SymbolType = "circle"
	DoubleCircle   SymbolType = "double_circle"
	Square         SymbolType = "square"
	Triangle       SymbolType = "triangle"
	Pentagon       SymbolType = "pentagon"
	Hexagon        SymbolType = "hexagon"
	Star           SymbolType = "star"
	SixPointedStar SymbolType = "six_pointed_star"
	EightPointedStar SymbolType = "eight_pointed_star"

	// Operators
	Convergence    SymbolType = "convergence"    // Addition
	Divergence     SymbolType = "divergence"     // Subtraction
	Amplification  SymbolType = "amplification"  // Multiplication
	Distribution   SymbolType = "distribution"   // Division
	Transfer       SymbolType = "transfer"       // Assignment
	Seal           SymbolType = "seal"           // Constant
	Circulation    SymbolType = "circulation"    // Loop

	// Comparison operators
	Equal         SymbolType = "equal"
	NotEqual      SymbolType = "not_equal"
	LessThan      SymbolType = "less_than"
	GreaterThan   SymbolType = "greater_than"
	LessEqual     SymbolType = "less_equal"
	GreaterEqual  SymbolType = "greater_equal"

	// Logical operators
	LogicalAnd    SymbolType = "logical_and"
	LogicalOr     SymbolType = "logical_or"
	LogicalNot    SymbolType = "logical_not"
	LogicalXor    SymbolType = "logical_xor"

	// Special
	ConnectionSymbol    SymbolType = "connection"
	Unknown            SymbolType = "unknown"
)

// Position represents a position in the image
type Position struct {
	X float64
	Y float64
}

// Symbol represents a detected symbol in the image
type Symbol struct {
	Type       SymbolType
	Position   Position
	Size       float64
	Confidence float64
	Pattern    string // Internal pattern (dots, lines, etc.)
	Properties map[string]interface{}
}

// Connection represents a connection between symbols
type Connection struct {
	From           *Symbol
	To             *Symbol
	ConnectionType string // solid, dashed, wavy, etc.
	Properties     map[string]interface{}
}

// DetectionResult contains all detected symbols and connections
type DetectionResult struct {
	Symbols     []Symbol
	Connections []Connection
	OuterCircle *Symbol
}