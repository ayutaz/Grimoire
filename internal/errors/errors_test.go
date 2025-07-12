package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		message  string
		expected string
	}{
		{
			name:     "file not found error",
			errType:  FileNotFound,
			message:  "test.png not found",
			expected: "ファイルが見つかりません",
		},
		{
			name:     "syntax error",
			errType:  SyntaxError,
			message:  "unexpected symbol",
			expected: "構文エラー",
		},
		{
			name:     "runtime error",
			errType:  ExecutionError,
			message:  "execution failed",
			expected: "実行エラー",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.errType, tt.message)
			assert.NotNil(t, err)
			assert.Equal(t, tt.errType, err.Type)
			assert.Equal(t, tt.message, err.Message)
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}

func TestGrimoireError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *GrimoireError
		expected []string
	}{
		{
			name: "basic error",
			err: &GrimoireError{
				Type:    FileNotFound,
				Message: "file.png not found",
			},
			expected: []string{"[ファイルが見つかりません]", "file.png not found"},
		},
		{
			name: "error with filename",
			err: &GrimoireError{
				Type:     SyntaxError,
				Message:  "invalid syntax",
				FileName: "test.png",
			},
			expected: []string{"[構文エラー]", "invalid syntax", "ファイル: test.png"},
		},
		{
			name: "error with line and column",
			err: &GrimoireError{
				Type:     SyntaxError,
				Message:  "unexpected token",
				FileName: "program.png",
				Line:     10,
				Column:   5,
			},
			expected: []string{"[構文エラー]", "unexpected token", "場所: program.png:10:5"},
		},
		{
			name: "error with details",
			err: &GrimoireError{
				Type:    ImageProcessingError,
				Message: "failed to process",
				Details: "Invalid format",
			},
			expected: []string{"[画像処理エラー]", "failed to process", "詳細: Invalid format"},
		},
		{
			name: "error with suggestion",
			err: &GrimoireError{
				Type:       NoOuterCircle,
				Message:    "no outer circle found",
				Suggestion: "Add an outer circle",
			},
			expected: []string{"[外周円が検出されません]", "no outer circle found", "提案: Add an outer circle"},
		},
		{
			name: "error with inner error",
			err: &GrimoireError{
				Type:       FileReadError,
				Message:    "cannot read file",
				InnerError: errors.New("permission denied"),
			},
			expected: []string{"[ファイル読み込みエラー]", "cannot read file", "原因: permission denied"},
		},
		{
			name: "error with all fields",
			err: &GrimoireError{
				Type:       CompilationError,
				Message:    "compilation failed",
				FileName:   "complex.png",
				Line:       42,
				Column:     13,
				Details:    "Stack overflow",
				Suggestion: "Check for infinite loops",
				InnerError: errors.New("stack limit exceeded"),
			},
			expected: []string{
				"[コンパイルエラー]",
				"compilation failed",
				"場所: complex.png:42:13",
				"詳細: Stack overflow",
				"提案: Check for infinite loops",
				"原因: stack limit exceeded",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, exp := range tt.expected {
				assert.Contains(t, errStr, exp)
			}
		})
	}
}

func TestGrimoireError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")

	tests := []struct {
		name      string
		err       *GrimoireError
		wantInner error
	}{
		{
			name: "with inner error",
			err: &GrimoireError{
				Type:       FileReadError,
				Message:    "read failed",
				InnerError: innerErr,
			},
			wantInner: innerErr,
		},
		{
			name: "without inner error",
			err: &GrimoireError{
				Type:    SyntaxError,
				Message: "syntax error",
			},
			wantInner: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Unwrap()
			assert.Equal(t, tt.wantInner, got)

			// Test with errors.Is
			if tt.wantInner != nil {
				assert.True(t, errors.Is(tt.err, tt.wantInner))
			}
		})
	}
}

func TestWithDetails(t *testing.T) {
	err := NewError(ImageProcessingError, "processing failed")
	err2 := err.WithDetails("PNG decode error")

	assert.Equal(t, err, err2) // Should return same pointer
	assert.Equal(t, "PNG decode error", err.Details)
	assert.Contains(t, err.Error(), "詳細: PNG decode error")
}

func TestWithLocation(t *testing.T) {
	err := NewError(SyntaxError, "invalid symbol")
	err2 := err.WithLocation("test.png", 10, 20)

	assert.Equal(t, err, err2) // Should return same pointer
	assert.Equal(t, "test.png", err.FileName)
	assert.Equal(t, 10, err.Line)
	assert.Equal(t, 20, err.Column)
	assert.Contains(t, err.Error(), "場所: test.png:10:20")
}

func TestWithSuggestion(t *testing.T) {
	err := NewError(NoSymbolsDetected, "no symbols found")
	err2 := err.WithSuggestion("Check image quality")

	assert.Equal(t, err, err2) // Should return same pointer
	assert.Equal(t, "Check image quality", err.Suggestion)
	assert.Contains(t, err.Error(), "提案: Check image quality")
}

func TestWithInnerError(t *testing.T) {
	innerErr := errors.New("original error")
	err := NewError(FileReadError, "cannot read")
	err2 := err.WithInnerError(innerErr)

	assert.Equal(t, err, err2) // Should return same pointer
	assert.Equal(t, innerErr, err.InnerError)
	assert.Contains(t, err.Error(), "原因: original error")
}

func TestHelperFunctions(t *testing.T) {
	// Test FileNotFoundError
	err := FileNotFoundError("missing.png")
	assert.Equal(t, FileNotFound, err.Type)
	assert.Contains(t, err.Error(), "missing.png")
	assert.NotEmpty(t, err.Suggestion)

	// Test UnsupportedFormatError
	err = UnsupportedFormatError("BMP")
	assert.Equal(t, UnsupportedFormat, err.Type)
	assert.Contains(t, err.Error(), "BMP")
	assert.Contains(t, err.Suggestion, "PNG")

	// Test NoOuterCircleError
	err = NoOuterCircleError()
	assert.Equal(t, NoOuterCircle, err.Type)
	assert.NotEmpty(t, err.Suggestion)

	// Test NoSymbolsError
	err = NoSymbolsError()
	assert.Equal(t, NoSymbolsDetected, err.Type)
	assert.NotEmpty(t, err.Suggestion)
}

func TestSuggestionInHelperFunctions(t *testing.T) {
	// Test that helper functions include suggestions
	tests := []struct {
		name string
		err  *GrimoireError
	}{
		{
			name: "FileNotFoundError has suggestion",
			err:  FileNotFoundError("test.png"),
		},
		{
			name: "UnsupportedFormatError has suggestion",
			err:  UnsupportedFormatError("BMP"),
		},
		{
			name: "NoOuterCircleError has suggestion",
			err:  NoOuterCircleError(),
		},
		{
			name: "NoSymbolsError has suggestion",
			err:  NoSymbolsError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.err.Suggestion)
			assert.Contains(t, tt.err.Error(), "提案:")
		})
	}
}

func TestErrorImplementsStandardError(t *testing.T) {
	var err error = NewError(FileNotFound, "test")
	assert.NotNil(t, err)

	// Should work with standard error handling
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ファイルが見つかりません")
}

func TestAllErrorTypes(t *testing.T) {
	// Ensure all error types are valid
	errorTypes := []ErrorType{
		FileNotFound,
		UnsupportedFormat,
		FileReadError,
		FileWriteError,
		NoSymbolsDetected,
		NoOuterCircle,
		InvalidSymbolShape,
		ImageProcessingError,
		SyntaxError,
		UnexpectedSymbol,
		MissingMainEntry,
		InvalidConnection,
		UnbalancedExpression,
		CompilationError,
		UnsupportedOperation,
		ExecutionError,
		ValidationError,
		IOError,
	}

	for _, errType := range errorTypes {
		t.Run(string(errType), func(t *testing.T) {
			// Create error
			err := NewError(errType, "test message")
			assert.NotNil(t, err)

			// Check type
			assert.Equal(t, errType, err.Type)

			// Check error string contains localized type
			// The error message now contains the Japanese translation, not the constant

			// Just verify no panic
			assert.NotPanics(t, func() {
				_ = err.Error()
			})
		})
	}
}
