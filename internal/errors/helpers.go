package errors

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ayutaz/grimoire/internal/i18n"
)

// ErrorWithHint creates an error with a helpful hint based on the error type
func ErrorWithHint(err *GrimoireError) *GrimoireError {
	switch err.Type {
	case MissingMainEntry:
		err.Suggestion = i18n.T("error.hint.missing_main_entry")
		err.Details = i18n.T("error.detail.missing_main_entry")
	case NoOuterCircle:
		err.Suggestion = i18n.T("error.hint.no_outer_circle")
		err.Details = i18n.T("error.detail.no_outer_circle")
	case UnbalancedExpression:
		err.Suggestion = i18n.T("error.hint.unbalanced_expression")
	case FileNotFound:
		if err.FileName != "" {
			dir := filepath.Dir(err.FileName)
			err.Suggestion = i18n.Tf("error.hint.file_not_found", dir)
		} else {
			err.Suggestion = i18n.T("suggest.check_file_path")
		}
	case ImageProcessingError:
		err.Suggestion = i18n.T("error.hint.image_processing")
		err.Details = i18n.T("error.detail.image_processing")
	}
	return err
}

// FormatErrorLocation formats error location information
func FormatErrorLocation(fileName string, line, column int) string {
	if fileName == "" {
		return ""
	}

	// Use relative path if possible
	relPath := fileName
	if cwd, err := filepath.Abs("."); err == nil {
		if rel, err := filepath.Rel(cwd, fileName); err == nil && !strings.HasPrefix(rel, "..") {
			relPath = rel
		}
	}

	if line > 0 && column > 0 {
		return fmt.Sprintf("%s:%d:%d", relPath, line, column)
	} else if line > 0 {
		return fmt.Sprintf("%s:%d", relPath, line)
	}
	return relPath
}

// SuggestSimilar suggests similar valid options for invalid input
func SuggestSimilar(invalid string, validOptions []string) string {
	if len(validOptions) == 0 {
		return ""
	}

	// Simple similarity check (could be enhanced with Levenshtein distance)
	var suggestions []string
	invalidLower := strings.ToLower(invalid)
	
	for _, option := range validOptions {
		optionLower := strings.ToLower(option)
		// Check if option contains the invalid string or vice versa
		if strings.Contains(optionLower, invalidLower) || strings.Contains(invalidLower, optionLower) {
			suggestions = append(suggestions, option)
		}
	}

	if len(suggestions) > 0 {
		if len(suggestions) == 1 {
			return i18n.Tf("error.did_you_mean_single", suggestions[0])
		}
		return i18n.Tf("error.did_you_mean_multiple", strings.Join(suggestions, ", "))
	}

	return ""
}

// ErrorContext provides additional context for errors
type ErrorContext struct {
	Operation   string                 // What operation was being performed
	InputFile   string                 // Input file being processed
	OutputFile  string                 // Output file being written
	Stage       string                 // Processing stage (detection, parsing, compilation)
	Metadata    map[string]interface{} // Additional metadata
}

// NewErrorContext creates a new error context
func NewErrorContext() *ErrorContext {
	return &ErrorContext{
		Metadata: make(map[string]interface{}),
	}
}

// WithOperation sets the operation being performed
func (ctx *ErrorContext) WithOperation(op string) *ErrorContext {
	ctx.Operation = op
	return ctx
}

// WithInputFile sets the input file
func (ctx *ErrorContext) WithInputFile(file string) *ErrorContext {
	ctx.InputFile = file
	return ctx
}

// WithOutputFile sets the output file
func (ctx *ErrorContext) WithOutputFile(file string) *ErrorContext {
	ctx.OutputFile = file
	return ctx
}

// WithStage sets the processing stage
func (ctx *ErrorContext) WithStage(stage string) *ErrorContext {
	ctx.Stage = stage
	return ctx
}

// WithMetadata adds metadata
func (ctx *ErrorContext) WithMetadata(key string, value interface{}) *ErrorContext {
	ctx.Metadata[key] = value
	return ctx
}

// ApplyToError applies the context to an error
func (ctx *ErrorContext) ApplyToError(err *GrimoireError) *GrimoireError {
	enhanced := NewEnhancedError(err)
	
	if ctx.Operation != "" {
		enhanced.WithContext("operation", ctx.Operation)
	}
	if ctx.InputFile != "" {
		enhanced.WithContext("input_file", ctx.InputFile)
		if err.FileName == "" {
			err.FileName = ctx.InputFile
		}
	}
	if ctx.OutputFile != "" {
		enhanced.WithContext("output_file", ctx.OutputFile)
	}
	if ctx.Stage != "" {
		enhanced.WithContext("stage", ctx.Stage)
	}
	for k, v := range ctx.Metadata {
		enhanced.WithContext(k, v)
	}

	// Return the enhanced error as a GrimoireError for compatibility
	// In practice, we might want to update the codebase to use EnhancedError directly
	return err
}