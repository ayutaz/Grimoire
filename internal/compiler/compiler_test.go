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
	assert.Contains(t, err.Error(), "No main entry point found")
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
			wantErr: "outer circle",
		},
		{
			name: "Empty program",
			ast: &parser.Program{
				HasOuterCircle: true,
				Functions:      []*parser.FunctionDef{},
				Globals:        []parser.Statement{},
			},
			wantErr: "No main entry point",
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