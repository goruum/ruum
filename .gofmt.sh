#!/bin/bash
# Script to format all Go files

find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*" | while read -r file; do
    echo "Formatting: $file"
    gofmt -s -w "$file"
done

echo "✅ All files formatted!"

