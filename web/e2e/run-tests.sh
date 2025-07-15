#!/bin/bash

# Build WASM if needed
echo "Building WASM..."
(cd ../.. && GOOS=js GOARCH=wasm go build -o web/static/wasm/grimoire.wasm cmd/grimoire-wasm/main.go)

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
fi

# Run tests
echo "Running E2E tests..."
npm test