package compiler

import (
	"strings"
	"testing"

	"github.com/ayutaz/grimoire/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompilerCreation tests that we can create a new compiler
func TestCompilerCreation(t *testing.T) {
	compiler := NewCompiler()
	assert.NotNil(t, compiler)
	assert.Equal(t, 0, compiler.indent)
	assert.Equal(t, "    ", compiler.indentStr)
}

// TestCompile_EmptyProgram tests compiling empty program
func TestCompile_EmptyProgram(t *testing.T) {
	ast := &parser.Program{
		HasOuterCircle: true,
		Functions:      []*parser.FunctionDef{},
		Globals:        []parser.Statement{},
	}

	code, err := Compile(ast)

	// Empty program should now return an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "メインエントリーポイントが見つかりません")
	assert.Empty(t, code)
}

// TestCompile_HelloWorld tests compiling hello world program
func TestCompile_HelloWorld(t *testing.T) {
	ast := &parser.Program{
		HasOuterCircle: true,
		MainEntry: &parser.FunctionDef{
			IsMain: true,
			Body: []parser.Statement{
				&parser.OutputStatement{
					Value: &parser.Literal{
						Value:       "Hello, World!",
						LiteralType: parser.String,
					},
				},
			},
		},
	}

	code, err := Compile(ast)

	require.NoError(t, err)
	assert.Contains(t, code, `if __name__ == "__main__":`)
	assert.Contains(t, code, `print("Hello, World!")`)
}

// TestCompile_Arithmetic tests compiling arithmetic operations
func TestCompile_Arithmetic(t *testing.T) {
	ast := &parser.Program{
		HasOuterCircle: true,
		MainEntry: &parser.FunctionDef{
			IsMain: true,
			Body: []parser.Statement{
				&parser.OutputStatement{
					Value: &parser.BinaryOp{
						Left:     &parser.Literal{Value: 1, LiteralType: parser.Integer},
						Operator: parser.Add,
						Right:    &parser.Literal{Value: 2, LiteralType: parser.Integer},
						DataType: parser.Integer,
					},
				},
			},
		},
	}

	code, err := Compile(ast)

	require.NoError(t, err)
	assert.Contains(t, code, "print((1 + 2))")
}

// TestCompileLiteral tests literal compilation
func TestCompileLiteral(t *testing.T) {
	compiler := NewCompiler()

	tests := []struct {
		literal  *parser.Literal
		expected string
	}{
		{&parser.Literal{Value: 42, LiteralType: parser.Integer}, "42"},
		{&parser.Literal{Value: 3.14, LiteralType: parser.Float}, "3.14"},
		{&parser.Literal{Value: "hello", LiteralType: parser.String}, `"hello"`},
		{&parser.Literal{Value: true, LiteralType: parser.Boolean}, "True"},
		{&parser.Literal{Value: false, LiteralType: parser.Boolean}, "False"},
	}

	for _, tc := range tests {
		result := compiler.compileLiteral(tc.literal)
		assert.Equal(t, tc.expected, result)
	}
}

// TestCompileBinaryOp tests binary operation compilation
func TestCompileBinaryOp(t *testing.T) {
	compiler := NewCompiler()

	tests := []struct {
		op       parser.OperatorType
		expected string
	}{
		{parser.Add, "(1 + 2)"},
		{parser.Subtract, "(1 - 2)"},
		{parser.Multiply, "(1 * 2)"},
		{parser.Divide, "(1 / 2)"},
		{parser.Equal, "(1 == 2)"},
		{parser.NotEqual, "(1 != 2)"},
		{parser.LessThan, "(1 < 2)"},
		{parser.GreaterThan, "(1 > 2)"},
		{parser.And, "(1 and 2)"},
		{parser.Or, "(1 or 2)"},
	}

	for _, tc := range tests {
		binOp := &parser.BinaryOp{
			Left:     &parser.Literal{Value: 1, LiteralType: parser.Integer},
			Operator: tc.op,
			Right:    &parser.Literal{Value: 2, LiteralType: parser.Integer},
		}
		result := compiler.compileBinaryOp(binOp)
		assert.Equal(t, tc.expected, result)
	}
}

// TestCompileIfStatement tests if statement compilation
func TestCompileIfStatement(t *testing.T) {
	ast := &parser.Program{
		HasOuterCircle: true,
		MainEntry: &parser.FunctionDef{
			IsMain: true,
			Body: []parser.Statement{
				&parser.IfStatement{
					Condition: &parser.Literal{Value: true, LiteralType: parser.Boolean},
					ThenBranch: []parser.Statement{
						&parser.OutputStatement{
							Value: &parser.Literal{Value: "True", LiteralType: parser.String},
						},
					},
					ElseBranch: []parser.Statement{
						&parser.OutputStatement{
							Value: &parser.Literal{Value: "False", LiteralType: parser.String},
						},
					},
				},
			},
		},
	}

	code, err := Compile(ast)

	require.NoError(t, err)
	assert.Contains(t, code, "if True:")
	assert.Contains(t, code, `print("True")`)
	assert.Contains(t, code, "else:")
	assert.Contains(t, code, `print("False")`)
}

// TestIndentation tests proper indentation
func TestIndentation(t *testing.T) {
	compiler := NewCompiler()

	// Test basic indentation
	compiler.writeLine("def test():")
	compiler.indent++
	compiler.writeLine("pass")
	compiler.indent--

	output := compiler.output.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	assert.Equal(t, "def test():", lines[0])
	assert.Equal(t, "    pass", lines[1])
}

// TestCompile_ErrorHandling tests compiler error handling
func TestCompile_ErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		ast     *parser.Program
		wantErr string
	}{
		{
			name: "No outer circle",
			ast: &parser.Program{
				HasOuterCircle: false,
			},
			wantErr: "外周円が検出されません",
		},
		{
			name: "Empty program",
			ast: &parser.Program{
				HasOuterCircle: true,
				Functions:      []*parser.FunctionDef{},
				Globals:        []parser.Statement{},
			},
			wantErr: "メインエントリーポイントが見つかりません",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, err := Compile(tc.ast)
			assert.Error(t, err)
			assert.Empty(t, code)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestCompile_ComplexProgram tests compiling a complex program
func TestCompile_ComplexProgram(t *testing.T) {
	ast := &parser.Program{
		HasOuterCircle: true,
		MainEntry: &parser.FunctionDef{
			IsMain: true,
			Body: []parser.Statement{
				// Assignment
				&parser.Assignment{
					Target: &parser.Identifier{Name: "x"},
					Value:  &parser.Literal{Value: 5, LiteralType: parser.Integer},
				},
				// For loop
				&parser.ForLoop{
					Counter: &parser.Identifier{Name: "i"},
					Start:   &parser.Literal{Value: 0, LiteralType: parser.Integer},
					End:     &parser.Identifier{Name: "x"},
					Step:    &parser.Literal{Value: 1, LiteralType: parser.Integer},
					Body: []parser.Statement{
						&parser.OutputStatement{
							Value: &parser.Identifier{Name: "i"},
						},
					},
				},
			},
		},
	}

	code, err := Compile(ast)

	require.NoError(t, err)
	assert.Contains(t, code, "x = 5")
	assert.Contains(t, code, "for i in range(0, x, 1):")
	assert.Contains(t, code, "print(i)")
}

// TestCompile_ParallelBlock tests parallel block compilation
func TestCompile_ParallelBlock(t *testing.T) {
	ast := &parser.Program{
		HasOuterCircle: true,
		MainEntry: &parser.FunctionDef{
			IsMain: true,
			Body: []parser.Statement{
				&parser.ParallelBlock{
					Branches: [][]parser.Statement{
						{
							&parser.OutputStatement{
								Value: &parser.Literal{Value: "Branch 1", LiteralType: parser.String},
							},
						},
						{
							&parser.OutputStatement{
								Value: &parser.Literal{Value: "Branch 2", LiteralType: parser.String},
							},
						},
					},
				},
			},
		},
	}

	code, err := Compile(ast)

	require.NoError(t, err)
	assert.Contains(t, code, "import threading")
	assert.Contains(t, code, "threads = []")
	assert.Contains(t, code, "threading.Thread(target=")
	assert.Contains(t, code, "t.start()")
	assert.Contains(t, code, "t.join()")
}

// TestCompile_Functions tests function compilation
func TestCompile_Functions(t *testing.T) {
	ast := &parser.Program{
		HasOuterCircle: true,
		Functions: []*parser.FunctionDef{
			{
				Name: "add",
				Parameters: []*parser.Parameter{
					{Name: "a", DataType: parser.Integer},
					{Name: "b", DataType: parser.Integer},
				},
				Body: []parser.Statement{
					&parser.OutputStatement{
						Value: &parser.BinaryOp{
							Left:     &parser.Identifier{Name: "a"},
							Operator: parser.Add,
							Right:    &parser.Identifier{Name: "b"},
							DataType: parser.Integer,
						},
					},
				},
				ReturnType: parser.Integer,
			},
		},
		MainEntry: &parser.FunctionDef{
			IsMain: true,
			Body: []parser.Statement{
				&parser.OutputStatement{
					Value: &parser.FunctionCall{
						Function: &parser.Identifier{Name: "add"},
						Arguments: []parser.Expression{
							&parser.Literal{Value: 1, LiteralType: parser.Integer},
							&parser.Literal{Value: 2, LiteralType: parser.Integer},
						},
					},
				},
			},
		},
	}

	code, err := Compile(ast)

	require.NoError(t, err)
	assert.Contains(t, code, "def add(a, b):")
	assert.Contains(t, code, "print((a + b))")
	assert.Contains(t, code, "print(add(1, 2))")
}

// TestCompile_WhileLoop tests while loop compilation
func TestCompile_WhileLoop(t *testing.T) {
	ast := &parser.Program{
		HasOuterCircle: true,
		MainEntry: &parser.FunctionDef{
			IsMain: true,
			Body: []parser.Statement{
				&parser.Assignment{
					Target: &parser.Identifier{Name: "count"},
					Value:  &parser.Literal{Value: 0, LiteralType: parser.Integer},
				},
				&parser.WhileLoop{
					Condition: &parser.BinaryOp{
						Left:     &parser.Identifier{Name: "count"},
						Operator: parser.LessThan,
						Right:    &parser.Literal{Value: 5, LiteralType: parser.Integer},
					},
					Body: []parser.Statement{
						&parser.OutputStatement{
							Value: &parser.Identifier{Name: "count"},
						},
						&parser.Assignment{
							Target: &parser.Identifier{Name: "count"},
							Value: &parser.BinaryOp{
								Left:     &parser.Identifier{Name: "count"},
								Operator: parser.Add,
								Right:    &parser.Literal{Value: 1, LiteralType: parser.Integer},
							},
						},
					},
				},
			},
		},
	}

	code, err := Compile(ast)

	require.NoError(t, err)
	assert.Contains(t, code, "count = 0")
	assert.Contains(t, code, "while (count < 5):")
	assert.Contains(t, code, "print(count)")
	assert.Contains(t, code, "count = (count + 1)")
}

// TestCompileReturnStatement tests return statement compilation
func TestCompileReturnStatement(t *testing.T) {
	compiler := NewCompiler()

	tests := []struct {
		name     string
		stmt     *parser.ReturnStatement
		expected string
	}{
		{
			name:     "return with value",
			stmt:     &parser.ReturnStatement{Value: &parser.Literal{Value: 42, LiteralType: parser.Integer}},
			expected: "return 42",
		},
		{
			name:     "return without value",
			stmt:     &parser.ReturnStatement{Value: nil},
			expected: "return",
		},
		{
			name: "return with expression",
			stmt: &parser.ReturnStatement{
				Value: &parser.BinaryOp{
					Left:     &parser.Identifier{Name: "x"},
					Operator: parser.Add,
					Right:    &parser.Literal{Value: 1, LiteralType: parser.Integer},
				},
			},
			expected: "return (x + 1)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compiler.output.Reset()
			compiler.compileReturnStatement(tc.stmt)
			output := strings.TrimSpace(compiler.output.String())
			assert.Equal(t, tc.expected, output)
		})
	}
}

// TestCompileUnaryOp tests unary operation compilation
func TestCompileUnaryOp(t *testing.T) {
	compiler := NewCompiler()

	tests := []struct {
		name     string
		op       *parser.UnaryOp
		expected string
	}{
		{
			name: "not operator",
			op: &parser.UnaryOp{
				Operator: parser.Not,
				Operand:  &parser.Literal{Value: true, LiteralType: parser.Boolean},
			},
			expected: "not True",
		},
		{
			name: "not with identifier",
			op: &parser.UnaryOp{
				Operator: parser.Not,
				Operand:  &parser.Identifier{Name: "flag"},
			},
			expected: "not flag",
		},
		{
			name: "not with expression",
			op: &parser.UnaryOp{
				Operator: parser.Not,
				Operand: &parser.BinaryOp{
					Left:     &parser.Identifier{Name: "x"},
					Operator: parser.Equal,
					Right:    &parser.Literal{Value: 0, LiteralType: parser.Integer},
				},
			},
			expected: "not (x == 0)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := compiler.compileUnaryOp(tc.op)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestCompileArrayLiteral tests array literal compilation
func TestCompileArrayLiteral(t *testing.T) {
	compiler := NewCompiler()

	tests := []struct {
		name     string
		arr      *parser.ArrayLiteral
		expected string
	}{
		{
			name:     "empty array",
			arr:      &parser.ArrayLiteral{Elements: []parser.Expression{}},
			expected: "[]",
		},
		{
			name: "array with integers",
			arr: &parser.ArrayLiteral{
				Elements: []parser.Expression{
					&parser.Literal{Value: 1, LiteralType: parser.Integer},
					&parser.Literal{Value: 2, LiteralType: parser.Integer},
					&parser.Literal{Value: 3, LiteralType: parser.Integer},
				},
			},
			expected: "[1, 2, 3]",
		},
		{
			name: "array with mixed types",
			arr: &parser.ArrayLiteral{
				Elements: []parser.Expression{
					&parser.Literal{Value: 1, LiteralType: parser.Integer},
					&parser.Literal{Value: "hello", LiteralType: parser.String},
					&parser.Literal{Value: true, LiteralType: parser.Boolean},
				},
			},
			expected: "[1, \"hello\", True]",
		},
		{
			name: "nested array",
			arr: &parser.ArrayLiteral{
				Elements: []parser.Expression{
					&parser.ArrayLiteral{
						Elements: []parser.Expression{
							&parser.Literal{Value: 1, LiteralType: parser.Integer},
							&parser.Literal{Value: 2, LiteralType: parser.Integer},
						},
					},
					&parser.ArrayLiteral{
						Elements: []parser.Expression{
							&parser.Literal{Value: 3, LiteralType: parser.Integer},
							&parser.Literal{Value: 4, LiteralType: parser.Integer},
						},
					},
				},
			},
			expected: "[[1, 2], [3, 4]]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := compiler.compileArrayLiteral(tc.arr)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestCompileMapLiteral tests map literal compilation
func TestCompileMapLiteral(t *testing.T) {
	compiler := NewCompiler()

	tests := []struct {
		name     string
		m        *parser.MapLiteral
		expected string
	}{
		{
			name:     "empty map",
			m:        &parser.MapLiteral{Pairs: [][2]parser.Expression{}},
			expected: "{}",
		},
		{
			name: "map with string keys",
			m: &parser.MapLiteral{
				Pairs: [][2]parser.Expression{
					{
						&parser.Literal{Value: "name", LiteralType: parser.String},
						&parser.Literal{Value: "Alice", LiteralType: parser.String},
					},
					{
						&parser.Literal{Value: "age", LiteralType: parser.String},
						&parser.Literal{Value: 30, LiteralType: parser.Integer},
					},
				},
			},
			expected: `{"name": "Alice", "age": 30}`,
		},
		{
			name: "map with integer keys",
			m: &parser.MapLiteral{
				Pairs: [][2]parser.Expression{
					{
						&parser.Literal{Value: 1, LiteralType: parser.Integer},
						&parser.Literal{Value: "one", LiteralType: parser.String},
					},
					{
						&parser.Literal{Value: 2, LiteralType: parser.Integer},
						&parser.Literal{Value: "two", LiteralType: parser.String},
					},
				},
			},
			expected: `{1: "one", 2: "two"}`,
		},
		{
			name: "nested map",
			m: &parser.MapLiteral{
				Pairs: [][2]parser.Expression{
					{
						&parser.Literal{Value: "data", LiteralType: parser.String},
						&parser.MapLiteral{
							Pairs: [][2]parser.Expression{
								{
									&parser.Literal{Value: "x", LiteralType: parser.String},
									&parser.Literal{Value: 10, LiteralType: parser.Integer},
								},
							},
						},
					},
				},
			},
			expected: `{"data": {"x": 10}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := compiler.compileMapLiteral(tc.m)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestCompileExpression_Coverage tests additional expression cases
func TestCompileExpression_Coverage(t *testing.T) {
	compiler := NewCompiler()

	tests := []struct {
		name     string
		expr     parser.Expression
		expected string
	}{
		{
			name:     "nil expression",
			expr:     nil,
			expected: "None",
		},
		{
			name: "function call",
			expr: &parser.FunctionCall{
				Function: &parser.Identifier{Name: "max"},
				Arguments: []parser.Expression{
					&parser.Literal{Value: 5, LiteralType: parser.Integer},
					&parser.Literal{Value: 10, LiteralType: parser.Integer},
				},
			},
			expected: "max(5, 10)",
		},
		{
			name: "array literal in expression",
			expr: &parser.ArrayLiteral{
				Elements: []parser.Expression{
					&parser.Literal{Value: "a", LiteralType: parser.String},
					&parser.Literal{Value: "b", LiteralType: parser.String},
				},
			},
			expected: `["a", "b"]`,
		},
		{
			name: "map literal in expression",
			expr: &parser.MapLiteral{
				Pairs: [][2]parser.Expression{
					{
						&parser.Literal{Value: "key", LiteralType: parser.String},
						&parser.Literal{Value: "value", LiteralType: parser.String},
					},
				},
			},
			expected: `{"key": "value"}`,
		},
		{
			name: "unary not expression",
			expr: &parser.UnaryOp{
				Operator: parser.Not,
				Operand:  &parser.Identifier{Name: "valid"},
			},
			expected: "not valid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := compiler.compileExpression(tc.expr)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestCompileStatement_Coverage tests additional statement cases
func TestCompileStatement_Coverage(t *testing.T) {
	compiler := NewCompiler()

	tests := []struct {
		name        string
		stmt        parser.Statement
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil statement",
			stmt:        nil,
			expectError: true,
			errorMsg:    "Cannot compile nil statement",
		},
		{
			name: "expression statement",
			stmt: &parser.ExpressionStatement{
				Expression: &parser.FunctionCall{
					Function: &parser.Identifier{Name: "process"},
					Arguments: []parser.Expression{
						&parser.Identifier{Name: "data"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "return statement",
			stmt: &parser.ReturnStatement{
				Value: &parser.BinaryOp{
					Left:     &parser.Identifier{Name: "a"},
					Operator: parser.Multiply,
					Right:    &parser.Identifier{Name: "b"},
				},
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compiler.output.Reset()
			err := compiler.compileStatement(tc.stmt)
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCompileBinaryOp_AdditionalOperators tests additional binary operators
func TestCompileBinaryOp_AdditionalOperators(t *testing.T) {
	compiler := NewCompiler()

	tests := []struct {
		name     string
		op       parser.OperatorType
		expected string
	}{
		{"less equal", parser.LessEqual, "(1 <= 2)"},
		{"greater equal", parser.GreaterEqual, "(1 >= 2)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			binOp := &parser.BinaryOp{
				Left:     &parser.Literal{Value: 1, LiteralType: parser.Integer},
				Operator: tc.op,
				Right:    &parser.Literal{Value: 2, LiteralType: parser.Integer},
			}
			result := compiler.compileBinaryOp(binOp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestCompile_WithFunctionReturn tests functions with return statements
func TestCompile_WithFunctionReturn(t *testing.T) {
	ast := &parser.Program{
		HasOuterCircle: true,
		Functions: []*parser.FunctionDef{
			{
				Name: "square",
				Parameters: []*parser.Parameter{
					{Name: "x", DataType: parser.Integer},
				},
				Body: []parser.Statement{
					&parser.ReturnStatement{
						Value: &parser.BinaryOp{
							Left:     &parser.Identifier{Name: "x"},
							Operator: parser.Multiply,
							Right:    &parser.Identifier{Name: "x"},
						},
					},
				},
				ReturnType: parser.Integer,
			},
		},
		MainEntry: &parser.FunctionDef{
			IsMain: true,
			Body: []parser.Statement{
				&parser.OutputStatement{
					Value: &parser.FunctionCall{
						Function: &parser.Identifier{Name: "square"},
						Arguments: []parser.Expression{
							&parser.Literal{Value: 5, LiteralType: parser.Integer},
						},
					},
				},
			},
		},
	}

	code, err := Compile(ast)

	require.NoError(t, err)
	assert.Contains(t, code, "def square(x):")
	assert.Contains(t, code, "return (x * x)")
	assert.Contains(t, code, "print(square(5))")
}
