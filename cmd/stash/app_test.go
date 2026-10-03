package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"

	"stash/internal/store"
)

func newTestApp(t *testing.T) (*stashApp, *store.Store) {
	t.Helper()
	app := test.NewTempApp(t)
	app.Settings().SetTheme(stashTheme{})
	window := test.NewTempWindow(t, nil)
	window.Resize(fyne.NewSize(980, 720))
	stashStore := store.NewAt(filepath.Join(t.TempDir(), "stash.json"))
	ui := newStashApp(window, stashStore, store.File{Settings: store.DefaultSettings()})
	return ui, stashStore
}

func snippetCards(ui *stashApp) []*card {
	var cards []*card
	for _, o := range ui.textList.Objects {
		if c, ok := o.(*card); ok {
			cards = append(cards, c)
		}
	}
	return cards
}

func TestTextFlow(t *testing.T) {
	ui, stashStore := newTestApp(t)

	test.Type(ui.input, "hello stash")
	if ui.charCount.Text != "11 characters" {
		t.Fatalf("char count = %q", ui.charCount.Text)
	}

	ui.input.TypedShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyReturn, Modifier: fyne.KeyModifierShortcutDefault})
	if ui.input.Text != "" || len(ui.snippets) != 1 || ui.textCount.Text != "1 saved" {
		t.Fatalf("command enter did not save: input=%q snippets=%d count=%q", ui.input.Text, len(ui.snippets), ui.textCount.Text)
	}
	if file, _ := stashStore.Load(); len(file.Snippets) != 1 || file.Snippets[0].Text != "hello stash" {
		t.Fatalf("snippet not persisted: %+v", file.Snippets)
	}

	cards := snippetCards(ui)
	if len(cards) != 1 {
		t.Fatalf("expected one card, got %d", len(cards))
	}
	test.Tap(cards[0].children[0])
	if got := fyne.CurrentApp().Clipboard().Content(); got != "hello stash" {
		t.Fatalf("clipboard = %q", got)
	}

	ui.selectPage(pageImages)
	ui.search.SetText("nothing like it")
	if ui.page != pageText || ui.textCount.Text != "0 of 1" || len(snippetCards(ui)) != 0 {
		t.Fatalf("search miss: page=%v count=%q", ui.page, ui.textCount.Text)
	}
	ui.search.SetText("STASH")
	if ui.textCount.Text != "1 of 1" || len(snippetCards(ui)) != 1 {
		t.Fatalf("search hit: count=%q", ui.textCount.Text)
	}
	ui.search.SetText("")

	test.Tap(snippetCards(ui)[0].children[1])
	if len(ui.snippets) != 0 {
		t.Fatalf("delete left %d snippets", len(ui.snippets))
	}
	if file, _ := stashStore.Load(); len(file.Snippets) != 0 {
		t.Fatalf("delete not persisted: %+v", file.Snippets)
	}
}

func TestLongSnippetExpands(t *testing.T) {
	ui, _ := newTestApp(t)
	long := strings.Repeat("line\n", 10)
	ui.input.SetText(long)
	ui.saveSnippet()

	id := ui.snippets[0].ID
	if ui.expanded[id] {
		t.Fatal("should start collapsed")
	}
	link := findLink(snippetCards(ui)[0])
	if link == nil || link.Text != "Show more" {
		t.Fatalf("expected a Show more link, got %+v", link)
	}
	test.TapAt(link, fyne.NewPos(link.MinSize().Width/2, link.MinSize().Height/2))
	if !ui.expanded[id] || findLink(snippetCards(ui)[0]).Text != "Show less" {
		t.Fatal("tapping Show more should expand the snippet")
	}
}

func findLink(c *card) *hoverLink {
	var found *hoverLink
	var walk func(fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if l, ok := o.(*hoverLink); ok {
			found = l
		}
		if cont, ok := o.(*fyne.Container); ok {
			for _, child := range cont.Objects {
				walk(child)
			}
		}
	}
	walk(c.content)
	return found
}

func TestImageFlow(t *testing.T) {
	ui, stashStore := newTestApp(t)

	var data bytes.Buffer
	img := image.NewNRGBA(image.Rect(0, 0, 4, 3))
	img.Set(0, 0, color.NRGBA{R: 255, A: 255})
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	if !ui.importImage(&data, "dot.png") {
		t.Fatal("import failed")
	}
	if ui.imageCount.Text != "1 saved" {
		t.Fatalf("image count = %q", ui.imageCount.Text)
	}
	path := stashStore.ImagePath(ui.images[0])
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("image file missing: %v", err)
	}

	ui.clearImages()
	if len(ui.images) != 0 || ui.imageCount.Text != "0 saved" {
		t.Fatalf("clear left %d images", len(ui.images))
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("image file should be removed, got %v", err)
	}
}

func TestPagesAndCompactSidebar(t *testing.T) {
	ui, _ := newTestApp(t)

	ui.selectPage(pageImages)
	if ui.pages[pageText].Visible() || !ui.pages[pageImages].Visible() || !ui.navImages.selected || ui.navText.selected {
		t.Fatal("images page should be the only visible page")
	}

	ui.window.Resize(fyne.NewSize(600, 600))
	if !ui.sidebar.compact || ui.sidebar.width() != sidebarCompactWidth {
		t.Fatal("narrow window should collapse the sidebar")
	}
	ui.focusSearch()
	if ui.sidebar.width() != sidebarWidth {
		t.Fatal("searching should expand the sidebar")
	}
	ui.window.Resize(fyne.NewSize(1200, 700))
	if ui.sidebar.compact {
		t.Fatal("wide window should show the full sidebar")
	}
}
