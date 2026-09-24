package store

import (
	"os"
	"path/filepath"
	"testing"

	"stash/internal/keybind"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	stashStore := newTestStore(t)
	snippets := Add(Add(nil, "second"), "first")
	settings := Settings{Shortcut: keybind.Binding{Key: "S", Command: true}}

	if err := stashStore.Save(File{Snippets: snippets, Settings: settings}); err != nil {
		t.Fatal(err)
	}

	loaded, err := stashStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Snippets) != 2 {
		t.Fatalf("got %d snippets, want 2: %#v", len(loaded.Snippets), loaded.Snippets)
	}
	if loaded.Snippets[0].Text != "first" || loaded.Snippets[1].Text != "second" {
		t.Errorf("got %#v, want order [first second]", loaded.Snippets)
	}
	if loaded.Settings.Shortcut != settings.Shortcut {
		t.Errorf("got shortcut %+v, want %+v", loaded.Settings.Shortcut, settings.Shortcut)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	loaded, err := newTestStore(t).Load()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Snippets) != 0 {
		t.Errorf("got %d snippets, want 0", len(loaded.Snippets))
	}
	if loaded.Settings != DefaultSettings() {
		t.Errorf("got settings %+v, want %+v", loaded.Settings, DefaultSettings())
	}
}

func TestLoadMigratesLegacyClips(t *testing.T) {
	stashStore := newTestStore(t)
	configDir := filepath.Dir(filepath.Dir(stashStore.Path()))
	legacyPath := filepath.Join(configDir, "QuickNote", "quicknote.json")
	writeFile(t, legacyPath, `{
	  "clips": [
	    {"text": "legacy one", "captured": "2026-07-06T10:00:00-04:00"},
	    {"text": "legacy two", "captured": "2026-07-06T10:01:00-04:00"}
	  ]
	}`)

	loaded, err := stashStore.Load()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Snippets) != 2 {
		t.Fatalf("got %d migrated snippets, want 2: %#v", len(loaded.Snippets), loaded.Snippets)
	}
	if loaded.Snippets[0].Text != "legacy one" || loaded.Snippets[1].Text != "legacy two" {
		t.Errorf("got %#v, want order [legacy one, legacy two]", loaded.Snippets)
	}
	if loaded.Settings.Shortcut != keybind.Default() {
		t.Errorf("got shortcut %+v, want the default %+v", loaded.Settings.Shortcut, keybind.Default())
	}
	if _, err := os.Stat(stashStore.Path()); err != nil {
		t.Errorf("migrated file was not written to %s: %v", stashStore.Path(), err)
	}
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
