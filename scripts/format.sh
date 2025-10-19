#!/bin/bash
set -e

echo "🔍 Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed!"
    exit 1
fi

echo "💅 Formatting all Go files with gofmt -s..."
find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*" -not -path "./examples/*/vendor/*" | while read -r file; do
    echo "  ✓ $file"
    gofmt -s -w "$file"
done

echo ""
echo "✅ All files formatted successfully!"
echo ""
echo "Run 'go vet ./...' to check for other issues"

