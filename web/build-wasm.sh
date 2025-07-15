#!/bin/bash

# WebAssembly向けのビルドスクリプト
# Build script for WebAssembly

set -e

echo "Building Grimoire for WebAssembly..."

# wasmディレクトリに移動
cd "$(dirname "$0")"

# Go WebAssemblyサポートファイルをコピー
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" static/

# wasmディレクトリを作成
mkdir -p static/wasm

# WebAssemblyバイナリをビルド
GOOS=js GOARCH=wasm go build -o static/wasm/grimoire.wasm ../cmd/grimoire-wasm/main.go

echo "Build complete!"
echo "Files generated:"
echo "  - static/wasm/grimoire.wasm"
echo "  - static/wasm_exec.js"