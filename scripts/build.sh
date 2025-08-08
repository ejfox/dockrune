#!/bin/bash
set -e

echo "🚀 Building dockrune..."

# Build Nuxt dashboard
echo "📦 Building dashboard..."
cd dashboard
npm install
npm run build
cd ..

# Build Go binary
echo "🔨 Building Go binary..."
go build -o dockrune ./cmd/dockrune

echo "✅ Build complete!"
echo ""
echo "To start dockrune:"
echo "  ./dockrune init    # First time setup"
echo "  ./dockrune serve   # Start server"