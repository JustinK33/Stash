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
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"stash/internal/hotkey"
	"stash/internal/imageclipboard"
	"stash/internal/keybind"
	"stash/internal/store"
)

type stashApp struct {
	window   fyne.Window
	store    *store.Store
	snippets []store.Snippet
	images   []store.Image
	settings store.Settings

	input          *widget.Entry
	textList       *fyne.Container
	imageList      *fyne.Container
	textCount      *widget.Label
	imageCount     *widget.Label
	tabs           *container.AppTabs
	imageTab       *container.TabItem
	shortcutButton *widget.Button
	status         *widget.Label
	hidden         bool
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

	window := fyneApp.NewWindow("Stash")
	window.Resize(fyne.NewSize(620, 700))

	ui := &stashApp{
		window:     window,
		store:      stashStore,
		snippets:   file.Snippets,
		images:     file.Images,
		settings:   file.Settings,
		input:      widget.NewMultiLineEntry(),
		textList:   container.NewVBox(),
		imageList:  container.NewVBox(),
		textCount:  widget.NewLabel("0 saved"),
		imageCount: widget.NewLabel("0 saved"),
		status:     widget.NewLabel("Ready"),
	}

	window.SetContent(ui.build())
	window.SetOnDropped(ui.handleDropped)
	window.Canvas().AddShortcut(&fyne.ShortcutPaste{}, func(fyne.Shortcut) {
		if ui.tabs.Selected() == ui.imageTab {
			ui.pasteImage()
		}
	})
	fyneApp.Lifecycle().SetOnEnteredForeground(func() {
		fyne.Do(ui.show)
	})
	ui.refreshTextList()
	ui.refreshImageList()

	if err := hotkey.Register(ui.settings.Shortcut, func() {
		fyne.Do(ui.toggle)
	}); err != nil {
		ui.status.SetText("Shortcut unavailable")
	}
	defer hotkey.Unregister()

	window.ShowAndRun()
}

func (ui *stashApp) build() fyne.CanvasObject {
	ui.shortcutButton = widget.NewButton(ui.settings.Shortcut.Display(), ui.openShortcutDialog)
	header := container.NewHBox(
		widget.NewLabelWithStyle("Stash", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		ui.shortcutButton,
	)

	textTab := container.NewTabItem("Text", ui.buildTextTab())
	ui.imageTab = container.NewTabItem("Images", ui.buildImageTab())
	ui.tabs = container.NewAppTabs(textTab, ui.imageTab)
	ui.tabs.SetTabLocation(container.TabLocationTop)

	content := container.NewBorder(header, ui.status, nil, nil, ui.tabs)
	return container.NewPadded(content)
}

func (ui *stashApp) buildTextTab() fyne.CanvasObject {
	ui.input.SetPlaceHolder("Paste text here to save it")
	ui.input.Wrapping = fyne.TextWrapWord
	ui.input.SetMinRowsVisible(6)

	saveButton := widget.NewButton("Save", ui.saveSnippet)
	saveButton.Importance = widget.HighImportance
	clearButton := widget.NewButton("Clear Input", func() {
		ui.input.SetText("")
		ui.status.SetText("Ready")
	})
	clearAllButton := widget.NewButton("Clear Text", func() {
		ui.snippets = nil
		if err := ui.save(); err != nil {
			return
		}
		ui.refreshTextList()
		ui.status.SetText("Text cleared")
	})

	actions := container.NewHBox(saveButton, clearButton, layout.NewSpacer(), clearAllButton)
	listHeader := container.NewHBox(
		widget.NewLabelWithStyle("Saved text", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		ui.textCount,
	)
	return container.NewBorder(
		container.NewVBox(ui.input, actions, listHeader),
		nil,
		nil,
		nil,
		container.NewVScroll(ui.textList),
	)
}

func (ui *stashApp) buildImageTab() fyne.CanvasObject {
	pasteButton := widget.NewButton("Paste Image", ui.pasteImage)
	pasteButton.Importance = widget.HighImportance
	clearImagesButton := widget.NewButton("Clear Images", ui.clearImages)

	instructions := widget.NewLabel("Drag PNG, JPEG, or GIF files here, or paste an image with Command + V.")
	instructions.Wrapping = fyne.TextWrapWord
	actions := container.NewHBox(pasteButton, layout.NewSpacer(), clearImagesButton)
	listHeader := container.NewHBox(
		widget.NewLabelWithStyle("Saved images", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		ui.imageCount,
	)
	return container.NewBorder(
		container.NewVBox(instructions, actions, listHeader),
		nil,
		nil,
		nil,
		container.NewVScroll(ui.imageList),
	)
}

func (ui *stashApp) saveSnippet() {
	text := strings.TrimSpace(ui.input.Text)
	if text == "" {
		ui.status.SetText("Nothing to save")
		return
	}

	ui.snippets = store.Add(ui.snippets, text)
	ui.input.SetText("")
	if err := ui.save(); err != nil {
		return
	}
	ui.refreshTextList()
	ui.status.SetText("Saved")
}

func (ui *stashApp) copySnippet(text string) {
	ui.window.Clipboard().SetContent(text)
	ui.status.SetText("Copied")
}

func (ui *stashApp) deleteSnippet(text string) {
	ui.snippets = store.Delete(ui.snippets, text)
	if err := ui.save(); err != nil {
		return
	}
	ui.refreshTextList()
	ui.status.SetText("Deleted")
}

func (ui *stashApp) refreshTextList() {
	ui.textList.Objects = nil

	if len(ui.snippets) == 0 {
		empty := widget.NewLabel("No saved text yet")
		empty.Alignment = fyne.TextAlignCenter
		ui.textList.Add(container.NewPadded(empty))
	} else {
		for _, snippet := range ui.snippets {
			ui.textList.Add(ui.snippetRow(snippet))
		}
	}

	ui.textCount.SetText(fmt.Sprintf("%d saved", len(ui.snippets)))
	ui.textList.Refresh()
}

func (ui *stashApp) snippetRow(snippet store.Snippet) fyne.CanvasObject {
	text := widget.NewLabel(snippet.Text)
	text.Wrapping = fyne.TextWrapWord

	copyButton := widget.NewButton("Copy", func() {
		ui.copySnippet(snippet.Text)
	})
	deleteButton := widget.NewButton("Delete", func() {
		ui.deleteSnippet(snippet.Text)
	})

	return separatedRow(container.NewBorder(
		nil,
		nil,
		nil,
		container.NewHBox(copyButton, deleteButton),
		text,
	))
}

func (ui *stashApp) pasteImage() {
	data, err := imageclipboard.ReadPNG()
	if err != nil {
		ui.status.SetText(err.Error())
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
			ui.status.SetText("Could not read " + uri.Name())
			continue
		}
		if ui.importImage(reader, uri.Name()) {
			imported++
		}
		_ = reader.Close()
	}
	if imported > 0 {
		ui.tabs.Select(ui.imageTab)
		ui.status.SetText(fmt.Sprintf("Imported %d image(s)", imported))
	}
}

func (ui *stashApp) importImage(reader io.Reader, name string) bool {
	file := ui.currentFile()
	_, err := ui.store.ImportImage(&file, reader, name)
	if err != nil {
		ui.status.SetText(err.Error())
		return false
	}
	ui.images = file.Images
	ui.refreshImageList()
	ui.status.SetText("Image saved")
	return true
}

func (ui *stashApp) refreshImageList() {
	ui.imageList.Objects = nil
	if len(ui.images) == 0 {
		empty := widget.NewLabel("No saved images yet")
		empty.Alignment = fyne.TextAlignCenter
		ui.imageList.Add(container.NewPadded(empty))
	} else {
		for _, savedImage := range ui.images {
			ui.imageList.Add(ui.imageRow(savedImage))
		}
	}
	ui.imageCount.SetText(fmt.Sprintf("%d saved", len(ui.images)))
	ui.imageList.Refresh()
}

func (ui *stashApp) imageRow(savedImage store.Image) fyne.CanvasObject {
	imagePath := ui.store.ImagePath(savedImage)
	thumbnail := canvas.NewImageFromFile(imagePath)
	thumbnail.FillMode = canvas.ImageFillContain
	thumbnail.SetMinSize(fyne.NewSize(180, 130))

	name := widget.NewLabelWithStyle(savedImage.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	details := widget.NewLabel(fmt.Sprintf(
		"%d × %d  •  %s",
		savedImage.Width,
		savedImage.Height,
		formatBytes(savedImage.Size),
	))
	copyButton := widget.NewButton("Copy", func() {
		ui.copyImage(savedImage)
	})
	deleteButton := widget.NewButton("Delete", func() {
		ui.deleteImage(savedImage)
	})
	metadata := container.NewVBox(name, details, layout.NewSpacer(), container.NewHBox(copyButton, deleteButton))

	return separatedRow(container.NewBorder(nil, nil, nil, metadata, thumbnail))
}

func (ui *stashApp) copyImage(savedImage store.Image) {
	file, err := os.Open(ui.store.ImagePath(savedImage))
	if err != nil {
		ui.status.SetText("Could not read saved image")
		return
	}
	defer file.Close()

	decoded, _, err := image.Decode(file)
	if err != nil {
		ui.status.SetText("Saved image is invalid")
		return
	}
	var data bytes.Buffer
	if err := png.Encode(&data, decoded); err != nil {
		ui.status.SetText("Could not prepare image")
		return
	}
	if err := imageclipboard.WritePNG(data.Bytes()); err != nil {
		ui.status.SetText(err.Error())
		return
	}
	ui.status.SetText("Image copied")
}

func (ui *stashApp) deleteImage(savedImage store.Image) {
	file := ui.currentFile()
	err := ui.store.DeleteImage(&file, savedImage.ID)
	ui.images = file.Images
	ui.refreshImageList()
	if err != nil {
		ui.status.SetText("Image removed, but its file could not be deleted")
		return
	}
	ui.status.SetText("Image deleted")
}

func (ui *stashApp) clearImages() {
	file := ui.currentFile()
	err := ui.store.ClearImages(&file)
	ui.images = file.Images
	ui.refreshImageList()
	if err != nil {
		ui.status.SetText("Images cleared, but some files could not be deleted")
		return
	}
	ui.status.SetText("Images cleared")
}

func separatedRow(content fyne.CanvasObject) fyne.CanvasObject {
	line := canvas.NewLine(color.NRGBA{R: 220, G: 225, B: 232, A: 255})
	line.StrokeWidth = 1
	return container.NewVBox(content, line)
}

func formatBytes(size int64) string {
	if size >= 1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.1f KB", float64(size)/1024)
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
		ui.status.SetText("Save failed")
		return err
	}
	return nil
}

func (ui *stashApp) openShortcutDialog() {
	current := ui.settings.Shortcut.Normalize()

	keySelect := widget.NewSelect(keybind.Keys(), nil)
	keySelect.SetSelected(current.Key)

	commandCheck := widget.NewCheck("Command", nil)
	commandCheck.SetChecked(current.Command)
	controlCheck := widget.NewCheck("Control", nil)
	controlCheck.SetChecked(current.Control)
	optionCheck := widget.NewCheck("Option", nil)
	optionCheck.SetChecked(current.Option)
	shiftCheck := widget.NewCheck("Shift", nil)
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
			ui.shortcutButton.SetText(binding.Display())
			if err := ui.save(); err != nil {
				return
			}
			ui.status.SetText("Shortcut saved")
		},
		OnCancel:   func() {},
		SubmitText: "Save Shortcut",
		CancelText: "Cancel",
	}

	resetButton := widget.NewButton("Reset to Control + Option + 0", func() {
		binding := keybind.Default()
		if err := hotkey.Register(binding, func() {
			fyne.Do(ui.toggle)
		}); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		ui.settings.Shortcut = binding
		ui.shortcutButton.SetText(binding.Display())
		if err := ui.save(); err != nil {
			return
		}
		ui.status.SetText("Shortcut reset")
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
