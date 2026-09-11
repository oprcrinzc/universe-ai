#!/bin/bash
set -e

echo "🪐 Building Gravity Mass Simulator 3D WebAssembly..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

mkdir -p web

# 1. Copy Go wasm_exec.js runtime
GOROOT="$(go env GOROOT)"
if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/lib/wasm/wasm_exec.js" web/wasm_exec.js
elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/misc/wasm/wasm_exec.js" web/wasm_exec.js
else
    echo "⚠️ Warning: wasm_exec.js not found in standard GOROOT location."
fi

# 2. Compile WebAssembly binary
echo "📦 Compiling Go engine to WebAssembly (web/gravitysim.wasm)..."
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/gravitysim.wasm ./web/engine

echo "✅ WebAssembly build complete! Binary size: $(du -h web/gravitysim.wasm | cut -f1)"
echo "🚀 To run the local web server:"
echo "   go run ./web/server.go"
echo "   or"
echo "   ./gravitysim --web"
