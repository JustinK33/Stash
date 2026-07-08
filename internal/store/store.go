package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"stash/internal/keybind"
)

const (
	AppName       = "Stash"
	FileName      = "stash.json"
	ImageDirName  = "images"
	MaxImageBytes = 100 * 1024 * 1024
)

type Snippet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

type File struct {
	Snippets []Snippet `json:"snippets"`
	Images   []Image   `json:"images,omitempty"`
	Settings Settings  `json:"settings"`
}

type Image struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	FileName  string    `json:"fileName"`
	MediaType string    `json:"mediaType"`
	Size      int64     `json:"size"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	CreatedAt time.Time `json:"createdAt"`
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
	file.Images = normalizeImages(file.Images)
	file.Settings = normalizeSettings(file.Settings)
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(store.path, append(data, '\n'), 0o644)
}

func (store *Store) ImportImage(file *File, reader io.Reader, originalName string) (Image, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxImageBytes+1))
	if err != nil {
		return Image{}, err
	}
	if len(data) > MaxImageBytes {
		return Image{}, fmt.Errorf("image exceeds the %d MB limit", MaxImageBytes/(1024*1024))
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return Image{}, errors.New("file is not a supported PNG, JPEG, or GIF image")
	}

	mediaType, extension := imageFormat(format)
	if mediaType == "" {
		return Image{}, errors.New("file is not a supported PNG, JPEG, or GIF image")
	}

	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	for index, savedImage := range file.Images {
		if savedImage.ID == digest {
			file.Images = append([]Image{savedImage}, slices.Delete(slices.Clone(file.Images), index, index+1)...)
			if err := store.Save(*file); err != nil {
				return Image{}, err
			}
			return savedImage, nil
		}
	}

	savedImage := Image{
		ID:        digest,
		Name:      cleanImageName(originalName, extension),
		FileName:  digest + extension,
		MediaType: mediaType,
		Size:      int64(len(data)),
		Width:     config.Width,
		Height:    config.Height,
		CreatedAt: time.Now(),
	}

	imageDir := filepath.Join(filepath.Dir(store.path), ImageDirName)
	if err := os.MkdirAll(imageDir, 0o755); err != nil {
		return Image{}, err
	}
	imagePath := store.ImagePath(savedImage)
	temporaryPath := imagePath + ".tmp"
	if err := os.WriteFile(temporaryPath, data, 0o644); err != nil {
		return Image{}, err
	}
	if err := os.Rename(temporaryPath, imagePath); err != nil {
		_ = os.Remove(temporaryPath)
		return Image{}, err
	}

	file.Images = append([]Image{savedImage}, file.Images...)
	if err := store.Save(*file); err != nil {
		file.Images = file.Images[1:]
		_ = os.Remove(imagePath)
		return Image{}, err
	}
	return savedImage, nil
}

func (store *Store) DeleteImage(file *File, id string) error {
	var deleted Image
	found := false
	remaining := make([]Image, 0, len(file.Images))
	for _, savedImage := range file.Images {
		if savedImage.ID == id {
			deleted = savedImage
			found = true
			continue
		}
		remaining = append(remaining, savedImage)
	}
	if !found {
		return nil
	}

	previous := file.Images
	file.Images = remaining
	if err := store.Save(*file); err != nil {
		file.Images = previous
		return err
	}
	if err := os.Remove(store.ImagePath(deleted)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (store *Store) ClearImages(file *File) error {
	images := slices.Clone(file.Images)
	file.Images = nil
	if err := store.Save(*file); err != nil {
		file.Images = images
		return err
	}

	var cleanupError error
	for _, savedImage := range images {
		if err := os.Remove(store.ImagePath(savedImage)); err != nil && !errors.Is(err, os.ErrNotExist) {
			cleanupError = errors.Join(cleanupError, err)
		}
	}
	return cleanupError
}

func (store *Store) ImagePath(savedImage Image) string {
	return filepath.Join(filepath.Dir(store.path), ImageDirName, filepath.Base(savedImage.FileName))
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

	var fields map[string]json.RawMessage
	_ = json.Unmarshal(data, &fields)
	var file File
	if err := json.Unmarshal(data, &file); err == nil && fields["snippets"] != nil {
		file.Snippets = normalize(file.Snippets)
		file.Images = normalizeImages(file.Images)
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

func normalizeImages(images []Image) []Image {
	normalized := make([]Image, 0, len(images))
	seen := map[string]bool{}
	for _, savedImage := range images {
		if savedImage.ID == "" || savedImage.FileName == "" || seen[savedImage.ID] {
			continue
		}
		savedImage.FileName = filepath.Base(savedImage.FileName)
		if savedImage.FileName == "." {
			continue
		}
		seen[savedImage.ID] = true
		normalized = append(normalized, savedImage)
	}
	return normalized
}

func imageFormat(format string) (mediaType string, extension string) {
	switch format {
	case "png":
		return "image/png", ".png"
	case "jpeg":
		return "image/jpeg", ".jpg"
	case "gif":
		return "image/gif", ".gif"
	default:
		return "", ""
	}
}

func cleanImageName(name string, extension string) string {
	name = strings.TrimSpace(filepath.Base(name))
	if name == "" || name == "." {
		return "Pasted image" + extension
	}
	return name
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
