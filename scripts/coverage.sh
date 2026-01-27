#!/bin/bash
set -e

echo "Running tests with coverage..."
cd pkg
go test -coverprofile=coverage.out ./...

echo ""
echo "Generating HTML report..."
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out | tail -1

echo ""
echo "Coverage report generated: pkg/coverage.html"
echo "Open with: open pkg/coverage.html"
