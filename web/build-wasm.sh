#!/bin/bash

# WebAssembly向けのビルドスクリプト
# Build script for WebAssembly

set -e

echo "Building Grimoire for WebAssembly..."

# スクリプトのディレクトリに移動
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# プロジェクトルートに移動してGo依存関係をダウンロード
echo "Downloading Go dependencies..."
cd ..
go mod download

# webディレクトリに戻る
cd web

# Go WebAssemblyサポートファイルをコピー
echo "Copying wasm_exec.js..."
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" static/

# wasmディレクトリを作成
mkdir -p static/wasm

# WebAssemblyバイナリをビルド
echo "Building WASM binary..."
GOOS=js GOARCH=wasm go build -o static/wasm/grimoire.wasm ../cmd/grimoire-wasm/main.go

echo "Build complete!"
echo "Files generated:"
echo "  - static/wasm/grimoire.wasm"
echo "  - static/wasm_exec.js"