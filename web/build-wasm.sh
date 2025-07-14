#!/bin/bash

# WebAssembly向けのビルドスクリプト
# Build script for WebAssembly

set -e

echo "Building Grimoire for WebAssembly..."

# wasmディレクトリに移動
cd "$(dirname "$0")"

# Go WebAssemblyサポートファイルをコピー
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" static/

# WebAssemblyバイナリをビルド
GOOS=js GOARCH=wasm go build -o wasm/grimoire.wasm ../cmd/grimoire-wasm/main.go

echo "Build complete!"
echo "Files generated:"
echo "  - wasm/grimoire.wasm"
echo "  - static/wasm_exec.js"