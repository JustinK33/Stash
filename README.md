# Stash

Stash is a small macOS desktop app for saving text snippets and images that you want to copy again later.
Paste text or images, drag image files into the app, and copy saved content whenever you need it.

## Install

Download `Stash-macOS.zip` from the latest GitHub Release.
Unzip it, move `Stash.app` into Applications, then open it.
Users do not need Go, Homebrew, or a terminal to install the release build.
If macOS blocks the app because it is not notarized yet, run this once:

```bash
xattr -dr com.apple.quarantine /Applications/Stash.app
open /Applications/Stash.app
```

You can also right-click `Stash.app`, choose Open, then confirm Open.

## Use

Paste text into the input and click Save.
Click Copy on any saved snippet to put it back on your clipboard.
Open the Images tab and paste an image with `Command + V`, click Paste Image, or drag PNG, JPEG, and GIF files into the window.
Click Copy on a saved image to put the image back on your clipboard.
Text and images have separate delete and clear controls.
The default global shortcut is `Control + Option + 0`.
Click the shortcut button in the top-right corner to change it.
Choose one key and at least one modifier.
Use the global shortcut to hide or bring back the window.
Close the window or press `Command + Q` to quit Stash.

## Build

Install Go, then build the app binary.

```bash
go build ./cmd/stash
```

Create a macOS app bundle and zip.

```bash
bash scripts/package-macos.sh build/Stash.app Stash-macOS.zip
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

Run the automated Go tests.

```bash
go test ./...
```

## Downloadable Build

The CI workflow builds the app on macOS, runs tests, packages `Stash.app`, and uploads `Stash-macOS.zip` as a workflow artifact.
The release workflow publishes `Stash-macOS.zip` to GitHub Releases when a version tag is pushed.

Create a release by pushing `main`, then pushing a version tag.

```bash
git push origin main
git tag v1.0.0
git push origin v1.0.0
```

The current release build is ad-hoc signed but not notarized.
The most professional macOS distribution path is a Developer ID signed and notarized DMG.
That requires an Apple Developer account and signing credentials stored as GitHub Actions secrets.

## Data

Saved metadata is stored at `~/Library/Application Support/Stash/stash.json`.
Managed image files are stored under `~/Library/Application Support/Stash/images/`.
Stash also migrates snippets from the previous app data file when available.

## What I Learned

Small desktop apps still need careful packaging.
The earlier Qt build could work locally but fail after installation because the bundle loaded two different Qt copies at runtime.
Moving to Go simplified distribution because the main binary is self-contained and the macOS app bundle is easier to inspect.

Global shortcuts should use the platform shortcut API when possible.
Listening to every key event is more fragile and can require extra permissions.
For this app, a Carbon hotkey for `Control + Option + 0` is simpler and more reliable than an `fn` shortcut.

Persistence should be boring and easy to migrate.
The app stores snippets in a small JSON file and keeps migration code for the old data path so existing saved text is not lost.

Simple UI is a product feature here.
The useful workflow is only paste, save, copy, delete, and clear.
Anything beyond that made the app feel heavier than the job it needed to do.
