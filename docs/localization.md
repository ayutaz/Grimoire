# Grimoire Localization Guide

## Overview

Grimoire supports multiple languages for error messages and user interface text. Currently, Japanese and English are supported, with Japanese as the default language.

## Setting the Language

There are three ways to control the language:

### 1. Command Line Flag

Use the `--lang` or `-l` flag with any Grimoire command:

```bash
# Use Japanese (default)
grimoire run magic_circle.png --lang ja

# Use English
grimoire run magic_circle.png --lang en
```

### 2. Environment Variable

Set the `GRIMOIRE_LANG` environment variable:

```bash
# Set to Japanese
export GRIMOIRE_LANG=ja
# or
export GRIMOIRE_LANG=japanese

# Set to English
export GRIMOIRE_LANG=en
# or
export GRIMOIRE_LANG=english
```

### 3. System Locale

If no explicit language is set, Grimoire will check the system's `LANG` environment variable:

```bash
# If LANG starts with "ja", Japanese will be used
export LANG=ja_JP.UTF-8
```

## Priority

The language selection follows this priority order:
1. Command line flag (`--lang`)
2. `GRIMOIRE_LANG` environment variable
3. System `LANG` environment variable
4. Default to Japanese

## Adding New Translations

To add translations for new messages:

1. Edit `internal/i18n/i18n.go`
2. Add a new `Message` struct in the `loadMessages()` function:

```go
{
    ID: "msg.new_message",
    En: "English message",
    Ja: "日本語メッセージ",
}
```

3. Use the translation in your code:

```go
// Simple message
message := i18n.T("msg.new_message")

// Message with formatting
message := i18n.Tf("msg.with_param", param1, param2)
```

## Error Message Structure

Error messages in Grimoire follow a consistent structure:

```
[エラータイプ] エラーメッセージ
  場所: ファイル名:行:列
  詳細: 詳細情報
  提案: 解決方法の提案
  原因: 内部エラー
```

In English:

```
[ERROR_TYPE] Error message
  at filename:line:column
  Details: Additional details
  Suggestion: How to fix
  Caused by: Internal error
```

## Testing Localization

Run the localization tests:

```bash
go test ./internal/i18n/...
```

## Example Usage

```go
// In error handling
return errors.NewError(errors.FileNotFound, i18n.Tf("msg.image_file_not_found", path)).
    WithSuggestion(i18n.T("suggest.check_file_path"))

// In CLI
fmt.Printf(i18n.T("cli.compile_success"), outputPath)

// Debug output
fmt.Printf(i18n.T("debug.detected_summary"), len(symbols), len(connections))
```