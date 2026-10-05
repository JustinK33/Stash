package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"stash/internal/appearance"
	"stash/internal/hotkey"
	"stash/internal/imageclipboard"
	"stash/internal/keybind"
	"stash/internal/store"
	"stash/resources"
)

type page int

const (
	pageText page = iota
	pageImages
)

type stashApp struct {
	window   fyne.Window
	store    *store.Store
	snippets []store.Snippet
	images   []store.Image
	settings store.Settings

	input         *saveEntry
	search        *widget.Entry
	charCount     *canvas.Text
	textList      *fyne.Container
	imageList     *fyne.Container
	textCount     *canvas.Text
	imageCount    *canvas.Text
	shortcutBadge *shortcutBadge
	sidebar       *sidebar
	navText       *navItem
	navImages     *navItem
	pages         map[page]fyne.CanvasObject
	page          page
	pageFade      *canvas.Rectangle
	fadeAnim      *fyne.Animation
	toaster       *toaster
	expanded      map[string]bool
	hidden        bool
}

func main() {
	stashStore, err := store.New()
	if err != nil {
		log.Fatal(err)
	}

	file, err := stashStore.Load()
	if err != nil {
		log.Fatal(err)
	}

	fyneApp := app.NewWithID("com.justink33.stash")
	fyneApp.Settings().SetTheme(stashTheme{})
	appIcon := fyne.NewStaticResource("stash.png", resources.AppIcon)
	fyneApp.SetIcon(appIcon)

	window := fyneApp.NewWindow("Stash")
	window.SetIcon(appIcon)
	window.Resize(fyne.NewSize(980, 720))

	ui := newStashApp(window, stashStore, file)
	// Windows has no Dock icon, so a hidden window would be unreachable
	// without the shortcut. The tray gives it a way back.
	if desk, ok := fyneApp.(desktop.App); ok && runtime.GOOS == "windows" {
		desk.SetSystemTrayIcon(appIcon)
		desk.SetSystemTrayMenu(fyne.NewMenu("Stash", fyne.NewMenuItem("Show Stash", ui.show)))
	}
	fyneApp.Lifecycle().SetOnStarted(appearance.UseLight)
	fyneApp.Lifecycle().SetOnEnteredForeground(func() {
		fyne.Do(ui.show)
	})

	if err := hotkey.Register(ui.settings.Shortcut, func() {
		fyne.Do(ui.toggle)
	}); err != nil {
		ui.toast("Shortcut unavailable")
	}
	defer hotkey.Unregister()

	window.ShowAndRun()
}

func newStashApp(window fyne.Window, stashStore *store.Store, file store.File) *stashApp {
	ui := &stashApp{
		window:     window,
		store:      stashStore,
		snippets:   file.Snippets,
		images:     file.Images,
		settings:   file.Settings,
		input:      newSaveEntry(),
		search:     widget.NewEntry(),
		charCount:  newText("", 11, colorMuted, false),
		textList:   container.New(layout.NewCustomPaddedVBoxLayout(8)),
		imageList:  container.NewVBox(),
		textCount:  newText("0 saved", 13, colorMuted, false),
		imageCount: newText("0 saved", 13, colorMuted, false),
		toaster:    newToaster(),
		expanded:   map[string]bool{},
	}
	ui.input.onSave = ui.saveSnippet

	window.SetContent(ui.build())
	window.SetOnDropped(ui.handleDropped)
	ui.addShortcuts()
	ui.refreshTextList()
	ui.refreshImageList()
	return ui
}

func (ui *stashApp) build() fyne.CanvasObject {
	ui.shortcutBadge = newShortcutBadge(ui.settings.Shortcut.Display(), ui.openShortcutDialog)

	ui.navText = newNavItem("Text", "text", func() { ui.selectPage(pageText) })
	ui.navImages = newNavItem("Images", "image", func() { ui.selectPage(pageImages) })
	settings := newNavItem("Settings", "settings", ui.openShortcutDialog)
	ui.sidebar = newSidebar(ui.search, []*navItem{ui.navText, ui.navImages}, settings, ui.focusSearch)

	ui.search.OnChanged = func(string) {
		if ui.page != pageText {
			ui.selectPage(pageText)
		}
		ui.refreshTextList()
	}
	ui.search.OnSubmitted = func(string) { ui.window.Canvas().Focus(ui.input) }

	ui.pages = map[page]fyne.CanvasObject{
		pageText:   ui.buildTextPage(),
		pageImages: ui.buildImagePage(),
	}
	ui.pages[pageImages].Hide()
	ui.pageFade = canvas.NewRectangle(colorBackground)
	ui.pageFade.Hide()
	ui.navText.setSelected(true)

	content := container.NewStack(ui.pages[pageText], ui.pages[pageImages], ui.pageFade)
	shell := container.New(&shellLayout{sidebar: ui.sidebar}, ui.sidebar.root, content)
	return container.NewStack(shell, ui.toaster.layer)
}

func (ui *stashApp) addShortcuts() {
	canvas := ui.window.Canvas()
	canvas.AddShortcut(&fyne.ShortcutPaste{}, func(fyne.Shortcut) {
		if ui.page == pageImages {
			ui.pasteImage()
		}
	})
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: fyne.KeyModifierShortcutDefault}, func(fyne.Shortcut) {
		ui.focusSearch()
	})
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.Key1, Modifier: fyne.KeyModifierShortcutDefault}, func(fyne.Shortcut) {
		ui.selectPage(pageText)
	})
	canvas.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.Key2, Modifier: fyne.KeyModifierShortcutDefault}, func(fyne.Shortcut) {
		ui.selectPage(pageImages)
	})
}

func (ui *stashApp) focusSearch() {
	ui.sidebar.setExpanded(true)
	ui.window.Canvas().Focus(ui.search)
}

func (ui *stashApp) selectPage(next page) {
	if ui.search.Text == "" {
		ui.sidebar.setExpanded(false)
	}
	if next == ui.page {
		return
	}
	ui.pages[ui.page].Hide()
	ui.page = next
	ui.pages[next].Show()
	ui.navText.setSelected(next == pageText)
	ui.navImages.setSelected(next == pageImages)

	// A quick fade from the background color softens the page swap.
	if ui.fadeAnim != nil {
		ui.fadeAnim.Stop()
	}
	ui.pageFade.Show()
	ui.fadeAnim = fyne.NewAnimation(150*time.Millisecond, func(progress float32) {
		ui.pageFade.FillColor = color.NRGBA{R: colorBackground.R, G: colorBackground.G, B: colorBackground.B, A: uint8(255 * (1 - progress))}
		ui.pageFade.Refresh()
		if progress == 1 {
			ui.pageFade.Hide()
		}
	})
	ui.fadeAnim.Curve = fyne.AnimationEaseOut
	ui.fadeAnim.Start()
}

func (ui *stashApp) toast(message string) {
	ui.toaster.show(message)
}

// saveEntry is the multiline input, where Command (Ctrl on Windows) + Enter saves.
type saveEntry struct {
	widget.Entry
	onSave func()
}

func newSaveEntry() *saveEntry {
	e := &saveEntry{}
	e.MultiLine = true
	e.Wrapping = fyne.TextWrapWord
	e.ExtendBaseWidget(e)
	return e
}

func (e *saveEntry) TypedShortcut(shortcut fyne.Shortcut) {
	if custom, ok := shortcut.(*desktop.CustomShortcut); ok && custom.Modifier == fyne.KeyModifierShortcutDefault &&
		(custom.KeyName == fyne.KeyReturn || custom.KeyName == fyne.KeyEnter) {
		e.onSave()
		return
	}
	e.Entry.TypedShortcut(shortcut)
}

func (ui *stashApp) saveSnippet() {
	text := strings.TrimSpace(ui.input.Text)
	if text == "" {
		ui.toast("Nothing to save")
		return
	}

	ui.snippets = store.Add(ui.snippets, text)
	ui.input.SetText("")
	if err := ui.save(); err != nil {
		return
	}
	ui.refreshTextList()
	ui.toast("Saved")
}

func (ui *stashApp) copySnippet(text string) {
	fyne.CurrentApp().Clipboard().SetContent(text)
	ui.toast("Copied to clipboard")
}

func (ui *stashApp) deleteSnippet(text string) {
	ui.snippets = store.Delete(ui.snippets, text)
	if err := ui.save(); err != nil {
		return
	}
	ui.refreshTextList()
	ui.toast("Deleted")
}

func (ui *stashApp) pasteImage() {
	data, err := imageclipboard.ReadPNG()
	if err != nil {
		ui.toast(err.Error())
		return
	}
	ui.importImage(bytes.NewReader(data), "Pasted image.png")
}

func (ui *stashApp) handleDropped(_ fyne.Position, uris []fyne.URI) {
	if len(uris) == 0 {
		return
	}
	imported := 0
	for _, uri := range uris {
		reader, err := storage.Reader(uri)
		if err != nil {
			ui.toast("Could not read " + uri.Name())
			continue
		}
		if ui.importImage(reader, uri.Name()) {
			imported++
		}
		_ = reader.Close()
	}
	if imported > 0 {
		ui.selectPage(pageImages)
		ui.toast(fmt.Sprintf("Imported %d image(s)", imported))
	}
}

func (ui *stashApp) importImage(reader io.Reader, name string) bool {
	file := ui.currentFile()
	_, err := ui.store.ImportImage(&file, reader, name)
	if err != nil {
		ui.toast(err.Error())
		return false
	}
	ui.images = file.Images
	ui.refreshImageList()
	ui.toast("Image saved")
	return true
}

func (ui *stashApp) copyImage(savedImage store.Image) bool {
	file, err := os.Open(ui.store.ImagePath(savedImage))
	if err != nil {
		ui.toast("Could not read saved image")
		return false
	}
	defer file.Close()

	decoded, _, err := image.Decode(file)
	if err != nil {
		ui.toast("Saved image is invalid")
		return false
	}
	var data bytes.Buffer
	if err := png.Encode(&data, decoded); err != nil {
		ui.toast("Could not prepare image")
		return false
	}
	if err := imageclipboard.WritePNG(data.Bytes()); err != nil {
		ui.toast(err.Error())
		return false
	}
	ui.toast("Image copied")
	return true
}

func (ui *stashApp) deleteImage(savedImage store.Image) {
	file := ui.currentFile()
	err := ui.store.DeleteImage(&file, savedImage.ID)
	ui.images = file.Images
	ui.refreshImageList()
	if err != nil {
		ui.toast("Image removed, but its file could not be deleted")
		return
	}
	ui.toast("Image deleted")
}

func (ui *stashApp) clearImages() {
	file := ui.currentFile()
	err := ui.store.ClearImages(&file)
	ui.images = file.Images
	ui.refreshImageList()
	if err != nil {
		ui.toast("Images cleared, but some files could not be deleted")
		return
	}
	ui.toast("Images cleared")
}

func (ui *stashApp) currentFile() store.File {
	return store.File{
		Snippets: ui.snippets,
		Images:   ui.images,
		Settings: ui.settings,
	}
}

func (ui *stashApp) save() error {
	if err := ui.store.Save(ui.currentFile()); err != nil {
		ui.toast("Save failed")
		return err
	}
	return nil
}

func (ui *stashApp) openShortcutDialog() {
	current := ui.settings.Shortcut.Normalize()
	labels := keybind.Labels()

	keySelect := widget.NewSelect(keybind.Keys(), nil)
	keySelect.SetSelected(current.Key)

	commandCheck := widget.NewCheck(labels.Command, nil)
	commandCheck.SetChecked(current.Command)
	controlCheck := widget.NewCheck(labels.Control, nil)
	controlCheck.SetChecked(current.Control)
	optionCheck := widget.NewCheck(labels.Option, nil)
	optionCheck.SetChecked(current.Option)
	shiftCheck := widget.NewCheck(labels.Shift, nil)
	shiftCheck.SetChecked(current.Shift)

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Key", keySelect),
			widget.NewFormItem("Modifiers", container.NewVBox(
				commandCheck,
				controlCheck,
				optionCheck,
				shiftCheck,
			)),
		},
		OnSubmit: func() {
			binding := keybind.Binding{
				Key:     keySelect.Selected,
				Command: commandCheck.Checked,
				Control: controlCheck.Checked,
				Option:  optionCheck.Checked,
				Shift:   shiftCheck.Checked,
			}.Normalize()

			if !binding.HasModifier() {
				dialog.ShowInformation("Shortcut needs a modifier", "Choose at least one modifier so Stash does not capture normal typing.", ui.window)
				return
			}

			if err := hotkey.Register(binding, func() {
				fyne.Do(ui.toggle)
			}); err != nil {
				dialog.ShowError(err, ui.window)
				return
			}

			ui.settings.Shortcut = binding
			ui.shortcutBadge.SetText(binding.Display())
			if err := ui.save(); err != nil {
				return
			}
			ui.toast("Shortcut saved")
		},
		OnCancel:   func() {},
		SubmitText: "Save Shortcut",
		CancelText: "Cancel",
	}

	resetButton := widget.NewButton("Reset to "+keybind.Default().Display(), func() {
		binding := keybind.Default()
		if err := hotkey.Register(binding, func() {
			fyne.Do(ui.toggle)
		}); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		ui.settings.Shortcut = binding
		ui.shortcutBadge.SetText(binding.Display())
		if err := ui.save(); err != nil {
			return
		}
		ui.toast("Shortcut reset")
	})

	dialog.ShowCustom("Change Shortcut", "Close", container.NewVBox(form, resetButton), ui.window)
}

func (ui *stashApp) toggle() {
	if ui.hidden {
		ui.show()
		return
	}
	ui.hide()
}

func (ui *stashApp) show() {
	ui.window.Show()
	ui.window.RequestFocus()
	ui.hidden = false
}

func (ui *stashApp) hide() {
	ui.window.Hide()
	ui.hidden = true
}
