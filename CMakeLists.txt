cmake_minimum_required(VERSION 3.20)
project(QuickNote LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

find_package(Qt6 REQUIRED COMPONENTS Widgets Gui Core Test)

set(SOURCES
  src/main.cpp
  src/MainWindow.cpp
  src/DataStore.cpp
  src/HotkeyManager.cpp
  src/widgets/NoteItemWidget.cpp
  src/widgets/ClipItemWidget.cpp
  src/widgets/KeybindPill.cpp
  resources/resources.qrc
)

qt_add_executable(QuickNote MACOSX_BUNDLE ${SOURCES})

target_include_directories(QuickNote PRIVATE src src/widgets)
target_link_libraries(QuickNote PRIVATE Qt6::Widgets Qt6::Gui Qt6::Core)

if(APPLE)
  target_link_libraries(QuickNote PRIVATE "-framework ApplicationServices")
endif()

enable_testing()

qt_add_executable(QuickNoteTests
  tests/QuickNoteTests.cpp
  src/MainWindow.cpp
  src/DataStore.cpp
  src/HotkeyManager.cpp
  src/widgets/ClipItemWidget.cpp
  src/widgets/KeybindPill.cpp
  resources/resources.qrc
)

target_include_directories(QuickNoteTests PRIVATE src src/widgets)
target_link_libraries(QuickNoteTests PRIVATE Qt6::Widgets Qt6::Gui Qt6::Core Qt6::Test)

if(APPLE)
  target_link_libraries(QuickNoteTests PRIVATE "-framework ApplicationServices")
endif()

add_test(NAME QuickNoteTests COMMAND QuickNoteTests)
