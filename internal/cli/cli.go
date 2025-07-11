package cli

import (
	"fmt"
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
	
	rootCmd.AddCommand(runCmd, compileCmd, debugCmd)
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
