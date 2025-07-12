package errors

import (
	"fmt"
	"strings"

	"github.com/ayutaz/grimoire/internal/i18n"
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

	// Validation errors
	ValidationError ErrorType = "VALIDATION_ERROR"

	// I/O errors
	IOError ErrorType = "IO_ERROR"
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

	// Main error message with localized error type
	errorTypeStr := getLocalizedErrorType(e.Type)
	parts = append(parts, fmt.Sprintf("[%s] %s", errorTypeStr, e.Message))

	// Add location if available
	if e.FileName != "" {
		switch {
		case e.Line > 0 && e.Column > 0:
			parts = append(parts, i18n.Tf("error.at_location", e.FileName, e.Line, e.Column))
		case e.Line > 0:
			parts = append(parts, i18n.Tf("error.at_line", e.FileName, e.Line))
		default:
			parts = append(parts, i18n.Tf("error.in_file", e.FileName))
		}
	}

	// Add details if available
	if e.Details != "" {
		parts = append(parts, i18n.Tf("error.details", e.Details))
	}

	// Add suggestion if available
	if e.Suggestion != "" {
		parts = append(parts, i18n.Tf("error.suggestion", e.Suggestion))
	}

	// Add inner error if available
	if e.InnerError != nil {
		parts = append(parts, i18n.Tf("error.caused_by", e.InnerError))
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
	return NewError(FileNotFound, i18n.Tf("msg.image_file_not_found", path)).
		WithSuggestion(i18n.T("suggest.check_file_path"))
}

// UnsupportedFormatError creates an unsupported format error
func UnsupportedFormatError(format string) *GrimoireError {
	return NewError(UnsupportedFormat, i18n.Tf("msg.unsupported_image_format", format)).
		WithSuggestion(i18n.T("suggest.supported_formats"))
}

// NoSymbolsError creates a no symbols detected error
func NoSymbolsError() *GrimoireError {
	return NewError(NoSymbolsDetected, i18n.T("msg.no_symbols_detected")).
		WithSuggestion(i18n.T("suggest.ensure_clear_symbols"))
}

// NoOuterCircleError creates a no outer circle error
func NoOuterCircleError() *GrimoireError {
	return NewError(NoOuterCircle, i18n.T("msg.no_outer_circle")).
		WithDetails(i18n.T("detail.all_programs_need_circle")).
		WithSuggestion(i18n.T("suggest.draw_clear_circle"))
}

// SyntaxErrorAt creates a syntax error with location
func SyntaxErrorAt(message, symbolType string, x, y float64) *GrimoireError {
	return NewError(SyntaxError, message).
		WithDetails(i18n.Tf("detail.symbol_type_at_position", symbolType, x, y))
}

// UnexpectedSymbolError creates an unexpected symbol error
func UnexpectedSymbolError(symbolType, expected string, x, y float64) *GrimoireError {
	return NewError(UnexpectedSymbol, i18n.Tf("msg.unexpected_symbol", symbolType)).
		WithDetails(i18n.Tf("detail.expected_at_position", expected, x, y)).
		WithSuggestion(i18n.T("suggest.check_symbol_placement"))
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

// getLocalizedErrorType returns the localized error type string
func getLocalizedErrorType(errorType ErrorType) string {
	switch errorType {
	case FileNotFound:
		return i18n.T("error.file_not_found")
	case UnsupportedFormat:
		return i18n.T("error.unsupported_format")
	case FileReadError:
		return i18n.T("error.file_read_error")
	case FileWriteError:
		return i18n.T("error.file_write_error")
	case NoSymbolsDetected:
		return i18n.T("error.no_symbols_detected")
	case NoOuterCircle:
		return i18n.T("error.no_outer_circle")
	case InvalidSymbolShape:
		return i18n.T("error.invalid_symbol_shape")
	case ImageProcessingError:
		return i18n.T("error.image_processing_error")
	case SyntaxError:
		return i18n.T("error.syntax_error")
	case UnexpectedSymbol:
		return i18n.T("error.unexpected_symbol")
	case MissingMainEntry:
		return i18n.T("error.missing_main_entry")
	case InvalidConnection:
		return i18n.T("error.invalid_connection")
	case UnbalancedExpression:
		return i18n.T("error.unbalanced_expression")
	case CompilationError:
		return i18n.T("error.compilation_error")
	case UnsupportedOperation:
		return i18n.T("error.unsupported_operation")
	case ExecutionError:
		return i18n.T("error.execution_error")
	default:
		return string(errorType)
	}
}
