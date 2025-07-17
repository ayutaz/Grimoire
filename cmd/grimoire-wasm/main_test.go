//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"testing"
)

func TestCreateResultWithDebug(t *testing.T) {
	tests := []struct {
		name      string
		success   bool
		output    string
		code      string
		ast       interface{}
		debugInfo map[string]interface{}
		warning   string
		validate  func(t *testing.T, result map[string]interface{})
	}{
		{
			name:    "with debug info as object",
			success: true,
			output:  "test output",
			code:    "print('test')",
			ast: map[string]interface{}{
				"type": "Program",
				"body": []interface{}{},
			},
			debugInfo: map[string]interface{}{
				"symbolCount": float64(3),
				"symbols": []interface{}{
					map[string]interface{}{
						"type": "circle",
						"position": map[string]interface{}{
							"x": float64(10),
							"y": float64(20),
						},
					},
				},
			},
			warning: "test warning",
			validate: func(t *testing.T, result map[string]interface{}) {
				// debugInfoがオブジェクトとして返されることを確認
				debug, ok := result["debug"].(map[string]interface{})
				if !ok {
					t.Errorf("Expected debug to be map[string]interface{}, got %T", result["debug"])
					return
				}

				// symbolCountの確認（JavaScriptでは数値はfloat64として扱われる）
				symbolCount, ok := debug["symbolCount"].(float64)
				if !ok || symbolCount != 3 {
					t.Errorf("Expected symbolCount to be 3, got %v (type: %T)", debug["symbolCount"], debug["symbolCount"])
				}

				// symbolsの確認
				symbols, ok := debug["symbols"].([]interface{})
				if !ok || len(symbols) != 1 {
					t.Errorf("Expected symbols to be array with 1 element, got %v", debug["symbols"])
				}

				// ASTは一時的に無効化されているため、存在しないことを確認
				if _, hasAst := result["ast"]; hasAst {
					t.Error("Expected no ast field while AST is temporarily disabled")
				}
			},
		},
		{
			name:      "without debug info",
			success:   true,
			output:    "test output",
			code:      "print('test')",
			ast:       nil,
			debugInfo: nil,
			warning:   "",
			validate: func(t *testing.T, result map[string]interface{}) {
				if _, ok := result["debug"]; ok {
					t.Error("Expected no debug field when debugInfo is nil")
				}
				if _, ok := result["ast"]; ok {
					t.Error("Expected no ast field when ast is nil")
				}
				if _, ok := result["warning"]; ok {
					t.Error("Expected no warning field when warning is empty")
				}
			},
		},
		{
			name:      "empty debug info",
			success:   true,
			output:    "test output",
			code:      "print('test')",
			ast:       nil,
			debugInfo: map[string]interface{}{},
			warning:   "",
			validate: func(t *testing.T, result map[string]interface{}) {
				if _, ok := result["debug"]; ok {
					t.Error("Expected no debug field when debugInfo is empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createResultWithDebug(tt.success, tt.output, tt.code, tt.ast, tt.debugInfo, tt.warning)

			// 基本フィールドの確認
			if result["success"] != tt.success {
				t.Errorf("Expected success to be %v, got %v", tt.success, result["success"])
			}
			if result["output"] != tt.output {
				t.Errorf("Expected output to be %v, got %v", tt.output, result["output"])
			}
			if result["code"] != tt.code {
				t.Errorf("Expected code to be %v, got %v", tt.code, result["code"])
			}

			// カスタム検証
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestCreateResultWithAST(t *testing.T) {
	tests := []struct {
		name     string
		success  bool
		output   string
		code     string
		ast      interface{}
		warning  string
		validate func(t *testing.T, result map[string]interface{})
	}{
		{
			name:    "with AST as object",
			success: true,
			output:  "test output",
			code:    "print('test')",
			ast: map[string]interface{}{
				"type": "Program",
				"body": []interface{}{
					map[string]interface{}{
						"type":  "Print",
						"value": "test",
					},
				},
			},
			warning: "",
			validate: func(t *testing.T, result map[string]interface{}) {
				// ASTは一時的に無効化されているため、存在しないことを確認
				if _, hasAst := result["ast"]; hasAst {
					t.Error("Expected no ast field while AST is temporarily disabled")
				}
			},
		},
		{
			name:    "without AST",
			success: true,
			output:  "test output",
			code:    "print('test')",
			ast:     nil,
			warning: "",
			validate: func(t *testing.T, result map[string]interface{}) {
				if _, ok := result["ast"]; ok {
					t.Error("Expected no ast field when ast is nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createResultWithAST(tt.success, tt.output, tt.code, tt.ast, tt.warning)

			// 基本フィールドの確認
			if result["success"] != tt.success {
				t.Errorf("Expected success to be %v, got %v", tt.success, result["success"])
			}
			if result["output"] != tt.output {
				t.Errorf("Expected output to be %v, got %v", tt.output, result["output"])
			}
			if result["code"] != tt.code {
				t.Errorf("Expected code to be %v, got %v", tt.code, result["code"])
			}

			// カスタム検証
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// JSONシリアライゼーションが不要になったことを確認するテスト
func TestNoJSONMarshalling(t *testing.T) {
	// 複雑なオブジェクト構造を作成
	complexDebugInfo := map[string]interface{}{
		"symbolCount": float64(5),
		"symbols": []interface{}{
			map[string]interface{}{
				"type": "circle",
				"position": map[string]interface{}{
					"x": float64(10.5),
					"y": float64(20.3),
				},
				"pattern": "solid",
			},
			map[string]interface{}{
				"type": "square",
				"position": map[string]interface{}{
					"x": float64(30.7),
					"y": float64(40.2),
				},
			},
		},
		"connections": []interface{}{
			map[string]interface{}{
				"from": 0,
				"to":   1,
				"type": "arrow",
			},
		},
	}

	complexAST := map[string]interface{}{
		"type": "Program",
		"body": []interface{}{
			map[string]interface{}{
				"type": "Assignment",
				"left": map[string]interface{}{
					"type": "Identifier",
					"name": "x",
				},
				"right": map[string]interface{}{
					"type":  "Number",
					"value": 42,
				},
			},
		},
	}

	result := createResultWithDebug(true, "output", "code", complexAST, complexDebugInfo, "")

	// debugInfoが同じオブジェクトであることを確認（JSON変換されていない）
	debug, ok := result["debug"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected debug to be map[string]interface{}, got %T", result["debug"])
	}

	// ネストされたデータが正しくアクセスできることを確認
	symbols, ok := debug["symbols"].([]interface{})
	if !ok || len(symbols) != 2 {
		t.Fatalf("Expected symbols to be array with 2 elements, got %v", debug["symbols"])
	}

	firstSymbol, ok := symbols[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected first symbol to be map, got %T", symbols[0])
	}

	position, ok := firstSymbol["position"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected position to be map, got %T", firstSymbol["position"])
	}

	x, ok := position["x"].(float64)
	if !ok || x != 10.5 {
		t.Errorf("Expected x to be 10.5, got %v", position["x"])
	}

	// ASTは一時的に無効化されているため、存在しないことを確認
	if _, hasAst := result["ast"]; hasAst {
		t.Error("Expected no ast field while AST is temporarily disabled")
	}
}

// 以前の実装（JSON文字列として返す）との互換性をテストするための確認
func TestJSONCompatibility(t *testing.T) {
	debugInfo := map[string]interface{}{
		"symbolCount": float64(1),
		"symbols": []interface{}{
			map[string]interface{}{
				"type": "star",
				"position": map[string]interface{}{
					"x": float64(15),
					"y": float64(25),
				},
			},
		},
	}

	result := createResultWithDebug(true, "", "print('hello')", nil, debugInfo, "")

	// JavaScript側でJSON.parseする必要がないことを確認
	// （オブジェクトとして直接アクセス可能）
	debug := result["debug"].(map[string]interface{})

	// オブジェクトをJSONにマーシャルして、正しい構造であることを確認
	jsonBytes, err := json.Marshal(debug)
	if err != nil {
		t.Fatalf("Failed to marshal debug info: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal debug info: %v", err)
	}

	// 元のオブジェクトと同じ構造であることを確認
	if unmarshaled["symbolCount"].(float64) != 1 {
		t.Errorf("Expected symbolCount to be 1, got %v", unmarshaled["symbolCount"])
	}
}
