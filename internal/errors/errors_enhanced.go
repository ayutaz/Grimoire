package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorCode represents a unique error code for programmatic handling
type ErrorCode string

// Error codes for each error type
const (
	// File error codes (1000-1999)
	ErrCodeFileNotFound      ErrorCode = "E1001"
	ErrCodeUnsupportedFormat ErrorCode = "E1002"
	ErrCodeFileReadError     ErrorCode = "E1003"
	ErrCodeFileWriteError    ErrorCode = "E1004"

	// Detection error codes (2000-2999)
	ErrCodeNoSymbolsDetected    ErrorCode = "E2001"
	ErrCodeNoOuterCircle        ErrorCode = "E2002"
	ErrCodeInvalidSymbolShape   ErrorCode = "E2003"
	ErrCodeImageProcessingError ErrorCode = "E2004"

	// Parser error codes (3000-3999)
	ErrCodeSyntaxError          ErrorCode = "E3001"
	ErrCodeUnexpectedSymbol     ErrorCode = "E3002"
	ErrCodeMissingMainEntry     ErrorCode = "E3003"
	ErrCodeInvalidConnection    ErrorCode = "E3004"
	ErrCodeUnbalancedExpression ErrorCode = "E3005"

	// Compiler error codes (4000-4999)
	ErrCodeCompilationError     ErrorCode = "E4001"
	ErrCodeUnsupportedOperation ErrorCode = "E4002"

	// Runtime error codes (5000-5999)
	ErrCodeExecutionError ErrorCode = "E5001"

	// Validation error codes (6000-6999)
	ErrCodeValidationError ErrorCode = "E6001"

	// I/O error codes (7000-7999)
	ErrCodeIOError ErrorCode = "E7001"
)

// errorCodeMap maps ErrorType to ErrorCode
var errorCodeMap = map[ErrorType]ErrorCode{
	FileNotFound:         ErrCodeFileNotFound,
	UnsupportedFormat:    ErrCodeUnsupportedFormat,
	FileReadError:        ErrCodeFileReadError,
	FileWriteError:       ErrCodeFileWriteError,
	NoSymbolsDetected:    ErrCodeNoSymbolsDetected,
	NoOuterCircle:        ErrCodeNoOuterCircle,
	InvalidSymbolShape:   ErrCodeInvalidSymbolShape,
	ImageProcessingError: ErrCodeImageProcessingError,
	SyntaxError:          ErrCodeSyntaxError,
	UnexpectedSymbol:     ErrCodeUnexpectedSymbol,
	MissingMainEntry:     ErrCodeMissingMainEntry,
	InvalidConnection:    ErrCodeInvalidConnection,
	UnbalancedExpression: ErrCodeUnbalancedExpression,
	CompilationError:     ErrCodeCompilationError,
	UnsupportedOperation: ErrCodeUnsupportedOperation,
	ExecutionError:       ErrCodeExecutionError,
	ValidationError:      ErrCodeValidationError,
	IOError:              ErrCodeIOError,
}

// StackFrame represents a single frame in the stack trace
type StackFrame struct {
	Function string
	File     string
	Line     int
}

// EnhancedError extends GrimoireError with additional context
type EnhancedError struct {
	*GrimoireError
	Code       ErrorCode
	StackTrace []StackFrame
	Context    map[string]interface{} // Additional context data
}

// NewEnhancedError creates an enhanced error from a GrimoireError
func NewEnhancedError(err *GrimoireError) *EnhancedError {
	code, exists := errorCodeMap[err.Type]
	if !exists {
		code = ErrorCode("E0000") // Unknown error code
	}

	enhanced := &EnhancedError{
		GrimoireError: err,
		Code:          code,
		Context:       make(map[string]interface{}),
	}

	// Capture stack trace if in debug mode
	if IsDebugMode() {
		enhanced.CaptureStackTrace()
	}

	return enhanced
}

// CaptureStackTrace captures the current stack trace
func (e *EnhancedError) CaptureStackTrace() {
	const maxDepth = 32
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(3, pcs) // Skip runtime.Callers, CaptureStackTrace, and NewEnhancedError
	
	if n == 0 {
		return
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		
		// Skip internal runtime frames
		if strings.Contains(frame.File, "runtime/") {
			if !more {
				break
			}
			continue
		}

		e.StackTrace = append(e.StackTrace, StackFrame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})

		if !more {
			break
		}
	}
}

// WithContext adds context information to the error
func (e *EnhancedError) WithContext(key string, value interface{}) *EnhancedError {
	e.Context[key] = value
	return e
}

// Error implements the error interface with enhanced formatting
func (e *EnhancedError) Error() string {
	var parts []string

	// Add error code
	parts = append(parts, fmt.Sprintf("[%s] %s", e.Code, e.GrimoireError.Error()))

	// Add context if available
	if len(e.Context) > 0 && IsDebugMode() {
		contextParts := []string{"Context:"}
		for k, v := range e.Context {
			contextParts = append(contextParts, fmt.Sprintf("  %s: %v", k, v))
		}
		parts = append(parts, strings.Join(contextParts, "\n"))
	}

	// Add stack trace if available and in debug mode
	if len(e.StackTrace) > 0 && IsDebugMode() {
		stackParts := []string{"Stack Trace:"}
		for i, frame := range e.StackTrace {
			stackParts = append(stackParts, fmt.Sprintf("  %d. %s\n     %s:%d", 
				i+1, frame.Function, frame.File, frame.Line))
		}
		parts = append(parts, strings.Join(stackParts, "\n"))
	}

	return strings.Join(parts, "\n")
}

// GetCode returns the error code
func (e *EnhancedError) GetCode() ErrorCode {
	return e.Code
}

// GetStackTrace returns the stack trace
func (e *EnhancedError) GetStackTrace() []StackFrame {
	return e.StackTrace
}

// IsDebugMode checks if debug mode is enabled
func IsDebugMode() bool {
	// Check environment variable or global flag
	return getDebugMode()
}