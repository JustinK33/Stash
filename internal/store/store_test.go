package store

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreSavesLoadsAndDeduplicatesSnippets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Stash", FileName)
	stashStore := NewAt(path)

	snippets := Add(nil, "first")
	snippets = Add(snippets, "second")
	snippets = Add(snippets, "first")

	if len(snippets) != 2 {
		t.Fatalf("expected 2 snippets, got %d", len(snippets))
	}
	if snippets[0].Text != "first" {
		t.Fatalf("expected duplicate to move to top, got %q", snippets[0].Text)
	}

	settings := DefaultSettings()
	settings.Shortcut.Key = "S"
	settings.Shortcut.Command = true
	settings.Shortcut.Control = false
	settings.Shortcut.Option = false

	if err := stashStore.Save(File{Snippets: snippets, Settings: settings}); err != nil {
		t.Fatal(err)
	}

	loadedFile, err := stashStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	loaded := loadedFile.Snippets
	if len(loaded) != 2 || loaded[0].Text != "first" || loaded[1].Text != "second" {
		t.Fatalf("unexpected snippets: %#v", loaded)
	}
	if loadedFile.Settings.Shortcut.Display() != "Command + S" {
		t.Fatalf("unexpected shortcut: %s", loadedFile.Settings.Shortcut.Display())
	}
}

func TestStoreMigratesLegacyClips(t *testing.T) {
	root := t.TempDir()
	currentPath := filepath.Join(root, "Stash", FileName)
	legacyPath := filepath.Join(root, "QuickNote", "quicknote.json")

	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPath, []byte(`{
	  "clips": [
	    {"text": "legacy one", "captured": "2026-07-06T10:00:00-04:00"},
	    {"text": "legacy two", "captured": "2026-07-06T10:01:00-04:00"}
	  ]
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	stashStore := NewAt(currentPath)
	loadedFile, err := stashStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	loaded := loadedFile.Snippets
	if len(loaded) != 2 || loaded[0].Text != "legacy one" || loaded[1].Text != "legacy two" {
		t.Fatalf("unexpected migrated snippets: %#v", loaded)
	}
	if loadedFile.Settings.Shortcut.Display() != "Control + Option + 0" {
		t.Fatalf("unexpected migrated shortcut: %s", loadedFile.Settings.Shortcut.Display())
	}
	if _, err := os.Stat(currentPath); err != nil {
		t.Fatalf("expected migrated file to be saved: %v", err)
	}
}

func TestStoreImportsLoadsAndDeletesManagedImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Stash", FileName)
	stashStore := NewAt(path)
	file := File{Settings: DefaultSettings()}
	imageData := testPNG(t)

	savedImage, err := stashStore.ImportImage(&file, bytes.NewReader(imageData), "example.png")
	if err != nil {
		t.Fatal(err)
	}
	if savedImage.Name != "example.png" || savedImage.MediaType != "image/png" {
		t.Fatalf("unexpected image metadata: %#v", savedImage)
	}
	if len(file.Images) != 1 {
		t.Fatalf("expected one image, got %d", len(file.Images))
	}
	if _, err := os.Stat(stashStore.ImagePath(savedImage)); err != nil {
		t.Fatalf("expected managed image file: %v", err)
	}

	loaded, err := stashStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Images) != 1 || loaded.Images[0].ID != savedImage.ID {
		t.Fatalf("unexpected loaded images: %#v", loaded.Images)
	}

	if err := stashStore.DeleteImage(&loaded, savedImage.ID); err != nil {
		t.Fatal(err)
	}
	if len(loaded.Images) != 0 {
		t.Fatalf("expected image metadata to be deleted: %#v", loaded.Images)
	}
	if _, err := os.Stat(stashStore.ImagePath(savedImage)); !os.IsNotExist(err) {
		t.Fatalf("expected managed image file to be deleted, got %v", err)
	}
}

func TestStoreDeduplicatesImagesByContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Stash", FileName)
	stashStore := NewAt(path)
	file := File{Settings: DefaultSettings()}
	imageData := testPNG(t)

	first, err := stashStore.ImportImage(&file, bytes.NewReader(imageData), "first.png")
	if err != nil {
		t.Fatal(err)
	}
	second, err := stashStore.ImportImage(&file, bytes.NewReader(imageData), "second.png")
	if err != nil {
		t.Fatal(err)
	}

	if first.ID != second.ID || len(file.Images) != 1 {
		t.Fatalf("expected duplicate to reuse existing image: %#v", file.Images)
	}
}

func TestStoreRejectsInvalidAndOversizedImages(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Stash", FileName)
	stashStore := NewAt(path)
	file := File{Settings: DefaultSettings()}

	if _, err := stashStore.ImportImage(&file, bytes.NewBufferString("not an image"), "bad.png"); err == nil {
		t.Fatal("expected invalid image to be rejected")
	}

	oversized := bytes.NewReader(make([]byte, MaxImageBytes+1))
	if _, err := stashStore.ImportImage(&file, oversized, "large.png"); err == nil {
		t.Fatal("expected oversized image to be rejected")
	}
}

func TestStoreClearImagesDoesNotDeleteText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Stash", FileName)
	stashStore := NewAt(path)
	file := File{
		Snippets: Add(nil, "keep me"),
		Settings: DefaultSettings(),
	}

	if _, err := stashStore.ImportImage(&file, bytes.NewReader(testPNG(t)), "example.png"); err != nil {
		t.Fatal(err)
	}
	if err := stashStore.ClearImages(&file); err != nil {
		t.Fatal(err)
	}

	if len(file.Images) != 0 {
		t.Fatalf("expected images to be cleared: %#v", file.Images)
	}
	if len(file.Snippets) != 1 || file.Snippets[0].Text != "keep me" {
		t.Fatalf("expected text snippets to remain: %#v", file.Snippets)
	}
}

func testPNG(t *testing.T) []byte {
	t.Helper()
	source := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	source.Set(0, 0, color.NRGBA{R: 255, A: 255})
	var data bytes.Buffer
	if err := png.Encode(&data, source); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}
