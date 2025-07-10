#!/bin/bash
# Goでクロスプラットフォームビルドを行うスクリプト

# OpenCVを静的リンクするための環境変数
export CGO_ENABLED=1
export CGO_CXXFLAGS="-std=c++11"

echo "Building Grimoire for multiple platforms..."

# macOS (Intel)
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o dist/grimoire-darwin-amd64 \
    ./cmd/grimoire

# macOS (Apple Silicon)
echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build \
    -ldflags="-s -w" \
    -o dist/grimoire-darwin-arm64 \
    ./cmd/grimoire

# Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o dist/grimoire-linux-amd64 \
    ./cmd/grimoire

# Windows (要mingw-w64)
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build \
    -ldflags="-s -w -H windowsgui" \
    -o dist/grimoire-windows-amd64.exe \
    ./cmd/grimoire

echo "Build complete!"