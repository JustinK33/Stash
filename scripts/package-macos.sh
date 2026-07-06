#!/usr/bin/env bash
set -euo pipefail

APP_PATH="${1:-build/QuickNote.app}"
ZIP_PATH="${2:-QuickNote-macOS.zip}"
QT_PREFIX="${QT_PREFIX:-$(brew --prefix qt)}"
QT_PLUGINS_DIR="${QT_PREFIX}/share/qt/plugins"

rm -rf "${APP_PATH}/Contents/PlugIns"
if [[ ! -d "${APP_PATH}/Contents/Frameworks/QtCore.framework" ]]; then
  "${QT_PREFIX}/bin/macdeployqt" "${APP_PATH}" -no-plugins
fi

mkdir -p "${APP_PATH}/Contents/PlugIns/platforms"
mkdir -p "${APP_PATH}/Contents/PlugIns/styles"
cp "${QT_PLUGINS_DIR}/platforms/libqcocoa.dylib" \
  "${APP_PATH}/Contents/PlugIns/platforms/"
cp "${QT_PLUGINS_DIR}/styles/libqmacstyle.dylib" \
  "${APP_PATH}/Contents/PlugIns/styles/"

for plugin in \
  "${APP_PATH}/Contents/PlugIns/platforms/libqcocoa.dylib" \
  "${APP_PATH}/Contents/PlugIns/styles/libqmacstyle.dylib"; do
  install_name_tool \
    -change "@rpath/QtCore.framework/Versions/A/QtCore" \
    "@executable_path/../Frameworks/QtCore.framework/Versions/A/QtCore" \
    -change "@rpath/QtGui.framework/Versions/A/QtGui" \
    "@executable_path/../Frameworks/QtGui.framework/Versions/A/QtGui" \
    -change "@rpath/QtWidgets.framework/Versions/A/QtWidgets" \
    "@executable_path/../Frameworks/QtWidgets.framework/Versions/A/QtWidgets" \
    "${plugin}"
done

codesign --force --deep --sign - "${APP_PATH}"
codesign --verify --deep --strict "${APP_PATH}"
ditto -c -k --sequesterRsrc --keepParent "${APP_PATH}" "${ZIP_PATH}"
