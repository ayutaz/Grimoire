package i18n

import (
	"os"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name         string
		grimoireLang string
		lang         string
		expected     Language
	}{
		{
			name:         "GRIMOIRE_LANG set to ja",
			grimoireLang: "ja",
			expected:     Japanese,
		},
		{
			name:         "GRIMOIRE_LANG set to japanese",
			grimoireLang: "japanese",
			expected:     Japanese,
		},
		{
			name:         "GRIMOIRE_LANG set to en",
			grimoireLang: "en",
			expected:     English,
		},
		{
			name:         "GRIMOIRE_LANG set to english",
			grimoireLang: "english",
			expected:     English,
		},
		{
			name:     "LANG set to ja_JP.UTF-8",
			lang:     "ja_JP.UTF-8",
			expected: Japanese,
		},
		{
			name:     "No environment variables",
			expected: Japanese, // Default to Japanese
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origGrimoireLang := os.Getenv("GRIMOIRE_LANG")
			origLang := os.Getenv("LANG")

			// Set test env vars
			if tt.grimoireLang != "" {
				os.Setenv("GRIMOIRE_LANG", tt.grimoireLang)
			} else {
				os.Unsetenv("GRIMOIRE_LANG")
			}

			if tt.lang != "" {
				os.Setenv("LANG", tt.lang)
			} else {
				os.Unsetenv("LANG")
			}

			// Test
			result := detectLanguage()
			if result != tt.expected {
				t.Errorf("detectLanguage() = %v, want %v", result, tt.expected)
			}

			// Restore original env vars
			os.Setenv("GRIMOIRE_LANG", origGrimoireLang)
			os.Setenv("LANG", origLang)
		})
	}
}

func TestTranslation(t *testing.T) {
	// Test Japanese translations
	localizer := NewLocalizer(Japanese)
	localizer.loadMessages()

	tests := []struct {
		id       string
		expected string
	}{
		{"error.file_not_found", "ファイルが見つかりません"},
		{"msg.no_symbols_detected", "画像内にシンボルが検出されませんでした"},
		{"suggest.check_file_path", "ファイルパスを確認し、ファイルが存在することを確認してください"},
		{"cli.description_short", "魔法陣を使用するビジュアルプログラミング言語"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			result := localizer.Translate(tt.id)
			if result != tt.expected {
				t.Errorf("Translate(%s) = %s, want %s", tt.id, result, tt.expected)
			}
		})
	}

	// Test English translations
	localizer = NewLocalizer(English)
	localizer.loadMessages()

	englishTests := []struct {
		id       string
		expected string
	}{
		{"error.file_not_found", "FILE_NOT_FOUND"},
		{"msg.no_symbols_detected", "No symbols were detected in the image"},
		{"suggest.check_file_path", "Please check the file path and ensure the file exists"},
		{"cli.description_short", "A visual programming language using magic circles"},
	}

	for _, tt := range englishTests {
		t.Run("en_"+tt.id, func(t *testing.T) {
			result := localizer.Translate(tt.id)
			if result != tt.expected {
				t.Errorf("Translate(%s) = %s, want %s", tt.id, result, tt.expected)
			}
		})
	}
}

func TestTranslateF(t *testing.T) {
	localizer := NewLocalizer(Japanese)
	localizer.loadMessages()

	tests := []struct {
		id       string
		args     []interface{}
		expected string
	}{
		{
			id:       "msg.image_file_not_found",
			args:     []interface{}{"test.png"},
			expected: "画像ファイルが見つかりません: test.png",
		},
		{
			id:       "debug.detected_summary",
			args:     []interface{}{5, 3},
			expected: "5個のシンボルと3個の接続を検出しました\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			result := localizer.TranslateF(tt.id, tt.args...)
			if result != tt.expected {
				t.Errorf("TranslateF(%s, %v) = %s, want %s", tt.id, tt.args, result, tt.expected)
			}
		})
	}
}

func TestGlobalFunctions(t *testing.T) {
	// Set to Japanese
	SetLanguage(Japanese)

	if GetLanguage() != Japanese {
		t.Errorf("GetLanguage() = %v, want %v", GetLanguage(), Japanese)
	}

	// Test T function
	result := T("error.file_not_found")
	expected := "ファイルが見つかりません"
	if result != expected {
		t.Errorf("T() = %s, want %s", result, expected)
	}

	// Test Tf function
	result = Tf("msg.image_file_not_found", "test.png")
	expected = "画像ファイルが見つかりません: test.png"
	if result != expected {
		t.Errorf("Tf() = %s, want %s", result, expected)
	}

	// Switch to English
	SetLanguage(English)

	if GetLanguage() != English {
		t.Errorf("GetLanguage() = %v, want %v", GetLanguage(), English)
	}

	result = T("error.file_not_found")
	expected = "FILE_NOT_FOUND"
	if result != expected {
		t.Errorf("T() = %s, want %s", result, expected)
	}
}
