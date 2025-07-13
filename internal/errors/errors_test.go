package errors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ayutaz/grimoire/internal/i18n"
)

func init() {
	// Initialize i18n for tests
	i18n.Init()
}

func TestEnhancedError(t *testing.T) {
	tests := []struct {
		name           string
		setupError     func() error
		expectCode     ErrorCode
		expectInOutput []string
		debugMode      bool
	}{
		{
			name: "basic enhanced error",
			setupError: func() error {
				err := NewError(MissingMainEntry, "Main entry point not found")
				return NewEnhancedError(err)
			},
			expectCode: ErrCodeMissingMainEntry,
			expectInOutput: []string{
				"E3003",
				"Main entry point not found",
			},
			debugMode: false,
		},
		{
			name: "error with context",
			setupError: func() error {
				err := NewError(FileNotFound, "File not found")
				enhanced := NewEnhancedError(err)
				enhanced.WithContext("file", "test.png")
				enhanced.WithContext("operation", "read")
				return enhanced
			},
			expectCode:     ErrCodeFileNotFound,
			expectInOutput: []string{"E1001", "File not found"},
			debugMode:      false,
		},
		{
			name: "error with stack trace in debug mode",
			setupError: func() error {
				err := NewError(SyntaxError, "Syntax error in expression")
				return NewEnhancedError(err)
			},
			expectCode: ErrCodeSyntaxError,
			expectInOutput: []string{
				"E3001",
				"Syntax error",
				"Stack Trace:",
			},
			debugMode: true,
		},
		{
			name: "error with automatic hints",
			setupError: func() error {
				return NewError(MissingMainEntry, "No main entry point")
			},
			expectCode: "",
			expectInOutput: []string{
				"ダブルサークル",
				"DoubleCircle",
			},
			debugMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.debugMode {
				EnableDebugMode()
				defer DisableDebugMode()
			}

			err := tt.setupError()
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			// Check error code if it's an enhanced error
			if enhanced, ok := err.(*EnhancedError); ok && tt.expectCode != "" {
				if enhanced.GetCode() != tt.expectCode {
					t.Errorf("Expected error code %s, got %s", tt.expectCode, enhanced.GetCode())
				}
			}

			// Check error message contains expected strings
			errStr := err.Error()
			for _, expected := range tt.expectInOutput {
				if !strings.Contains(errStr, expected) {
					t.Errorf("Expected error to contain '%s', but it didn't.\nError: %s", expected, errStr)
				}
			}
		})
	}
}

func TestErrorWithHint(t *testing.T) {
	tests := []struct {
		errorType     ErrorType
		expectHint    bool
		expectDetails bool
	}{
		{MissingMainEntry, true, true},
		{NoOuterCircle, true, true},
		{UnbalancedExpression, true, false},
		{FileNotFound, true, false},
		{ImageProcessingError, true, true},
		{CompilationError, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			err := NewError(tt.errorType, "Test error")

			if tt.expectHint && err.Suggestion == "" {
				t.Errorf("Expected hint for error type %s, but got none", tt.errorType)
			}

			if tt.expectDetails && err.Details == "" {
				t.Errorf("Expected details for error type %s, but got none", tt.errorType)
			}
		})
	}
}

func TestDebugMode(t *testing.T) {
	// Test initial state
	if IsDebugMode() {
		t.Error("Debug mode should be disabled by default")
	}

	// Test enabling
	EnableDebugMode()
	if !IsDebugMode() {
		t.Error("Debug mode should be enabled after EnableDebugMode()")
	}

	// Test disabling
	DisableDebugMode()
	if IsDebugMode() {
		t.Error("Debug mode should be disabled after DisableDebugMode()")
	}
}

func TestStackTraceCapture(t *testing.T) {
	EnableDebugMode()
	defer DisableDebugMode()

	err := NewError(ExecutionError, "Test execution error")
	enhanced := NewEnhancedError(err)

	if len(enhanced.GetStackTrace()) == 0 {
		t.Error("Expected stack trace to be captured in debug mode")
	}

	// Check that stack frames have required fields
	for i, frame := range enhanced.GetStackTrace() {
		if frame.Function == "" {
			t.Errorf("Stack frame %d has empty function name", i)
		}
		if frame.File == "" {
			t.Errorf("Stack frame %d has empty file name", i)
		}
		if frame.Line == 0 {
			t.Errorf("Stack frame %d has zero line number", i)
		}
	}
}

func TestErrorContext(t *testing.T) {
	ctx := NewErrorContext().
		WithOperation("compile").
		WithInputFile("test.png").
		WithOutputFile("test.py").
		WithStage("parsing").
		WithMetadata("symbols", 42)

	if ctx.Operation != "compile" {
		t.Errorf("Expected operation 'compile', got '%s'", ctx.Operation)
	}
	if ctx.InputFile != "test.png" {
		t.Errorf("Expected input file 'test.png', got '%s'", ctx.InputFile)
	}
	if ctx.OutputFile != "test.py" {
		t.Errorf("Expected output file 'test.py', got '%s'", ctx.OutputFile)
	}
	if ctx.Stage != "parsing" {
		t.Errorf("Expected stage 'parsing', got '%s'", ctx.Stage)
	}
	if ctx.Metadata["symbols"] != 42 {
		t.Errorf("Expected metadata symbols=42, got %v", ctx.Metadata["symbols"])
	}
}

func TestSuggestSimilar(t *testing.T) {
	validOptions := []string{"circle", "square", "triangle", "pentagon"}

	tests := []struct {
		input            string
		expectSuggestion bool
	}{
		{"circ", true},   // Partial match
		{"CIRCLE", true}, // Case insensitive
		{"squar", true},  // Partial match
		{"xyz", false},   // No match
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			suggestion := SuggestSimilar(tt.input, validOptions)
			if tt.expectSuggestion && suggestion == "" {
				t.Errorf("Expected suggestion for '%s', got none", tt.input)
			}
			if !tt.expectSuggestion && suggestion != "" {
				t.Errorf("Expected no suggestion for '%s', got '%s'", tt.input, suggestion)
			}
		})
	}
}

func TestFormatErrorLocation(t *testing.T) {
	tests := []struct {
		fileName string
		line     int
		column   int
		expected string
	}{
		{"test.go", 10, 5, "test.go:10:5"},
		{"test.go", 10, 0, "test.go:10"},
		{"test.go", 0, 0, "test.go"},
		{"/absolute/path/test.go", 5, 3, "/absolute/path/test.go:5:3"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s:%d:%d", tt.fileName, tt.line, tt.column), func(t *testing.T) {
			result := FormatErrorLocation(tt.fileName, tt.line, tt.column)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected location format to contain '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
