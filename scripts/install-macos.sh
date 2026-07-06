#!/usr/bin/env bash
set -euo pipefail

APP_NAME="QuickNote.app"
BUILD_DIR="${BUILD_DIR:-build}"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/Applications}"
APP_PATH="${BUILD_DIR}/${APP_NAME}"
QT_PREFIX="${QT_PREFIX:-$(brew --prefix qt)}"

if [[ ! -d "${APP_PATH}" ]]; then
  cmake -S . -B "${BUILD_DIR}" -DCMAKE_PREFIX_PATH="${QT_PREFIX}"
  cmake --build "${BUILD_DIR}" --target QuickNote
fi

bash scripts/package-macos.sh "${APP_PATH}" QuickNote-macOS.zip

mkdir -p "${INSTALL_DIR}"
rm -rf "${INSTALL_DIR:?}/${APP_NAME}"
ditto "${APP_PATH}" "${INSTALL_DIR}/${APP_NAME}"
codesign --verify --deep --strict "${INSTALL_DIR}/${APP_NAME}"

if [[ -x "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister" ]]; then
  if ! "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister" \
    -f "${INSTALL_DIR}/${APP_NAME}" >/dev/null 2>&1; then
    echo "Launch Services registration skipped."
  fi
fi

echo "Installed ${INSTALL_DIR}/${APP_NAME}"
echo "Run it with: open \"${INSTALL_DIR}/${APP_NAME}\""
