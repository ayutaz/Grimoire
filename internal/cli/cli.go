package cli

import (
	"fmt"
	"os"
	"os/exec"
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
	return executePython(code)
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
	symbols, connections, err := detector.DetectSymbols(imagePath)
	if err != nil {
		return fmt.Errorf("failed to detect symbols: %w", err)
	}

	// Display debug information
	fmt.Printf("Detected %d symbols and %d connections in %s:\n", len(symbols), len(connections), filepath.Base(imagePath))
	for i, symbol := range symbols {
		fmt.Printf("[%d] %+v\n", i, symbol)
	}
	if len(connections) > 0 {
		fmt.Println("\nConnections:")
		for i, conn := range connections {
			fmt.Printf("[%d] %s -> %s (%s)\n", i, conn.From.Type, conn.To.Type, conn.ConnectionType)
		}
	}

	return nil
}

func processImage(imagePath string) (string, error) {
	// 1. Detect symbols
	symbols, connections, err := detector.DetectSymbols(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to detect symbols: %w", err)
	}

	// 2. Parse to AST
	ast, err := parser.Parse(symbols, connections)
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

func executePython(code string) error {
	// Create a temporary Python file
	tmpFile, err := os.CreateTemp("", "grimoire_*.py")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the code
	if _, err := tmpFile.WriteString(code); err != nil {
		return fmt.Errorf("failed to write code: %w", err)
	}
	tmpFile.Close()

	// Execute the Python code
	cmd := exec.Command("python3", tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}