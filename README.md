# Stash

Stash is a small macOS desktop app for saving text snippets that you want to copy again later.
Paste text into the input, save it, copy it whenever you need it, and delete or clear saved snippets when they are no longer useful.

## Build

Install Qt 6, then build with CMake.

```bash
brew install qt
cmake -S . -B build -DCMAKE_PREFIX_PATH="$(brew --prefix qt)"
cmake --build build
```

Run the app from the generated bundle.

```bash
open build/Stash.app
```

## Install From Terminal

Install Stash into `~/Applications` from the terminal.

```bash
bash scripts/install-macos.sh
```

Run the installed app.

```bash
open ~/Applications/Stash.app
```

## Test

Run the automated Qt tests after building.

```bash
ctest --test-dir build --output-on-failure
```

## Shortcut

The default global shortcut is `Control + Option + 0`.
Click the shortcut pill in the app to set your own shortcut.
Existing installs that still have the old `fn + 0` default are migrated to `Control + Option + 0`.
Shortcuts that include `fn` use macOS Accessibility permissions.
If an `fn` shortcut does not work, enable Stash in System Settings > Privacy and Security > Accessibility.

## Downloadable Build

The GitHub Actions workflow builds the app on macOS, packages `Stash.app`, and uploads `Stash-macOS.zip` as a workflow artifact.
Download that artifact from a successful workflow run.
You can create the same zip locally after building.

```bash
bash scripts/package-macos.sh build/Stash.app Stash-macOS.zip
```

## Data

Saved snippets are stored at `~/Library/Application Support/Stash/stash.json`.
