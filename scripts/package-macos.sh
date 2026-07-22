#!/usr/bin/env bash
set -euo pipefail

APP_PATH="${1:-build/Stash.app}"
ZIP_PATH="${2:-Stash-macOS.zip}"
APP_BINARY="${APP_PATH}/Contents/MacOS/Stash"
VERSION="${VERSION:-1.0.0}"
BUILD_NUMBER="${BUILD_NUMBER:-1}"

rm -rf "${APP_PATH}" "${ZIP_PATH}"
mkdir -p "${APP_PATH}/Contents/MacOS"
mkdir -p "${APP_PATH}/Contents/Resources"

go build -trimpath -ldflags="-s -w" -o "${APP_BINARY}" ./cmd/stash

cat > "${APP_PATH}/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleDevelopmentRegion</key>
  <string>en</string>
  <key>CFBundleExecutable</key>
  <string>Stash</string>
  <key>CFBundleIdentifier</key>
  <string>com.justink33.stash</string>
  <key>CFBundleInfoDictionaryVersion</key>
  <string>6.0</string>
  <key>CFBundleName</key>
  <string>Stash</string>
  <key>CFBundleIconFile</key>
  <string>Stash</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleShortVersionString</key>
  <string>${VERSION}</string>
  <key>CFBundleVersion</key>
  <string>${BUILD_NUMBER}</string>
  <key>LSMinimumSystemVersion</key>
  <string>11.0</string>
  <key>NSHighResolutionCapable</key>
  <true/>
</dict>
</plist>
PLIST

ICON_SRC="resources/icons/stash.png"
ICONSET="build/Stash.iconset"
rm -rf "${ICONSET}"
mkdir -p "${ICONSET}"
for size in 16 32 128 256 512; do
  sips -z "${size}" "${size}"         "${ICON_SRC}" --out "${ICONSET}/icon_${size}x${size}.png" >/dev/null
  sips -z "$((size*2))" "$((size*2))" "${ICON_SRC}" --out "${ICONSET}/icon_${size}x${size}@2x.png" >/dev/null
done
iconutil -c icns "${ICONSET}" -o "${APP_PATH}/Contents/Resources/Stash.icns"
rm -rf "${ICONSET}"

codesign --force --deep --sign - "${APP_PATH}"
codesign --verify --deep --strict "${APP_PATH}"
ditto -c -k --sequesterRsrc --keepParent "${APP_PATH}" "${ZIP_PATH}"
