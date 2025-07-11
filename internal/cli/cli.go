package cli

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ayutaz/grimoire/internal/compiler"
	"github.com/ayutaz/grimoire/internal/detector"
	grimoireErrors "github.com/ayutaz/grimoire/internal/errors"
	"github.com/ayutaz/grimoire/internal/i18n"
	"github.com/ayutaz/grimoire/internal/parser"
	"github.com/spf13/cobra"
)

// Execute runs the CLI
func Execute(version, commit, date string) error {
	// Initialize i18n before creating commands
	i18n.Init()
	
	rootCmd := &cobra.Command{
		Use:   "grimoire",
		Short: i18n.T("cli.description_short"),
		Long:  i18n.T("cli.description_long"),
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// Run command
	runCmd := &cobra.Command{
		Use:   "run [image]",
		Short: i18n.T("cli.run_description"),
		Args:  cobra.ExactArgs(1),
		RunE:  runCommand,
	}

	// Compile command
	compileCmd := &cobra.Command{
		Use:   "compile [image]",
		Short: i18n.T("cli.compile_description"),
		Args:  cobra.ExactArgs(1),
		RunE:  compileCommand,
	}
	compileCmd.Flags().StringP("output", "o", "", i18n.T("cli.output_flag_description"))

	// Debug command
	debugCmd := &cobra.Command{
		Use:   "debug [image]",
		Short: i18n.T("cli.debug_description"),
		Args:  cobra.ExactArgs(1),
		RunE:  debugCommand,
	}

	// Validate command
	validateCmd := &cobra.Command{
		Use:   "validate [image]",
		Short: i18n.T("cli.validate_description"),
		Args:  cobra.ExactArgs(1),
		RunE:  validateCommand,
	}

	// Format command
	formatCmd := &cobra.Command{
		Use:   "format [image]",
		Short: i18n.T("cli.format_description"),
		Args:  cobra.ExactArgs(1),
		RunE:  formatCommand,
	}
	formatCmd.Flags().StringP("output", "o", "", i18n.T("cli.format_output_flag_description"))

	// Optimize command
	optimizeCmd := &cobra.Command{
		Use:   "optimize [image]",
		Short: i18n.T("cli.optimize_description"),
		Args:  cobra.ExactArgs(1),
		RunE:  optimizeCommand,
	}
	optimizeCmd.Flags().StringP("output", "o", "", i18n.T("cli.optimize_output_flag_description"))

	// Add global language flag
	rootCmd.PersistentFlags().StringP("lang", "l", "", i18n.T("cli.language_flag_description"))
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if lang, _ := cmd.Flags().GetString("lang"); lang != "" {
			switch strings.ToLower(lang) {
			case "ja", "japanese":
				i18n.SetLanguage(i18n.Japanese)
			case "en", "english":
				i18n.SetLanguage(i18n.English)
			}
		}
		return nil
	}
	
	rootCmd.AddCommand(runCmd, compileCmd, debugCmd, validateCmd, formatCmd, optimizeCmd)
	return rootCmd.Execute()
}

func runCommand(_ *cobra.Command, args []string) error {
	imagePath := args[0]

	// Process the image
	code, err := processImage(imagePath)
	if err != nil {
		return formatError(err, imagePath)
	}

	// Execute the generated code
	if err := executePython(code); err != nil {
		return grimoireErrors.NewError(grimoireErrors.ExecutionError, i18n.T("msg.failed_execute_python")).
			WithInnerError(err).
			WithSuggestion(i18n.T("suggest.check_python_installed"))
	}
	return nil
}

func compileCommand(cmd *cobra.Command, args []string) error {
	imagePath := args[0]
	outputPath, _ := cmd.Flags().GetString("output")

	// Process the image
	code, err := processImage(imagePath)
	if err != nil {
		return formatError(err, imagePath)
	}

	// Output the code
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(code), 0o644); err != nil {
			return grimoireErrors.NewError(grimoireErrors.FileWriteError, i18n.T("msg.failed_write_output")).
				WithInnerError(err).
				WithLocation(outputPath, 0, 0)
		}
		fmt.Printf(i18n.T("cli.compile_success"), outputPath)
	} else {
		fmt.Print(code)
	}
	return nil
}

func debugCommand(_ *cobra.Command, args []string) error {
	imagePath := args[0]

	// Detect symbols
	symbols, connections, err := detector.DetectSymbols(imagePath)
	if err != nil {
		return formatError(err, imagePath)
	}

	// Display debug information
	fmt.Printf(i18n.T("debug.header"), filepath.Base(imagePath))
	fmt.Printf(i18n.T("debug.detected_summary"), len(symbols), len(connections))

	fmt.Println(i18n.T("debug.symbols_header"))
	for i, symbol := range symbols {
		fmt.Printf(i18n.T("debug.symbol_info"),
			i, symbol.Type, symbol.Position.X, symbol.Position.Y, symbol.Size, symbol.Pattern)
	}

	if len(connections) > 0 {
		fmt.Println(i18n.T("debug.connections_header"))
		for i, conn := range connections {
			fmt.Printf(i18n.T("debug.connection_info"), i, conn.From.Type, conn.To.Type, conn.ConnectionType)
		}
	}

	return nil
}

func processImage(imagePath string) (string, error) {
	// 1. Detect symbols
	symbols, connections, err := detector.DetectSymbols(imagePath)
	if err != nil {
		return "", err // Already formatted error
	}

	// 2. Parse to AST
	ast, err := parser.Parse(symbols, connections)
	if err != nil {
		return "", err // Already formatted error
	}

	// 3. Compile to Python
	code, err := compiler.Compile(ast)
	if err != nil {
		return "", err // Already formatted error
	}

	return code, nil
}

func executePython(code string) error {
	// Create a temporary Python file
	tmpFile, err := os.CreateTemp("", "grimoire_*.py")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	// Write the code
	if _, err := tmpFile.WriteString(code); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// Execute the Python code
	cmd := exec.Command("python3", tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// formatError formats an error for user-friendly display
func formatError(err error, imagePath string) error {
	if grimoireErrors.IsGrimoireError(err) {
		// Already formatted
		return err
	}

	// Wrap generic errors
	if strings.Contains(err.Error(), "no such file") {
		return grimoireErrors.FileNotFoundError(imagePath)
	}

	return grimoireErrors.NewError(grimoireErrors.ExecutionError, i18n.T("msg.error_occurred")).
		WithInnerError(err).
		WithLocation(imagePath, 0, 0)
}

// validateCommand validates a magic circle image
func validateCommand(_ *cobra.Command, args []string) error {
	imagePath := args[0]

	// Detect symbols
	symbols, connections, err := detector.DetectSymbols(imagePath)
	if err != nil {
		return formatError(err, imagePath)
	}

	// Validation checks
	var issues []string

	// Check for outer circle
	hasOuterCircle := false
	for _, s := range symbols {
		if s.Type == detector.OuterCircle {
			hasOuterCircle = true
			break
		}
	}
	if !hasOuterCircle {
		issues = append(issues, i18n.T("validate.no_outer_circle"))
	}

	// Check for main function
	hasMain := false
	for _, s := range symbols {
		if s.Type == detector.DoubleCircle {
			hasMain = true
			break
		}
	}
	if !hasMain {
		issues = append(issues, i18n.T("validate.no_main_function"))
	}

	// Check for orphaned symbols
	connected := make(map[*detector.Symbol]bool)
	for _, c := range connections {
		connected[c.From] = true
		connected[c.To] = true
	}
	for _, s := range symbols {
		if s.Type != detector.OuterCircle && !connected[s] {
			issues = append(issues, fmt.Sprintf(i18n.T("validate.orphaned_symbol"), s.Type, s.Position.X, s.Position.Y))
		}
	}

	// Output results
	if len(issues) == 0 {
		fmt.Println(i18n.T("validate.success"))
		fmt.Printf(i18n.T("validate.symbols_found"), len(symbols))
		fmt.Printf(i18n.T("validate.connections_found"), len(connections))
	} else {
		fmt.Println(i18n.T("validate.issues_found"))
		for i, issue := range issues {
			fmt.Printf("%d. %s\n", i+1, issue)
		}
		return grimoireErrors.NewError(grimoireErrors.ValidationError, i18n.T("validate.failed"))
	}

	return nil
}

// formatCommand formats a magic circle image
func formatCommand(cmd *cobra.Command, args []string) error {
	imagePath := args[0]
	outputPath, _ := cmd.Flags().GetString("output")

	// For now, format command will analyze and provide suggestions
	// In a full implementation, this would create a cleaned-up version of the magic circle

	// Detect symbols
	symbols, connections, err := detector.DetectSymbols(imagePath)
	if err != nil {
		return formatError(err, imagePath)
	}

	fmt.Println(i18n.T("format.analyzing"))

	// Analyze and provide formatting suggestions
	var suggestions []string

	// Check symbol alignment
	for i, s1 := range symbols {
		for j, s2 := range symbols {
			if i >= j || s1.Type == detector.OuterCircle || s2.Type == detector.OuterCircle {
				continue
			}
			// Check if symbols are nearly aligned but not quite
			dx := math.Abs(float64(s1.Position.X - s2.Position.X))
			dy := math.Abs(float64(s1.Position.Y - s2.Position.Y))
			if (dx > 0 && dx < 10) || (dy > 0 && dy < 10) {
				suggestions = append(suggestions, fmt.Sprintf(i18n.T("format.align_symbols"), s1.Type, s2.Type))
			}
		}
	}

	// Check connection angles
	for _, c := range connections {
		angle := math.Atan2(float64(c.To.Position.Y-c.From.Position.Y), float64(c.To.Position.X-c.From.Position.X))
		angleDeg := angle * 180 / math.Pi
		// Check if angle is close to but not exactly 0, 45, 90, 135, 180, -45, -90, -135
		standardAngles := []float64{0, 45, 90, 135, 180, -45, -90, -135}
		for _, std := range standardAngles {
			if diff := math.Abs(angleDeg - std); diff > 0 && diff < 5 {
				suggestions = append(suggestions, fmt.Sprintf(i18n.T("format.straighten_connection"), c.From.Type, c.To.Type))
				break
			}
		}
	}

	if len(suggestions) == 0 {
		fmt.Println(i18n.T("format.well_formatted"))
	} else {
		fmt.Println(i18n.T("format.suggestions"))
		for i, s := range suggestions {
			fmt.Printf("%d. %s\n", i+1, s)
		}
	}

	if outputPath != "" {
		fmt.Printf(i18n.T("format.output_note"), outputPath)
	}

	return nil
}

// optimizeCommand optimizes a magic circle program
func optimizeCommand(cmd *cobra.Command, args []string) error {
	imagePath := args[0]
	outputPath, _ := cmd.Flags().GetString("output")

	// Process the image
	code, err := processImage(imagePath)
	if err != nil {
		return formatError(err, imagePath)
	}

	// Parse to get AST
	symbols, connections, err := detector.DetectSymbols(imagePath)
	if err != nil {
		return formatError(err, imagePath)
	}

	ast, err := parser.Parse(symbols, connections)
	if err != nil {
		return formatError(err, imagePath)
	}

	fmt.Println(i18n.T("optimize.analyzing"))

	// Optimization analysis
	var optimizations []string

	// Check for redundant operations
	{
		// Check for unused variables
		defined := make(map[string]bool)
		used := make(map[string]bool)
		
		// Analyze global statements
		for _, stmt := range ast.Globals {
			analyzeDefined(stmt, defined)
			analyzeUsed(stmt, used)
		}
		
		// Analyze functions
		for _, fn := range ast.Functions {
			for _, stmt := range fn.Body {
				analyzeDefined(stmt, defined)
				analyzeUsed(stmt, used)
			}
		}
		
		// Analyze main entry
		if ast.MainEntry != nil {
			for _, stmt := range ast.MainEntry.Body {
				analyzeDefined(stmt, defined)
				analyzeUsed(stmt, used)
			}
		}
		
		for var_ := range defined {
			if !used[var_] {
				optimizations = append(optimizations, fmt.Sprintf(i18n.T("optimize.unused_variable"), var_))
			}
		}

		// Check for duplicate operations in globals
		for i, stmt1 := range ast.Globals {
			for j, stmt2 := range ast.Globals {
				if i < j && statementsEqual(stmt1, stmt2) {
					optimizations = append(optimizations, i18n.T("optimize.duplicate_operation"))
				}
			}
		}
	}

	if len(optimizations) == 0 {
		fmt.Println(i18n.T("optimize.well_optimized"))
	} else {
		fmt.Println(i18n.T("optimize.suggestions"))
		for i, opt := range optimizations {
			fmt.Printf("%d. %s\n", i+1, opt)
		}
	}

	// Output optimized code if requested
	if outputPath != "" {
		if outputPath == "-" {
			fmt.Println("\n" + i18n.T("optimize.optimized_code"))
			fmt.Println(code)
		} else {
			if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
				return grimoireErrors.NewError(grimoireErrors.IOError, i18n.T("msg.failed_write_file")).
					WithInnerError(err).
					WithLocation(outputPath, 0, 0)
			}
			fmt.Printf(i18n.T("optimize.saved_to"), outputPath)
		}
	}

	return nil
}

// Helper functions for optimization analysis
func analyzeDefined(stmt parser.Statement, defined map[string]bool) {
	switch s := stmt.(type) {
	case *parser.Assignment:
		if id := s.Target; id != nil {
			defined[id.Name] = true
		}
	case *parser.ForLoop:
		if id := s.Counter; id != nil {
			defined[id.Name] = true
		}
		for _, innerStmt := range s.Body {
			analyzeDefined(innerStmt, defined)
		}
	case *parser.IfStatement:
		for _, innerStmt := range s.ThenBranch {
			analyzeDefined(innerStmt, defined)
		}
		for _, innerStmt := range s.ElseBranch {
			analyzeDefined(innerStmt, defined)
		}
	case *parser.WhileLoop:
		for _, innerStmt := range s.Body {
			analyzeDefined(innerStmt, defined)
		}
	case *parser.ParallelBlock:
		for _, branch := range s.Branches {
			for _, innerStmt := range branch {
				analyzeDefined(innerStmt, defined)
			}
		}
	}
}

func analyzeUsed(stmt parser.Statement, used map[string]bool) {
	switch s := stmt.(type) {
	case *parser.Assignment:
		analyzeUsedExpr(s.Value, used)
	case *parser.OutputStatement:
		analyzeUsedExpr(s.Value, used)
	case *parser.IfStatement:
		analyzeUsedExpr(s.Condition, used)
		for _, innerStmt := range s.ThenBranch {
			analyzeUsed(innerStmt, used)
		}
		for _, innerStmt := range s.ElseBranch {
			analyzeUsed(innerStmt, used)
		}
	case *parser.ForLoop:
		analyzeUsedExpr(s.Start, used)
		analyzeUsedExpr(s.End, used)
		for _, innerStmt := range s.Body {
			analyzeUsed(innerStmt, used)
		}
	case *parser.WhileLoop:
		analyzeUsedExpr(s.Condition, used)
		for _, innerStmt := range s.Body {
			analyzeUsed(innerStmt, used)
		}
	case *parser.ParallelBlock:
		for _, branch := range s.Branches {
			for _, innerStmt := range branch {
				analyzeUsed(innerStmt, used)
			}
		}
	}
}

func analyzeUsedExpr(expr parser.Expression, used map[string]bool) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *parser.Identifier:
		used[e.Name] = true
	case *parser.BinaryOp:
		analyzeUsedExpr(e.Left, used)
		analyzeUsedExpr(e.Right, used)
	case *parser.UnaryOp:
		analyzeUsedExpr(e.Operand, used)
	case *parser.FunctionCall:
		for _, arg := range e.Arguments {
			analyzeUsedExpr(arg, used)
		}
	case *parser.ArrayLiteral:
		for _, elem := range e.Elements {
			analyzeUsedExpr(elem, used)
		}
	}
}

func statementsEqual(s1, s2 parser.Statement) bool {
	// Simple equality check - in a real implementation this would be more sophisticated
	if s1 == nil || s2 == nil {
		return s1 == s2
	}
	return fmt.Sprintf("%T", s1) == fmt.Sprintf("%T", s2)
}
