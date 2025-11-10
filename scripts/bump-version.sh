#!/bin/bash
# scripts/bump-version.sh

PACKAGE=$1
BUMP_TYPE=$2 # major, minor, patch

# Get latest version
LATEST_TAG=$(git tag -l "${PACKAGE}/v*" | sort -V | tail -1)

if [ -z "$LATEST_TAG" ]; then
    echo "No previous version found for $PACKAGE, starting with v0.0.0"
    CURRENT_VERSION="0.0.0"
else
    CURRENT_VERSION=${LATEST_TAG#${PACKAGE}/v}
fi

# Parse version
IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

# Bump version
case $BUMP_TYPE in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "Invalid bump type. Use: major, minor, or patch"
        exit 1
        ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
echo "Bumping $PACKAGE from v$CURRENT_VERSION to v$NEW_VERSION"

# Run release script
./scripts/release.sh "$PACKAGE" "$NEW_VERSION"