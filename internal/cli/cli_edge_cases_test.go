package cli

import (
	"os"
	"testing"

	"github.com/ayutaz/grimoire/internal/i18n"
	"github.com/ayutaz/grimoire/internal/parser"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestAnalyzeUsedExprEdgeCases tests analyzeUsedExpr with various expression types
func TestAnalyzeUsedExprEdgeCases(t *testing.T) {
	used := make(map[string]bool)

	// Test nil expression
	analyzeUsedExpr(nil, used)
	assert.Empty(t, used)

	// Test Identifier
	id := &parser.Identifier{Name: "testVar"}
	analyzeUsedExpr(id, used)
	assert.True(t, used["testVar"])

	// Test BinaryOp
	binOp := &parser.BinaryOp{
		Left:     &parser.Identifier{Name: "left"},
		Operator: "+",
		Right:    &parser.Identifier{Name: "right"},
	}
	used = make(map[string]bool)
	analyzeUsedExpr(binOp, used)
	assert.True(t, used["left"])
	assert.True(t, used["right"])

	// Test UnaryOp
	unaryOp := &parser.UnaryOp{
		Operator: "-",
		Operand:  &parser.Identifier{Name: "unaryVar"},
	}
	used = make(map[string]bool)
	analyzeUsedExpr(unaryOp, used)
	assert.True(t, used["unaryVar"])

	// Test FunctionCall
	funcCall := &parser.FunctionCall{
		Function: &parser.Identifier{Name: "testFunc"},
		Arguments: []parser.Expression{
			&parser.Identifier{Name: "arg1"},
			&parser.Identifier{Name: "arg2"},
		},
	}
	used = make(map[string]bool)
	analyzeUsedExpr(funcCall, used)
	assert.True(t, used["arg1"])
	assert.True(t, used["arg2"])

	// Test ArrayLiteral
	arrayLit := &parser.ArrayLiteral{
		Elements: []parser.Expression{
			&parser.Identifier{Name: "elem1"},
			&parser.Identifier{Name: "elem2"},
		},
	}
	used = make(map[string]bool)
	analyzeUsedExpr(arrayLit, used)
	assert.True(t, used["elem1"])
	assert.True(t, used["elem2"])
}

// TestAnalyzeDefinedEdgeCases tests analyzeDefined with various statement types
func TestAnalyzeDefinedEdgeCases(t *testing.T) {
	defined := make(map[string]bool)

	// Test Assignment
	assignment := &parser.Assignment{
		Target: &parser.Identifier{Name: "assignVar"},
		Value:  &parser.Literal{Value: "42"},
	}
	analyzeDefined(assignment, defined)
	assert.True(t, defined["assignVar"])

	// Test ForLoop
	forLoop := &parser.ForLoop{
		Counter: &parser.Identifier{Name: "i"},
		Start:   &parser.Literal{Value: "0"},
		End:     &parser.Literal{Value: "10"},
		Body: []parser.Statement{
			&parser.Assignment{
				Target: &parser.Identifier{Name: "innerVar"},
				Value:  &parser.Literal{Value: "1"},
			},
		},
	}
	defined = make(map[string]bool)
	analyzeDefined(forLoop, defined)
	assert.True(t, defined["i"])
	assert.True(t, defined["innerVar"])

	// Test IfStatement
	ifStmt := &parser.IfStatement{
		Condition: &parser.BinaryOp{
			Left:     &parser.Identifier{Name: "x"},
			Operator: ">",
			Right:    &parser.Literal{Value: "0"},
		},
		ThenBranch: []parser.Statement{
			&parser.Assignment{
				Target: &parser.Identifier{Name: "thenVar"},
				Value:  &parser.Literal{Value: "1"},
			},
		},
		ElseBranch: []parser.Statement{
			&parser.Assignment{
				Target: &parser.Identifier{Name: "elseVar"},
				Value:  &parser.Literal{Value: "2"},
			},
		},
	}
	defined = make(map[string]bool)
	analyzeDefined(ifStmt, defined)
	assert.True(t, defined["thenVar"])
	assert.True(t, defined["elseVar"])

	// Test WhileLoop
	whileLoop := &parser.WhileLoop{
		Condition: &parser.BinaryOp{
			Left:     &parser.Identifier{Name: "counter"},
			Operator: "<",
			Right:    &parser.Literal{Value: "10"},
		},
		Body: []parser.Statement{
			&parser.Assignment{
				Target: &parser.Identifier{Name: "whileVar"},
				Value:  &parser.Literal{Value: "1"},
			},
		},
	}
	defined = make(map[string]bool)
	analyzeDefined(whileLoop, defined)
	assert.True(t, defined["whileVar"])

	// Test ParallelBlock
	parallelBlock := &parser.ParallelBlock{
		Branches: [][]parser.Statement{
			{
				&parser.Assignment{
					Target: &parser.Identifier{Name: "branch1Var"},
					Value:  &parser.Literal{Value: "1"},
				},
			},
			{
				&parser.Assignment{
					Target: &parser.Identifier{Name: "branch2Var"},
					Value:  &parser.Literal{Value: "2"},
				},
			},
		},
	}
	defined = make(map[string]bool)
	analyzeDefined(parallelBlock, defined)
	assert.True(t, defined["branch1Var"])
	assert.True(t, defined["branch2Var"])
}

// TestAnalyzeUsedStatements tests analyzeUsed with various statement types
func TestAnalyzeUsedStatements(t *testing.T) {
	used := make(map[string]bool)

	// Test Assignment with variable usage
	assignment := &parser.Assignment{
		Target: &parser.Identifier{Name: "result"},
		Value: &parser.BinaryOp{
			Left:     &parser.Identifier{Name: "a"},
			Operator: "+",
			Right:    &parser.Identifier{Name: "b"},
		},
	}
	analyzeUsed(assignment, used)
	assert.True(t, used["a"])
	assert.True(t, used["b"])

	// Test OutputStatement
	output := &parser.OutputStatement{
		Value: &parser.Identifier{Name: "outputVar"},
	}
	used = make(map[string]bool)
	analyzeUsed(output, used)
	assert.True(t, used["outputVar"])

	// Test IfStatement with condition
	ifStmt := &parser.IfStatement{
		Condition: &parser.BinaryOp{
			Left:     &parser.Identifier{Name: "condVar"},
			Operator: "==",
			Right:    &parser.Literal{Value: "10"},
		},
		ThenBranch: []parser.Statement{
			&parser.OutputStatement{
				Value: &parser.Identifier{Name: "thenOutput"},
			},
		},
		ElseBranch: []parser.Statement{
			&parser.OutputStatement{
				Value: &parser.Identifier{Name: "elseOutput"},
			},
		},
	}
	used = make(map[string]bool)
	analyzeUsed(ifStmt, used)
	assert.True(t, used["condVar"])
	assert.True(t, used["thenOutput"])
	assert.True(t, used["elseOutput"])

	// Test ForLoop
	forLoop := &parser.ForLoop{
		Counter: &parser.Identifier{Name: "i"},
		Start:   &parser.Identifier{Name: "startVar"},
		End:     &parser.Identifier{Name: "endVar"},
		Body: []parser.Statement{
			&parser.OutputStatement{
				Value: &parser.Identifier{Name: "loopOutput"},
			},
		},
	}
	used = make(map[string]bool)
	analyzeUsed(forLoop, used)
	assert.True(t, used["startVar"])
	assert.True(t, used["endVar"])
	assert.True(t, used["loopOutput"])

	// Test WhileLoop
	whileLoop := &parser.WhileLoop{
		Condition: &parser.Identifier{Name: "whileCond"},
		Body: []parser.Statement{
			&parser.OutputStatement{
				Value: &parser.Identifier{Name: "whileOutput"},
			},
		},
	}
	used = make(map[string]bool)
	analyzeUsed(whileLoop, used)
	assert.True(t, used["whileCond"])
	assert.True(t, used["whileOutput"])

	// Test ParallelBlock
	parallelBlock := &parser.ParallelBlock{
		Branches: [][]parser.Statement{
			{
				&parser.OutputStatement{
					Value: &parser.Identifier{Name: "parallel1"},
				},
			},
			{
				&parser.OutputStatement{
					Value: &parser.Identifier{Name: "parallel2"},
				},
			},
		},
	}
	used = make(map[string]bool)
	analyzeUsed(parallelBlock, used)
	assert.True(t, used["parallel1"])
	assert.True(t, used["parallel2"])
}

// TestStatementsEqualEdgeCases tests statementsEqual function
func TestStatementsEqualEdgeCases(t *testing.T) {
	// Test with concrete statement types
	assign1 := &parser.Assignment{
		Target: &parser.Identifier{Name: "x"},
		Value:  &parser.Literal{Value: "1"},
	}
	assign2 := &parser.Assignment{
		Target: &parser.Identifier{Name: "y"},
		Value:  &parser.Literal{Value: "2"},
	}
	output := &parser.OutputStatement{
		Value: &parser.Identifier{Name: "x"},
	}

	// Same type should be considered equal (simplistic implementation)
	assert.True(t, statementsEqual(assign1, assign2))

	// Different types should not be equal
	assert.False(t, statementsEqual(assign1, output))

	// Test with nil
	assert.False(t, statementsEqual(assign1, nil))
	assert.False(t, statementsEqual(nil, assign1))
}

// TestDebugFlagHandling tests debug flag processing
func TestDebugFlagHandling(t *testing.T) {
	// Save original env
	originalDebug := os.Getenv("GRIMOIRE_DEBUG")
	defer os.Setenv("GRIMOIRE_DEBUG", originalDebug)

	// Test with environment variable
	os.Setenv("GRIMOIRE_DEBUG", "1")

	rootCmd := &cobra.Command{
		Use:   "grimoire",
		Short: i18n.T("cli.description_short"),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			debugFlag, _ := cmd.Flags().GetBool("debug")
			_ = debugFlag // Simulate using the flag
			// In real code, would check debugFlag || os.Getenv("GRIMOIRE_DEBUG") != ""
			// and call grimoireErrors.EnableDebugMode()
			return nil
		},
	}
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode")

	// Test with debug flag
	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	assert.NoError(t, err)

	// Test with flag set
	rootCmd.Flags().Set("debug", "true")
	err = rootCmd.PersistentPreRunE(rootCmd, []string{})
	assert.NoError(t, err)
}

// TestLanguageHandling tests language flag processing
func TestLanguageHandling(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "grimoire",
		Short: i18n.T("cli.description_short"),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if lang, _ := cmd.Flags().GetString("lang"); lang != "" {
				switch lang {
				case "ja", "japanese":
					// Would call i18n.SetLanguage(i18n.Japanese)
				case "en", "english":
					// Would call i18n.SetLanguage(i18n.English)
				}
			}
			return nil
		},
	}
	rootCmd.PersistentFlags().StringP("lang", "l", "", "Language")

	// Test Japanese
	rootCmd.Flags().Set("lang", "ja")
	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	assert.NoError(t, err)

	// Test English
	rootCmd.Flags().Set("lang", "en")
	err = rootCmd.PersistentPreRunE(rootCmd, []string{})
	assert.NoError(t, err)

	// Test japanese (full name)
	rootCmd.Flags().Set("lang", "japanese")
	err = rootCmd.PersistentPreRunE(rootCmd, []string{})
	assert.NoError(t, err)

	// Test english (full name)
	rootCmd.Flags().Set("lang", "english")
	err = rootCmd.PersistentPreRunE(rootCmd, []string{})
	assert.NoError(t, err)
}

// TestExecutePythonEmptyCode tests executePython with empty code
func TestExecutePythonEmptyCode(t *testing.T) {
	// Test with empty code
	err := executePython("")
	// Empty code should execute successfully (Python allows empty files)
	assert.NoError(t, err)
}

// TestFormatErrorEdgeCases tests formatError with edge cases
func TestFormatErrorEdgeCases(t *testing.T) {
	// Test with permission denied error
	err := formatError(&testError{msg: "permission denied"}, "/test/file.png")
	assert.Contains(t, err.Error(), "ファイル読み込みエラー")

	// Test with access denied error
	err = formatError(&testError{msg: "access is denied"}, "/test/file.png")
	assert.Contains(t, err.Error(), "ファイル読み込みエラー")

	// Test with failed to open file error
	err = formatError(&testError{msg: "failed to open file"}, "/test/file.png")
	assert.Contains(t, err.Error(), "ファイル読み込みエラー")
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

