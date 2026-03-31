#!/bin/bash
# Build script for cross-compiling soverstack launcher
# Builds binaries for Windows, Linux, and macOS

set -e

# Version can be passed as first argument, defaults to "dev"
VERSION=${1:-"dev"}
OUTPUT_DIR="./dist"

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo "Building soverstack launcher v$VERSION"
echo "========================================"

# Platforms to build for
PLATFORMS=(
    "windows/amd64"
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    # Split platform into OS and architecture
    IFS="/" read -r GOOS GOARCH <<< "$PLATFORM"

    # Determine output filename
    OUTPUT_NAME="soverstack-$GOOS-$GOARCH"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="$OUTPUT_NAME.exe"
    fi

    echo ""
    echo "Building for $GOOS/$GOARCH..."

    # Build with optimizations
    # -s: Omit symbol table
    # -w: Omit DWARF debugging information
    # These reduce binary size significantly
    env GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-s -w -X main.Version=$VERSION" \
        -o "$OUTPUT_DIR/$OUTPUT_NAME" \
        .

    if [ $? -eq 0 ]; then
        # Get file size
        if [ "$GOOS" = "windows" ]; then
            SIZE=$(stat -c%s "$OUTPUT_DIR/$OUTPUT_NAME" 2>/dev/null || stat -f%z "$OUTPUT_DIR/$OUTPUT_NAME" 2>/dev/null || echo "unknown")
        else
            SIZE=$(stat -c%s "$OUTPUT_DIR/$OUTPUT_NAME" 2>/dev/null || stat -f%z "$OUTPUT_DIR/$OUTPUT_NAME" 2>/dev/null || echo "unknown")
        fi

        # Convert size to human readable format
        if [ "$SIZE" != "unknown" ]; then
            SIZE_MB=$(echo "scale=2; $SIZE / 1024 / 1024" | bc)
            echo "✓ Built: $OUTPUT_DIR/$OUTPUT_NAME (${SIZE_MB}MB)"
        else
            echo "✓ Built: $OUTPUT_DIR/$OUTPUT_NAME"
        fi
    else
        echo "✗ Failed to build for $GOOS/$GOARCH"
        exit 1
    fi
done

echo ""
echo "========================================"
echo "Build complete!"
echo ""
echo "Binaries are available in: $OUTPUT_DIR/"
ls -lh "$OUTPUT_DIR/"
echo ""
echo "To test locally:"
echo "  $OUTPUT_DIR/soverstack-$(uname | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') --version"
