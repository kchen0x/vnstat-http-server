#!/bin/bash
# Script to create GitHub Release and upload binaries
# Usage: GITHUB_TOKEN=your_token ./scripts/create-release.sh

set -e

REPO="kchen0x/vnstat-http-server"
TAG="v0.1"
RELEASE_NAME="v0.1"
RELEASE_NOTES="## v0.1 - Initial Release

### Features
- Zero-dependency single binary
- Token-based authentication
- CORS support for all endpoints
- Multiple vnstat endpoints (JSON, summary, daily, hourly, weekly, yearly, top, oneline)
- iOS Scriptable widget support
- Full internationalization (English/Chinese)

### Downloads
- Linux amd64: \`vnstat-http-server-linux-amd64\`
- Linux arm64: \`vnstat-http-server-linux-arm64\`

### Installation
1. Download the appropriate binary for your system
2. Make it executable: \`chmod +x vnstat-http-server-linux-amd64\`
3. Run: \`./vnstat-http-server-linux-amd64 -port 8080\`"

# Check if GITHUB_TOKEN is set
if [ -z "$GITHUB_TOKEN" ]; then
  echo "Error: GITHUB_TOKEN environment variable is not set."
  echo ""
  echo "To create the release, you need a GitHub Personal Access Token."
  echo "1. Go to https://github.com/settings/tokens"
  echo "2. Generate a new token with 'repo' scope"
  echo "3. Run: GITHUB_TOKEN=your_token ./scripts/create-release.sh"
  echo ""
  echo "Alternatively, create the release manually:"
  echo "1. Go to https://github.com/$REPO/releases/new"
  echo "2. Select tag: $TAG"
  echo "3. Title: $RELEASE_NAME"
  echo "4. Description: (copy from above)"
  echo "5. Upload binaries from bin/ directory"
  exit 1
fi

# Check if binaries exist
if [ ! -f "bin/vnstat-http-server-linux-amd64" ] || [ ! -f "bin/vnstat-http-server-linux-arm64" ]; then
  echo "Error: Binary files not found. Please run 'make build' first."
  exit 1
fi

echo "Creating release $TAG..."

# Create release
RESPONSE=$(curl -s -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/$REPO/releases" \
  -d "{
    \"tag_name\": \"$TAG\",
    \"name\": \"$RELEASE_NAME\",
    \"body\": \"$RELEASE_NOTES\",
    \"draft\": false,
    \"prerelease\": false
  }")

# Check for errors
if echo "$RESPONSE" | grep -q '"message"'; then
  ERROR_MSG=$(echo "$RESPONSE" | grep -o '"message":"[^"]*"' | cut -d'"' -f4)
  echo "Error creating release: $ERROR_MSG"
  echo "Full response: $RESPONSE"
  exit 1
fi

RELEASE_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$RELEASE_ID" ]; then
  echo "Failed to create release. Response:"
  echo "$RESPONSE"
  exit 1
fi

echo "✓ Release created with ID: $RELEASE_ID"
echo "Uploading binaries..."

# Upload binaries
for file in bin/vnstat-http-server-linux-amd64 bin/vnstat-http-server-linux-arm64; do
  filename=$(basename "$file")
  echo "  Uploading $filename..."
  UPLOAD_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: application/octet-stream" \
    --data-binary "@$file" \
    "https://uploads.github.com/repos/$REPO/releases/$RELEASE_ID/assets?name=$filename")
  
  HTTP_CODE=$(echo "$UPLOAD_RESPONSE" | tail -1)
  if [ "$HTTP_CODE" = "201" ]; then
    echo "  ✓ $filename uploaded successfully"
  else
    echo "  ✗ Failed to upload $filename (HTTP $HTTP_CODE)"
    echo "$UPLOAD_RESPONSE" | head -1
  fi
done

echo ""
echo "Release created successfully!"
echo "View at: https://github.com/$REPO/releases/tag/$TAG"

