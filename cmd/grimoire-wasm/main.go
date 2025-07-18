//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"syscall/js"

	"github.com/ayutaz/grimoire/internal/compiler"
	"github.com/ayutaz/grimoire/internal/detector"
	"github.com/ayutaz/grimoire/internal/parser"
)

// ProcessImageResult は処理結果を表す構造体
type ProcessImageResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error"`
	Code    string `json:"code"`
}

func main() {
	// WebAssembly用のグローバル関数を登録
	js.Global().Set("processGrimoireImage", js.FuncOf(processImage))
	js.Global().Set("validateGrimoireCode", js.FuncOf(validateCode))
	js.Global().Set("formatGrimoireCode", js.FuncOf(formatCode))

	// プログラムが終了しないようにブロック
	select {}
}

// processImage は画像を処理してPythonコードを生成・実行する
func processImage(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorResult("No image data provided")
	}

	// Base64エンコードされた画像データを取得
	imageDataBase64 := args[0].String()

	// Base64デコード
	imageData, err := base64.StdEncoding.DecodeString(imageDataBase64)
	if err != nil {
		return createErrorResult(fmt.Sprintf("Failed to decode image: %v", err))
	}

	// 画像からプログラムを検出
	det := detector.NewDetector(detector.Config{Debug: true})
	symbols, connections, err := det.DetectFromBytes(imageData)
	if err != nil {
		return createErrorResult(fmt.Sprintf("Failed to detect program: %v", err))
	}

	// デバッグ用：検出されたシンボル数を確認
	if len(symbols) == 0 {
		// シンボルが検出されない場合は、デモ用の簡単なコードを生成
		pythonCode := `# Grimoire Web Demo
print("Hello from Grimoire!")
print("魔法陣が正しく検出されませんでした。")
print("検出されたシンボル数: 0")
`
		debugInfo := map[string]interface{}{
			"symbolCount": 0,
			"symbols":     make([]interface{}, 0),
		}
		return createResultWithDebug(true, "", pythonCode, nil, debugInfo, "No symbols detected, showing demo code")
	}

	// デバッグ情報を作成
	symbolInfo := make([]interface{}, len(symbols))
	for i, sym := range symbols {
		info := map[string]interface{}{
			"type": string(sym.Type),
			"position": map[string]interface{}{
				"x": float64(sym.Position.X), // 明示的にfloat64に変換
				"y": float64(sym.Position.Y), // 明示的にfloat64に変換
			},
		}
		// Patternが空でない場合のみ設定
		if sym.Pattern != "" {
			info["pattern"] = string(sym.Pattern) // 明示的にstringに変換
		}
		symbolInfo[i] = info
	}
	debugInfo := map[string]interface{}{
		"symbolCount": float64(len(symbols)), // intをfloat64に変換
		"symbols":     symbolInfo,
	}

	// パース
	p := parser.NewParser()
	ast, err := p.Parse(symbols, connections)
	if err != nil {
		// パースエラーの場合もデモコードを返す
		pythonCode := fmt.Sprintf(`# Grimoire Web Demo
print("パースエラー: %s")
print("検出されたシンボル数: %d")
for i in range(5):
    print(f"カウント: {i}")
`, err.Error(), len(symbols))
		return createResultWithDebug(true, "", pythonCode, nil, debugInfo, "Parse error, showing demo code")
	}

	// コンパイル
	pythonCode, err := compiler.Compile(ast)
	if err != nil {
		// コンパイルエラーの場合もデモコードを返す
		pythonCode := fmt.Sprintf(`# Grimoire Web Demo
print("コンパイルエラー: %s")
print("Hello from Grimoire!")
`, err.Error())
		return createResultWithDebug(true, "", pythonCode, ast, debugInfo, "Compile error, showing demo code")
	}

	// 実行（WebAssemblyでは制限あり）
	output, err := executeInSandbox(pythonCode)
	if err != nil {
		return createResultWithDebug(true, output, pythonCode, ast, debugInfo, fmt.Sprintf("Code generated successfully, but execution is limited in browser: %v", err))
	}

	return createResultWithDebug(true, output, pythonCode, ast, debugInfo, "")
}

// validateCode はGrimoireコードを検証する
func validateCode(this js.Value, args []js.Value) interface{} {
	// WebAssembly版では簡易的な実装
	return createResult(true, "Validation is not implemented in WebAssembly version", "", "")
}

// formatCode はGrimoireコードをフォーマットする
func formatCode(this js.Value, args []js.Value) interface{} {
	// WebAssembly版では簡易的な実装
	return createResult(true, "Formatting is not implemented in WebAssembly version", "", "")
}

// executeInSandbox は制限された環境でPythonコードを実行する
func executeInSandbox(pythonCode string) (string, error) {
	// WebAssemblyでは直接Pythonを実行できないため、
	// JavaScriptのPythonインタープリタ（Pyodideなど）を使用する必要がある
	// ここでは仮の実装として、コードを返すだけにする
	return "Python execution in browser requires Pyodide integration", nil
}

// createResult は成功結果を作成する
func createResult(success bool, output, code, warning string) map[string]interface{} {
	result := map[string]interface{}{
		"success": success,
		"output":  output,
		"code":    code,
	}
	if warning != "" {
		result["warning"] = warning
	}
	return result
}

// createResultWithAST は成功結果をASTと共に作成する
func createResultWithAST(success bool, output, code string, ast interface{}, warning string) map[string]interface{} {
	result := map[string]interface{}{
		"success": success,
		"output":  output,
		"code":    code,
	}

	// astがnilでない場合のみ設定
	if ast != nil {
		// ASTを安全にJavaScript用に変換
		result["ast"] = toJSValue(ast)
	}

	if warning != "" {
		result["warning"] = warning
	}
	return result
}

// createResultWithDebug は成功結果をデバッグ情報と共に作成する
func createResultWithDebug(success bool, output, code string, ast interface{}, debugInfo map[string]interface{}, warning string) map[string]interface{} {
	result := map[string]interface{}{
		"success": success,
		"output":  output,
		"code":    code,
	}

	// debugInfoが空でない場合のみ設定
	if debugInfo != nil && len(debugInfo) > 0 {
		result["debug"] = debugInfo
	}

	// astがnilでない場合のみ設定
	if ast != nil {
		// ASTを安全にJavaScript用に変換
		result["ast"] = toJSValue(ast)
	}

	if warning != "" {
		result["warning"] = warning
	}
	return result
}

// createErrorResult はエラー結果を作成する
func createErrorResult(errorMsg string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error":   errorMsg,
	}
}

// toJSValue は任意のGo値をJavaScriptに安全に変換する
func toJSValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)

	switch val.Kind() {
	case reflect.Invalid:
		return nil
	case reflect.Bool:
		return val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.String:
		return val.String()
	case reflect.Slice, reflect.Array:
		arr := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			arr[i] = toJSValue(val.Index(i).Interface())
		}
		return arr
	case reflect.Map:
		m := make(map[string]interface{})
		for _, key := range val.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			valueConverted := toJSValue(val.MapIndex(key).Interface())
			// 値がnilでないことを確認
			if valueConverted != nil {
				m[keyStr] = valueConverted
			}
		}
		return m
	case reflect.Struct:
		return structToMap(val)
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		return toJSValue(val.Elem().Interface())
	case reflect.Interface:
		if val.IsNil() {
			return nil
		}
		return toJSValue(val.Elem().Interface())
	default:
		// その他の型は文字列として返す
		return fmt.Sprintf("%v", v)
	}
}

// structToMap は構造体をマップに変換する
func structToMap(v reflect.Value) map[string]interface{} {
	result := make(map[string]interface{})
	t := v.Type()

	// 型名を最初に追加（すべての構造体に対して）
	if t.Name() != "" {
		result["_type"] = t.Name()
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			// プライベートフィールドはスキップ
			continue
		}

		fieldValue := v.Field(i)

		// フィールドの値を変換
		convertedValue := toJSValue(fieldValue.Interface())
		// 値がnilでないことを確認
		if convertedValue != nil {
			result[field.Name] = convertedValue
		}
	}

	return result
}
