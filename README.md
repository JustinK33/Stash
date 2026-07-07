# Stash

Stash is a small macOS desktop app for saving text snippets that you want to copy again later.
Paste text into the input, save it, copy it whenever you need it, and delete or clear saved snippets when they are no longer useful.

## Install

Download `Stash-macOS.zip` from the latest GitHub Release.
Unzip it, move `Stash.app` into Applications, then open it.
If macOS blocks the app because it is not notarized yet, right-click `Stash.app`, choose Open, then confirm Open.

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

## Shortcut

The default global shortcut is `Control + Option + 0`.
Press it once to hide Stash.
Press it again to bring Stash back.
Click the shortcut button in the top-right corner to change it.
Choose one key and at least one modifier.

## Downloadable Build

The CI workflow builds the app on macOS, runs tests, packages `Stash.app`, and uploads `Stash-macOS.zip` as a workflow artifact.
The release workflow publishes `Stash-macOS.zip` to GitHub Releases when a version tag is pushed.

Create a release by pushing a tag.

```bash
git tag v1.0.0
git push origin v1.0.0
```

The current release build is ad-hoc signed but not notarized.
The most professional macOS distribution path is a Developer ID signed and notarized DMG.
That requires an Apple Developer account and signing credentials stored as GitHub Actions secrets.

## Data

Saved snippets are stored at `~/Library/Application Support/Stash/stash.json`.
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
