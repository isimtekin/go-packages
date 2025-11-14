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

# Update package README.md version badge
echo "Updating $PACKAGE/README.md with version v${VERSION}..."
if [ -f "$PACKAGE/README.md" ]; then
    # Update version badge if it exists
    sed -i.bak "s|Latest Version.*|Latest Version: v${VERSION}|g" "$PACKAGE/README.md"
    sed -i.bak "s|go get github.com/isimtekin/go-packages/${PACKAGE}.*|go get github.com/isimtekin/go-packages/${PACKAGE}@v${VERSION}|g" "$PACKAGE/README.md"
    rm -f "$PACKAGE/README.md.bak"
fi

# Update root README.md package list
echo "Updating root README.md package list..."
if [ -f "README.md" ]; then
    # Use awk for more reliable table update
    awk -v pkg="$PACKAGE" -v ver="v${VERSION}" '
    BEGIN { in_table = 0 }
    /^\| Package \| Description \| Latest Version \| Install \|/ { in_table = 1; print; next }
    /^\|[-]+\|[-]+\|[-]+\|[-]+\|/ && in_table { print; next }
    in_table && /^\|/ {
        if ($0 ~ "\\[" pkg "\\]") {
            # Extract the parts
            split($0, parts, "|")
            # Update version and install command
            parts[4] = " " ver " "
            parts[5] = " `go get github.com/isimtekin/go-packages/" pkg "@" ver "` "
            print "|" parts[2] "|" parts[3] "|" parts[4] "|" parts[5] "|"
            next
        }
    }
    /^$/ && in_table { in_table = 0 }
    { print }
    ' README.md > README.md.tmp && mv README.md.tmp README.md
fi

# Commit version updates
if [ -n "$(git status --porcelain)" ]; then
    echo "Committing version updates..."
    git add "$PACKAGE/README.md" README.md
    git commit -m "${PACKAGE}: update version to v${VERSION}"
    git push origin main || git push origin master
fi

# Create and push tag
echo "Creating tag $TAG..."
git tag -a "$TAG" -m "Release ${PACKAGE} v${VERSION}"
git push origin "$TAG"

echo "Successfully released ${PACKAGE} v${VERSION}"
echo "Users can now import with:"
echo "go get github.com/isimtekin/go-packages/${PACKAGE}@v${VERSION}"