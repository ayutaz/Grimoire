package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// File errors
	FileNotFound      ErrorType = "FILE_NOT_FOUND"
	UnsupportedFormat ErrorType = "UNSUPPORTED_FORMAT"
	FileReadError     ErrorType = "FILE_READ_ERROR"
	FileWriteError    ErrorType = "FILE_WRITE_ERROR"

	// Detection errors
	NoSymbolsDetected    ErrorType = "NO_SYMBOLS_DETECTED"
	NoOuterCircle        ErrorType = "NO_OUTER_CIRCLE"
	InvalidSymbolShape   ErrorType = "INVALID_SYMBOL_SHAPE"
	ImageProcessingError ErrorType = "IMAGE_PROCESSING_ERROR"

	// Parser errors
	SyntaxError          ErrorType = "SYNTAX_ERROR"
	UnexpectedSymbol     ErrorType = "UNEXPECTED_SYMBOL"
	MissingMainEntry     ErrorType = "MISSING_MAIN_ENTRY"
	InvalidConnection    ErrorType = "INVALID_CONNECTION"
	UnbalancedExpression ErrorType = "UNBALANCED_EXPRESSION"

	// Compiler errors
	CompilationError     ErrorType = "COMPILATION_ERROR"
	UnsupportedOperation ErrorType = "UNSUPPORTED_OPERATION"

	// Runtime errors
	ExecutionError ErrorType = "EXECUTION_ERROR"
)

// GrimoireError represents a custom error with context
type GrimoireError struct {
	Type       ErrorType
	Message    string
	Details    string
	Suggestion string
	Line       int
	Column     int
	FileName   string
	InnerError error
}

// Error implements the error interface
func (e *GrimoireError) Error() string {
	var parts []string

	// Main error message
	parts = append(parts, fmt.Sprintf("[%s] %s", e.Type, e.Message))

	// Add location if available
	if e.FileName != "" {
		if e.Line > 0 && e.Column > 0 {
			parts = append(parts, fmt.Sprintf("  at %s:%d:%d", e.FileName, e.Line, e.Column))
		} else if e.Line > 0 {
			parts = append(parts, fmt.Sprintf("  at %s:%d", e.FileName, e.Line))
		} else {
			parts = append(parts, fmt.Sprintf("  in %s", e.FileName))
		}
	}

	// Add details if available
	if e.Details != "" {
		parts = append(parts, fmt.Sprintf("  Details: %s", e.Details))
	}

	// Add suggestion if available
	if e.Suggestion != "" {
		parts = append(parts, fmt.Sprintf("  Suggestion: %s", e.Suggestion))
	}

	// Add inner error if available
	if e.InnerError != nil {
		parts = append(parts, fmt.Sprintf("  Caused by: %v", e.InnerError))
	}

	return strings.Join(parts, "\n")
}

// Unwrap returns the inner error
func (e *GrimoireError) Unwrap() error {
	return e.InnerError
}

// NewError creates a new GrimoireError
func NewError(errorType ErrorType, message string) *GrimoireError {
	return &GrimoireError{
		Type:    errorType,
		Message: message,
	}
}

// WithDetails adds details to the error
func (e *GrimoireError) WithDetails(details string) *GrimoireError {
	e.Details = details
	return e
}

// WithSuggestion adds a suggestion to the error
func (e *GrimoireError) WithSuggestion(suggestion string) *GrimoireError {
	e.Suggestion = suggestion
	return e
}

// WithLocation adds file location to the error
func (e *GrimoireError) WithLocation(fileName string, line, column int) *GrimoireError {
	e.FileName = fileName
	e.Line = line
	e.Column = column
	return e
}

// WithInnerError wraps another error
func (e *GrimoireError) WithInnerError(err error) *GrimoireError {
	e.InnerError = err
	return e
}

// Helper functions for common errors

// FileNotFoundError creates a file not found error
func FileNotFoundError(path string) *GrimoireError {
	return NewError(FileNotFound, fmt.Sprintf("Image file not found: %s", path)).
		WithSuggestion("Please check the file path and ensure the file exists")
}

// UnsupportedFormatError creates an unsupported format error
func UnsupportedFormatError(format string) *GrimoireError {
	return NewError(UnsupportedFormat, fmt.Sprintf("Unsupported image format: %s", format)).
		WithSuggestion("Grimoire supports PNG and JPEG image formats")
}

// NoSymbolsError creates a no symbols detected error
func NoSymbolsError() *GrimoireError {
	return NewError(NoSymbolsDetected, "No symbols were detected in the image").
		WithSuggestion("Ensure the image contains clear magical symbols with good contrast")
}

// NoOuterCircleError creates a no outer circle error
func NoOuterCircleError() *GrimoireError {
	return NewError(NoOuterCircle, "No outer circle detected in the magic diagram").
		WithDetails("All Grimoire programs must be enclosed in a magic circle").
		WithSuggestion("Draw a clear circle around your entire program")
}

// SyntaxErrorAt creates a syntax error with location
func SyntaxErrorAt(message string, symbolType string, x, y float64) *GrimoireError {
	return NewError(SyntaxError, message).
		WithDetails(fmt.Sprintf("Symbol type: %s at position (%.0f, %.0f)", symbolType, x, y))
}

// UnexpectedSymbolError creates an unexpected symbol error
func UnexpectedSymbolError(symbolType string, expected string, x, y float64) *GrimoireError {
	return NewError(UnexpectedSymbol, fmt.Sprintf("Unexpected symbol: %s", symbolType)).
		WithDetails(fmt.Sprintf("Expected: %s at position (%.0f, %.0f)", expected, x, y)).
		WithSuggestion("Check the symbol placement and connections in your diagram")
}

// IsGrimoireError checks if an error is a GrimoireError
func IsGrimoireError(err error) bool {
	_, ok := err.(*GrimoireError)
	return ok
}

// GetErrorType returns the error type if it's a GrimoireError
func GetErrorType(err error) (ErrorType, bool) {
	if ge, ok := err.(*GrimoireError); ok {
		return ge.Type, true
	}
	return "", false
}
