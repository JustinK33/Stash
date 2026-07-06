package store

import (
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

	if err := stashStore.Save(snippets); err != nil {
		t.Fatal(err)
	}

	loaded, err := stashStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 || loaded[0].Text != "first" || loaded[1].Text != "second" {
		t.Fatalf("unexpected snippets: %#v", loaded)
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
	loaded, err := stashStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 || loaded[0].Text != "legacy one" || loaded[1].Text != "legacy two" {
		t.Fatalf("unexpected migrated snippets: %#v", loaded)
	}
	if _, err := os.Stat(currentPath); err != nil {
		t.Fatalf("expected migrated file to be saved: %v", err)
	}
}
