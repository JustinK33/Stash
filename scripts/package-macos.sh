#!/usr/bin/env bash
set -euo pipefail

APP_PATH="${1:-build/QuickNote.app}"
ZIP_PATH="${2:-QuickNote-macOS.zip}"
QT_PREFIX="${QT_PREFIX:-$(brew --prefix qt)}"
QTBASE_PREFIX="${QTBASE_PREFIX:-$(brew --prefix qtbase)}"
QT_PLUGINS_DIR="${QT_PREFIX}/share/qt/plugins"
APP_BINARY="${APP_PATH}/Contents/MacOS/QuickNote"

rewrite_dependency() {
  local binary="$1"
  local old_path="$2"
  local new_path="$3"

  if otool -L "${binary}" | grep -Fq "${old_path}"; then
    install_name_tool -change "${old_path}" "${new_path}" "${binary}"
  fi
}

rewrite_qt_dependencies() {
  local binary="$1"

  for framework in QtCore QtGui QtWidgets; do
    local bundled_path="@executable_path/../Frameworks/${framework}.framework/Versions/A/${framework}"
    rewrite_dependency "${binary}" \
      "@rpath/${framework}.framework/Versions/A/${framework}" \
      "${bundled_path}"
    rewrite_dependency "${binary}" \
      "${QT_PREFIX}/lib/${framework}.framework/Versions/A/${framework}" \
      "${bundled_path}"
    rewrite_dependency "${binary}" \
      "${QTBASE_PREFIX}/lib/${framework}.framework/Versions/A/${framework}" \
      "${bundled_path}"
  done
}

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
  rewrite_qt_dependencies "${plugin}"
done

rewrite_qt_dependencies "${APP_BINARY}"

codesign --force --deep --sign - "${APP_PATH}"
codesign --verify --deep --strict "${APP_PATH}"
ditto -c -k --sequesterRsrc --keepParent "${APP_PATH}" "${ZIP_PATH}"
