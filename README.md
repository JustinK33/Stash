# Stash

Stash is a small macOS desktop app for saving text snippets and images that you want to copy again later.

<!-- TODO: add a screenshot or gif of the app here -->

## What It Does

Paste text or an image, or drag an image file into the window, and Stash saves it for later.
Click Copy on any saved snippet or image to put it back on your clipboard, or use the global keyboard shortcut to hide and bring back the window.
The whole workflow is just paste, save, copy, delete, and clear - nothing more.

## Tech Stack

- Go
- macOS app bundle (Carbon hotkey API for global shortcuts)
- JSON file (local persistence)

## Install and Run

Download `Stash-macOS.zip` from the latest GitHub Release, unzip it, and move `Stash.app` into Applications.
If macOS blocks the app because it isn't notarized, run this once:

```bash
xattr -dr com.apple.quarantine /Applications/Stash.app
open /Applications/Stash.app
```

To build from source instead:

```bash
go build ./cmd/stash
bash scripts/package-macos.sh build/Stash.app Stash-macOS.zip
open build/Stash.app
```

## What I Learned

- **Packaging matters for small desktop apps** - the earlier Qt build could work locally but fail after install because the bundle loaded two different Qt copies at runtime; moving to Go made distribution simpler because the binary is self-contained.
- **Use the platform shortcut API for global shortcuts** - listening to every key event is more fragile and needs extra permissions; a Carbon hotkey for `Control + Option + 0` is simpler and more reliable.
- **Keep persistence boring and easy to migrate** - snippets are stored in a small JSON file, with migration code for the old data path so existing saved text isn't lost.
- **Simple UI is a feature here** - the workflow is only paste, save, copy, delete, and clear; anything beyond that made the app feel heavier than the job needed.
