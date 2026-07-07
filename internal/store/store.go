package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"stash/internal/keybind"
)

const (
	AppName  = "Stash"
	FileName = "stash.json"
)

type Snippet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

type File struct {
	Snippets []Snippet `json:"snippets"`
	Settings Settings  `json:"settings"`
}

type Settings struct {
	Shortcut keybind.Binding `json:"shortcut"`
}

type Store struct {
	path string
}

func New() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	return &Store{path: filepath.Join(configDir, AppName, FileName)}, nil
}

func NewAt(path string) *Store {
	return &Store{path: path}
}

func (store *Store) Path() string {
	return store.path
}

func DefaultSettings() Settings {
	return Settings{Shortcut: keybind.Default()}
}

func (store *Store) Load() (File, error) {
	file, err := store.loadFromPath(store.path)
	if err == nil {
		return file, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return File{}, err
	}

	legacyPaths := legacyPaths(store.path)
	for _, legacyPath := range legacyPaths {
		file, legacyErr := store.loadFromPath(legacyPath)
		if legacyErr == nil {
			_ = store.Save(file)
			return file, nil
		}
		if !errors.Is(legacyErr, os.ErrNotExist) {
			return File{}, legacyErr
		}
	}

	return File{Settings: DefaultSettings()}, nil
}

func (store *Store) Save(file File) error {
	if err := os.MkdirAll(filepath.Dir(store.path), 0o755); err != nil {
		return err
	}

	file.Snippets = normalize(file.Snippets)
	file.Settings = normalizeSettings(file.Settings)
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(store.path, append(data, '\n'), 0o644)
}

func Add(snippets []Snippet, text string) []Snippet {
	text = strings.TrimSpace(text)
	if text == "" {
		return snippets
	}

	snippets = Delete(snippets, text)
	return append([]Snippet{{
		ID:        time.Now().Format("20060102150405.000000000"),
		Text:      text,
		CreatedAt: time.Now(),
	}}, snippets...)
}

func Delete(snippets []Snippet, text string) []Snippet {
	return slices.DeleteFunc(slices.Clone(snippets), func(snippet Snippet) bool {
		return snippet.Text == text
	})
}

func (store *Store) loadFromPath(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}

	var file File
	if err := json.Unmarshal(data, &file); err == nil && file.Snippets != nil {
		file.Snippets = normalize(file.Snippets)
		file.Settings = normalizeSettings(file.Settings)
		return file, nil
	}

	var legacy legacyFile
	if err := json.Unmarshal(data, &legacy); err != nil {
		return File{}, err
	}

	snippets := make([]Snippet, 0, len(legacy.Clips))
	for index, clip := range legacy.Clips {
		createdAt := clip.Captured
		if createdAt.IsZero() {
			createdAt = time.Now()
		}
		snippets = append(snippets, Snippet{
			ID:        createdAt.Format("20060102150405.000000000") + "-" + string(rune('a'+index)),
			Text:      clip.Text,
			CreatedAt: createdAt,
		})
	}
	return File{
		Snippets: normalize(snippets),
		Settings: DefaultSettings(),
	}, nil
}

func normalize(snippets []Snippet) []Snippet {
	normalized := make([]Snippet, 0, len(snippets))
	seen := map[string]bool{}

	for _, snippet := range snippets {
		snippet.Text = strings.TrimSpace(snippet.Text)
		if snippet.Text == "" || seen[snippet.Text] {
			continue
		}
		if snippet.ID == "" {
			snippet.ID = snippet.CreatedAt.Format("20060102150405.000000000")
		}
		if snippet.CreatedAt.IsZero() {
			snippet.CreatedAt = time.Now()
		}
		seen[snippet.Text] = true
		normalized = append(normalized, snippet)
	}

	return normalized
}

func normalizeSettings(settings Settings) Settings {
	settings.Shortcut = settings.Shortcut.Normalize()
	if !settings.Shortcut.HasModifier() {
		settings.Shortcut = keybind.Default()
	}
	return settings
}

func legacyPaths(currentPath string) []string {
	configDir := filepath.Dir(filepath.Dir(currentPath))
	return []string{
		filepath.Join(configDir, "QuickNote", "quicknote.json"),
		filepath.Join(configDir, "QuickDraft", "quickdraft.json"),
	}
}

type legacyFile struct {
	Clips []legacyClip `json:"clips"`
}

type legacyClip struct {
	Text     string    `json:"text"`
	Captured time.Time `json:"captured"`
}
