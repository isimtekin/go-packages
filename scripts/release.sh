#!/bin/bash
# scripts/release.sh

PACKAGE=$1
VERSION=$2

if [ -z "$PACKAGE" ] || [ -z "$VERSION" ]; then
    echo "Usage: ./scripts/release.sh <package-name> <version>"
    echo "Example: ./scripts/release.sh mongo-client 1.0.0"
    exit 1
fi

# Check if package exists
if [ ! -d "$PACKAGE" ]; then
    echo "Package $PACKAGE does not exist"
    exit 1
fi

# Create tag
TAG="${PACKAGE}/v${VERSION}"

# Check if tag already exists
if git rev-parse "$TAG" >/dev/null 2>&1; then
    echo "Tag $TAG already exists"
    exit 1
fi

# Run tests
echo "Running tests for $PACKAGE..."
cd "$PACKAGE"
go test ./...
if [ $? -ne 0 ]; then
    echo "Tests failed, aborting release"
    exit 1
fi
cd ..

# Create and push tag
echo "Creating tag $TAG..."
git tag -a "$TAG" -m "Release ${PACKAGE} v${VERSION}"
git push origin "$TAG"

echo "Successfully released ${PACKAGE} v${VERSION}"
echo "Users can now import with:"
echo "go get github.com/finbyte/go-packages/${PACKAGE}@v${VERSION}"