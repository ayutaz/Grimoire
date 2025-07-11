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
	"github.com/ayutaz/grimoire/internal/parser"
	"github.com/spf13/cobra"
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

func runCommand(_ *cobra.Command, args []string) error {
	imagePath := args[0]

	// Process the image
	code, err := processImage(imagePath)
	if err != nil {
		return formatError(err, imagePath)
	}

	// Execute the generated code
	if err := executePython(code); err != nil {
		return grimoireErrors.NewError(grimoireErrors.ExecutionError, "Failed to execute generated Python code").
			WithInnerError(err).
			WithSuggestion("Check that Python 3 is installed and in your PATH")
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
			return grimoireErrors.NewError(grimoireErrors.FileWriteError, "Failed to write output file").
				WithInnerError(err).
				WithLocation(outputPath, 0, 0)
		}
		fmt.Printf("Successfully compiled to %s\n", outputPath)
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
	fmt.Printf("\n=== Debug Information for %s ===\n", filepath.Base(imagePath))
	fmt.Printf("Detected %d symbols and %d connections\n\n", len(symbols), len(connections))

	fmt.Println("Symbols:")
	for i, symbol := range symbols {
		fmt.Printf("  [%d] Type: %-15s Position: (%.0f, %.0f) Size: %.1f Pattern: %s\n",
			i, symbol.Type, symbol.Position.X, symbol.Position.Y, symbol.Size, symbol.Pattern)
	}

	if len(connections) > 0 {
		fmt.Println("\nConnections:")
		for i, conn := range connections {
			fmt.Printf("  [%d] %s -> %s (%s)\n", i, conn.From.Type, conn.To.Type, conn.ConnectionType)
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

	return grimoireErrors.NewError(grimoireErrors.ExecutionError, "An error occurred").
		WithInnerError(err).
		WithLocation(imagePath, 0, 0)
}
