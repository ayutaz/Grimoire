package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ayutaz/grimoire/internal/compiler"
	"github.com/ayutaz/grimoire/internal/detector"
	"github.com/ayutaz/grimoire/internal/parser"
)

// Execute runs the CLI
func Execute(version, commit, date string) error {
	rootCmd := &cobra.Command{
		Use:   "grimoire",
		Short: "A visual programming language using magic circles",
		Long: `Grimoire is a visual programming language where programs are expressed as magic circles.
Draw your spells and watch them come to life!`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// Run command
	runCmd := &cobra.Command{
		Use:   "run [image]",
		Short: "Run a Grimoire program",
		Args:  cobra.ExactArgs(1),
		RunE:  runCommand,
	}

	// Compile command
	compileCmd := &cobra.Command{
		Use:   "compile [image]",
		Short: "Compile a Grimoire program to Python",
		Args:  cobra.ExactArgs(1),
		RunE:  compileCommand,
	}
	compileCmd.Flags().StringP("output", "o", "", "Output file path")

	// Debug command
	debugCmd := &cobra.Command{
		Use:   "debug [image]",
		Short: "Debug a Grimoire program (show detected symbols)",
		Args:  cobra.ExactArgs(1),
		RunE:  debugCommand,
	}

	rootCmd.AddCommand(runCmd, compileCmd, debugCmd)
	return rootCmd.Execute()
}

func runCommand(cmd *cobra.Command, args []string) error {
	imagePath := args[0]

	// Process the image
	code, err := processImage(imagePath)
	if err != nil {
		return err
	}

	// Execute the generated code
	// TODO: Implement actual Python execution
	fmt.Print(code)
	return nil
}

func compileCommand(cmd *cobra.Command, args []string) error {
	imagePath := args[0]
	outputPath, _ := cmd.Flags().GetString("output")

	// Process the image
	code, err := processImage(imagePath)
	if err != nil {
		return err
	}

	// Output the code
	if outputPath != "" {
		return os.WriteFile(outputPath, []byte(code), 0644)
	}
	fmt.Print(code)
	return nil
}

func debugCommand(cmd *cobra.Command, args []string) error {
	imagePath := args[0]

	// Detect symbols
	symbols, err := detector.DetectSymbols(imagePath)
	if err != nil {
		return fmt.Errorf("failed to detect symbols: %w", err)
	}

	// Display debug information
	fmt.Printf("Detected %d symbols in %s:\n", len(symbols), filepath.Base(imagePath))
	for i, symbol := range symbols {
		fmt.Printf("[%d] %+v\n", i, symbol)
	}

	return nil
}

func processImage(imagePath string) (string, error) {
	// 1. Detect symbols
	symbols, err := detector.DetectSymbols(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to detect symbols: %w", err)
	}

	// 2. Parse to AST
	ast, err := parser.Parse(symbols)
	if err != nil {
		return "", fmt.Errorf("failed to parse: %w", err)
	}

	// 3. Compile to Python
	code, err := compiler.Compile(ast)
	if err != nil {
		return "", fmt.Errorf("failed to compile: %w", err)
	}

	return code, nil
}