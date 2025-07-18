package i18n

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// Language represents a supported language
type Language string

const (
	English  Language = "en"
	Japanese Language = "ja"
)

// Message represents a localizable message
type Message struct {
	ID string
	En string
	Ja string
}

// Localizer handles message localization
type Localizer struct {
	lang     Language
	messages map[string]Message
	mu       sync.RWMutex
}

var (
	defaultLocalizer *Localizer
	once             sync.Once
)

// Init initializes the default localizer
func Init() {
	once.Do(func() {
		lang := detectLanguage()
		defaultLocalizer = NewLocalizer(lang)
		defaultLocalizer.loadMessages()
	})
}

// detectLanguage determines the language from environment
func detectLanguage() Language {
	// Check GRIMOIRE_LANG environment variable first
	if lang := os.Getenv("GRIMOIRE_LANG"); lang != "" {
		switch strings.ToLower(lang) {
		case "ja", "japanese":
			return Japanese
		case "en", "english":
			return English
		}
	}

	// Check standard LANG environment variable
	if lang := os.Getenv("LANG"); lang != "" {
		if strings.HasPrefix(strings.ToLower(lang), "ja") {
			return Japanese
		}
	}

	// Default to Japanese as requested
	return Japanese
}

// NewLocalizer creates a new localizer with the specified language
func NewLocalizer(lang Language) *Localizer {
	return &Localizer{
		lang:     lang,
		messages: make(map[string]Message),
	}
}

// SetLanguage changes the current language
func SetLanguage(lang Language) {
	Init()
	defaultLocalizer.mu.Lock()
	defer defaultLocalizer.mu.Unlock()
	defaultLocalizer.lang = lang
}

// GetLanguage returns the current language
func GetLanguage() Language {
	Init()
	defaultLocalizer.mu.RLock()
	defer defaultLocalizer.mu.RUnlock()
	return defaultLocalizer.lang
}

// T translates a message ID
func T(id string) string {
	Init()
	return defaultLocalizer.Translate(id)
}

// Tf translates a message ID with formatting
func Tf(id string, args ...interface{}) string {
	Init()
	return defaultLocalizer.TranslateF(id, args...)
}

// Translate returns the localized message for the given ID
func (l *Localizer) Translate(id string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	msg, ok := l.messages[id]
	if !ok {
		return id // Return ID if translation not found
	}

	switch l.lang {
	case Japanese:
		return msg.Ja
	default:
		return msg.En
	}
}

// TranslateF returns the localized message with formatting
func (l *Localizer) TranslateF(id string, args ...interface{}) string {
	template := l.Translate(id)
	return fmt.Sprintf(template, args...)
}

// loadMessages loads all message translations
func (l *Localizer) loadMessages() {
	messages := []Message{
		// Error type descriptions
		{ID: "error.file_not_found", En: "FILE_NOT_FOUND", Ja: "ファイルが見つかりません"},
		{ID: "error.unsupported_format", En: "UNSUPPORTED_FORMAT", Ja: "サポートされていない形式"},
		{ID: "error.file_read_error", En: "FILE_READ_ERROR", Ja: "ファイル読み込みエラー"},
		{ID: "error.file_write_error", En: "FILE_WRITE_ERROR", Ja: "ファイル書き込みエラー"},
		{ID: "error.no_symbols_detected", En: "NO_SYMBOLS_DETECTED", Ja: "シンボルが検出されません"},
		{ID: "error.no_outer_circle", En: "NO_OUTER_CIRCLE", Ja: "外周円が検出されません"},
		{ID: "error.invalid_symbol_shape", En: "INVALID_SYMBOL_SHAPE", Ja: "無効なシンボル形状"},
		{ID: "error.image_processing_error", En: "IMAGE_PROCESSING_ERROR", Ja: "画像処理エラー"},
		{ID: "error.syntax_error", En: "SYNTAX_ERROR", Ja: "構文エラー"},
		{ID: "error.unexpected_symbol", En: "UNEXPECTED_SYMBOL", Ja: "予期しないシンボル"},
		{ID: "error.missing_main_entry", En: "MISSING_MAIN_ENTRY", Ja: "メインエントリーポイントが見つかりません"},
		{ID: "error.cannot_compile_nil_ast", En: "Cannot compile nil AST", Ja: "nilのASTはコンパイルできません"},
		{ID: "error.invalid_connection", En: "INVALID_CONNECTION", Ja: "無効な接続"},
		{ID: "error.unbalanced_expression", En: "UNBALANCED_EXPRESSION", Ja: "式のバランスが取れていません"},
		{ID: "error.compilation_error", En: "COMPILATION_ERROR", Ja: "コンパイルエラー"},
		{ID: "error.unsupported_operation", En: "UNSUPPORTED_OPERATION", Ja: "サポートされていない操作"},
		{ID: "error.execution_error", En: "EXECUTION_ERROR", Ja: "実行エラー"},
		{ID: "error.validation_error", En: "VALIDATION_ERROR", Ja: "検証エラー"},
		{ID: "error.io_error", En: "IO_ERROR", Ja: "I/Oエラー"},

		// Error messages
		{ID: "msg.image_file_not_found", En: "Image file not found: %s", Ja: "画像ファイルが見つかりません: %s"},
		{ID: "msg.unsupported_image_format", En: "Unsupported image format: %s", Ja: "サポートされていない画像形式: %s"},
		{ID: "msg.no_symbols_detected", En: "No symbols were detected in the image", Ja: "画像内にシンボルが検出されませんでした"},
		{ID: "msg.no_outer_circle", En: "No outer circle detected in the magic diagram", Ja: "魔法陣に外周円が検出されませんでした"},
		{ID: "msg.unexpected_symbol", En: "Unexpected symbol: %s", Ja: "予期しないシンボル: %s"},
		{ID: "msg.failed_execute_python", En: "Failed to execute generated Python code", Ja: "生成されたPythonコードの実行に失敗しました"},
		{ID: "msg.failed_write_output", En: "Failed to write output file", Ja: "出力ファイルの書き込みに失敗しました"},
		{ID: "msg.error_occurred", En: "An error occurred", Ja: "エラーが発生しました"},

		// Suggestions
		{ID: "suggest.check_file_path", En: "Please check the file path and ensure the file exists",
			Ja: "ファイルパスを確認し、ファイルが存在することを確認してください"},
		{ID: "suggest.supported_formats", En: "Grimoire supports PNG and JPEG image formats",
			Ja: "GrimoireはPNGおよびJPEG画像形式をサポートしています"},
		{ID: "suggest.ensure_clear_symbols", En: "Ensure the image contains clear magical symbols with good contrast",
			Ja: "画像に明確でコントラストの良い魔法シンボルが含まれていることを確認してください"},
		{ID: "suggest.add_double_circle", En: "Add a double circle symbol to define the main entry point",
			Ja: "メインエントリーポイントを定義するためにダブルサークルシンボルを追加してください"},
		{ID: "suggest.draw_clear_circle", En: "Draw a clear circle around your entire program", Ja: "プログラム全体を囲む明確な円を描いてください"},
		{ID: "suggest.check_symbol_placement", En: "Check the symbol placement and connections in your diagram",
			Ja: "図のシンボルの配置と接続を確認してください"},
		{ID: "suggest.check_python_installed", En: "Check that Python 3 is installed and in your PATH",
			Ja: "Python 3がインストールされ、PATHに含まれていることを確認してください"},

		// Enhanced error hints
		{ID: "error.hint.missing_main_entry",
			En: "Add a double circle (◎) symbol at the top of your magic circle to define the main entry point",
			Ja: "魔法陣の上部にダブルサークル（◎）シンボルを追加してメインエントリーポイントを定義してください"},
		{ID: "error.detail.missing_main_entry",
			En: "The double circle (DoubleCircle) symbol was not detected in the image",
			Ja: "二重円（DoubleCircle）シンボルが画像内で検出されませんでした"},
		{ID: "error.hint.no_outer_circle",
			En: "Draw a complete circle around your entire program. The circle should be closed and clearly visible",
			Ja: "プログラム全体を囲む完全な円を描いてください。円は閉じていて明確に見える必要があります"},
		{ID: "error.detail.no_outer_circle",
			En: "No outer circle was detected. All Grimoire programs must be enclosed in a magic circle",
			Ja: "外周円が検出されませんでした。すべてのGrimoireプログラムは魔法陣で囲まれている必要があります"},
		{ID: "error.hint.unbalanced_expression",
			En: "Check that all operators have the correct number of operands connected",
			Ja: "すべての演算子に正しい数のオペランドが接続されていることを確認してください"},
		{ID: "error.hint.file_not_found",
			En: "Check the file path. Current directory: %s",
			Ja: "ファイルパスを確認してください。現在のディレクトリ: %s"},
		{ID: "error.hint.image_processing",
			En: "Ensure the image is not corrupted and has sufficient resolution (recommended: at least 600x600 pixels)",
			Ja: "画像が破損していないこと、十分な解像度があることを確認してください（推奨: 最低600x600ピクセル）"},
		{ID: "error.detail.image_processing",
			En: "Failed to process the image. Common causes: corrupted file, unsupported color mode, or insufficient memory",
			Ja: "画像の処理に失敗しました。一般的な原因: ファイルの破損、サポートされていないカラーモード、メモリ不足"},

		// Error formatting helpers
		{ID: "error.did_you_mean_single", En: "Did you mean: %s?", Ja: "もしかして: %s?"},
		{ID: "error.did_you_mean_multiple", En: "Did you mean one of: %s?", Ja: "もしかして次のいずれか: %s?"},

		// Details
		{ID: "detail.all_programs_need_circle", En: "All Grimoire programs must be enclosed in a magic circle",
			Ja: "すべてのGrimoireプログラムは魔法陣で囲まれている必要があります"},
		{ID: "detail.symbol_type_at_position", En: "Symbol type: %s at position (%.0f, %.0f)",
			Ja: "シンボルタイプ: %s 位置: (%.0f, %.0f)"},
		{ID: "detail.expected_at_position", En: "Expected: %s at position (%.0f, %.0f)", Ja: "期待される値: %s 位置: (%.0f, %.0f)"},

		// CLI messages
		{ID: "cli.description_short", En: "A visual programming language using magic circles", Ja: "魔法陣を使用するビジュアルプログラミング言語"},
		{ID: "cli.description_long",
			En: "Grimoire is a visual programming language where programs are expressed as magic circles.\n" +
				"Draw your spells and watch them come to life!",
			Ja: "Grimoireはプログラムを魔法陣として表現するビジュアルプログラミング言語です。\n呪文を描いて、それが実現するのを見てください！"},
		{ID: "cli.run_description", En: "Run a Grimoire program", Ja: "Grimoireプログラムを実行"},
		{ID: "cli.compile_description", En: "Compile a Grimoire program to Python", Ja: "GrimoireプログラムをPythonにコンパイル"},
		{ID: "cli.debug_description", En: "Debug a Grimoire program (show detected symbols)",
			Ja: "Grimoireプログラムをデバッグ（検出されたシンボルを表示）"},
		{ID: "cli.validate_description", En: "Validate a magic circle for correctness", Ja: "魔法陣の正しさを検証"},
		{ID: "cli.format_description", En: "Format and beautify a magic circle", Ja: "魔法陣をフォーマットして美しくする"},
		{ID: "cli.optimize_description", En: "Optimize a magic circle program", Ja: "魔法陣プログラムを最適化"},
		{ID: "cli.output_flag_description", En: "Output file path", Ja: "出力ファイルパス"},
		{ID: "cli.format_output_flag_description", En: "Output formatted image path", Ja: "フォーマット済み画像の出力パス"},
		{ID: "cli.optimize_output_flag_description", En: "Output optimized code path (use - for stdout)",
			Ja: "最適化されたコードの出力パス（標準出力には-を使用）"},
		{ID: "cli.language_flag_description", En: "Language (en/ja)", Ja: "言語 (en/ja)"},
		{ID: "cli.compile_success", En: "Successfully compiled to %s\n", Ja: "%s へのコンパイルに成功しました\n"},

		// Debug messages
		{ID: "debug.header", En: "\n=== Debug Information for %s ===\n", Ja: "\n=== %s のデバッグ情報 ===\n"},
		{ID: "debug.detected_summary", En: "Detected %d symbols and %d connections\n\n", Ja: "%d個のシンボルと%d個の接続を検出しました\n\n"},
		{ID: "debug.symbols_header", En: "Symbols:", Ja: "シンボル:"},
		{ID: "debug.connections_header", En: "\nConnections:", Ja: "\n接続:"},
		{ID: "debug.symbol_info",
			En: "  [%d] Type: %-15s Position: (%.0f, %.0f) Size: %.1f Pattern: %s\n",
			Ja: "  [%d] タイプ: %-15s 位置: (%.0f, %.0f) サイズ: %.1f パターン: %s\n"},
		{ID: "debug.connection_info", En: "  [%d] %s -> %s (%s)\n", Ja: "  [%d] %s -> %s (%s)\n"},

		// Validate messages
		{ID: "validate.no_outer_circle", En: "No outer magic circle found", Ja: "外周の魔法陣が見つかりません"},
		{ID: "validate.no_main_function", En: "No main function (double circle) found", Ja: "メイン関数（二重円）が見つかりません"},
		{ID: "validate.orphaned_symbol", En: "Orphaned symbol: %s at (%.0f, %.0f)", Ja: "孤立したシンボル: %s 位置: (%.0f, %.0f)"},
		{ID: "validate.success", En: "✓ Magic circle is valid!\n", Ja: "✓ 魔法陣は有効です！\n"},
		{ID: "validate.symbols_found", En: "  - Symbols found: %d\n", Ja: "  - シンボル数: %d\n"},
		{ID: "validate.connections_found", En: "  - Connections found: %d\n", Ja: "  - 接続数: %d\n"},
		{ID: "validate.issues_found", En: "✗ Validation issues found:\n", Ja: "✗ 検証で問題が見つかりました:\n"},
		{ID: "validate.failed", En: "Magic circle validation failed", Ja: "魔法陣の検証に失敗しました"},

		// Format messages
		{ID: "format.analyzing", En: "Analyzing magic circle structure...\n", Ja: "魔法陣の構造を分析中...\n"},
		{ID: "format.align_symbols", En: "Consider aligning %s and %s symbols", Ja: "%sと%sのシンボルを整列させることを検討してください"},
		{ID: "format.straighten_connection", En: "Straighten connection between %s and %s", Ja: "%sと%s間の接続を真っ直ぐにしてください"},
		{ID: "format.well_formatted", En: "✓ Magic circle is well-formatted!\n", Ja: "✓ 魔法陣は適切にフォーマットされています！\n"},
		{ID: "format.suggestions", En: "Formatting suggestions:\n", Ja: "フォーマットの提案:\n"},
		{ID: "format.output_note", En: "\nNote: Automatic formatting to %s is not yet implemented.\n",
			Ja: "\n注意: %sへの自動フォーマットはまだ実装されていません。\n"},

		// Optimize messages
		{ID: "optimize.analyzing", En: "Analyzing program for optimization opportunities...\n",
			Ja: "最適化の機会を探してプログラムを分析中...\n"},
		{ID: "optimize.unused_variable", En: "Unused variable: %s", Ja: "未使用の変数: %s"},
		{ID: "optimize.duplicate_operation", En: "Duplicate operation detected", Ja: "重複した操作が検出されました"},
		{ID: "optimize.well_optimized", En: "✓ Program is well-optimized!\n", Ja: "✓ プログラムは十分に最適化されています！\n"},
		{ID: "optimize.suggestions", En: "Optimization suggestions:\n", Ja: "最適化の提案:\n"},
		{ID: "optimize.optimized_code", En: "Optimized code:", Ja: "最適化されたコード:"},
		{ID: "optimize.saved_to", En: "Optimized code saved to: %s\n", Ja: "最適化されたコードを保存しました: %s\n"},

		// Error formatting
		{ID: "error.at_location", En: "  at %s:%d:%d", Ja: "  場所: %s:%d:%d"},
		{ID: "error.at_line", En: "  at %s:%d", Ja: "  場所: %s:%d"},
		{ID: "error.in_file", En: "  in %s", Ja: "  ファイル: %s"},
		{ID: "error.details", En: "  Details: %s", Ja: "  詳細: %s"},
		{ID: "error.suggestion", En: "  Suggestion: %s", Ja: "  提案: %s"},
		{ID: "error.caused_by", En: "  Caused by: %v", Ja: "  原因: %v"},
		{ID: "error.error_prefix", En: "Error: %v\n", Ja: "エラー: %v\n"},
		{ID: "error.execution_time", En: "Execution time: %v\n", Ja: "実行時間: %v\n"},
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	for _, msg := range messages {
		l.messages[msg.ID] = msg
	}
}
