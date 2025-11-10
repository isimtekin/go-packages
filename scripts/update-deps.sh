#!/bin/bash
# scripts/update-deps.sh

# Update all package dependencies
for dir in $(find . -name go.mod -not -path "./scripts/*" -exec dirname {} \;); do
    echo "Updating dependencies for $dir..."
    (cd "$dir" && go get -u ./... && go mod tidy)
done