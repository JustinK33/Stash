package main

import (
	"fmt"
	"image/color"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"stash/internal/hotkey"
	"stash/internal/keybind"
	"stash/internal/store"
)

type stashApp struct {
	window   fyne.Window
	store    *store.Store
	snippets []store.Snippet
	settings store.Settings

	input          *widget.Entry
	list           *fyne.Container
	count          *widget.Label
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

	fyneApp := app.NewWithID("local.stash")
	fyneApp.Settings().SetTheme(stashTheme{})

	window := fyneApp.NewWindow("Stash")
	window.Resize(fyne.NewSize(560, 620))

	ui := &stashApp{
		window:   window,
		store:    stashStore,
		snippets: file.Snippets,
		settings: file.Settings,
		input:    widget.NewMultiLineEntry(),
		list:     container.NewVBox(),
		count:    widget.NewLabel("0 saved"),
		status:   widget.NewLabel("Ready"),
	}

	window.SetContent(ui.build())
	window.SetCloseIntercept(func() {
		ui.hide()
	})
	fyneApp.Lifecycle().SetOnEnteredForeground(func() {
		fyne.Do(ui.show)
	})
	ui.refreshList()

	if err := hotkey.Register(ui.settings.Shortcut, func() {
		fyne.Do(ui.toggle)
	}); err != nil {
		ui.status.SetText("Shortcut unavailable")
	}
	defer hotkey.Unregister()

	window.ShowAndRun()
}

func (ui *stashApp) build() fyne.CanvasObject {
	ui.input.SetPlaceHolder("Paste text here to save it")
	ui.input.Wrapping = fyne.TextWrapWord
	ui.input.SetMinRowsVisible(6)

	saveButton := widget.NewButton("Save", ui.saveSnippet)
	saveButton.Importance = widget.HighImportance

	clearButton := widget.NewButton("Clear", func() {
		ui.input.SetText("")
		ui.status.SetText("Ready")
	})

	clearAllButton := widget.NewButton("Clear All", func() {
		ui.snippets = nil
		ui.save()
		ui.refreshList()
		ui.status.SetText("Cleared")
	})

	ui.shortcutButton = widget.NewButton(ui.settings.Shortcut.Display(), ui.openShortcutDialog)
	header := container.NewHBox(
		widget.NewLabelWithStyle("Stash", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		ui.shortcutButton,
	)

	actions := container.NewHBox(saveButton, clearButton, layout.NewSpacer(), clearAllButton)
	listHeader := container.NewHBox(
		widget.NewLabelWithStyle("Saved snippets", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		ui.count,
	)

	content := container.NewBorder(
		container.NewVBox(header, ui.input, actions, listHeader),
		ui.status,
		nil,
		nil,
		container.NewVScroll(ui.list),
	)

	return container.NewPadded(content)
}

func (ui *stashApp) saveSnippet() {
	text := strings.TrimSpace(ui.input.Text)
	if text == "" {
		ui.status.SetText("Nothing to save")
		return
	}

	ui.snippets = store.Add(ui.snippets, text)
	ui.input.SetText("")
	ui.save()
	ui.refreshList()
	ui.status.SetText("Saved")
}

func (ui *stashApp) copySnippet(text string) {
	ui.window.Clipboard().SetContent(text)
	ui.status.SetText("Copied")
}

func (ui *stashApp) deleteSnippet(text string) {
	ui.snippets = store.Delete(ui.snippets, text)
	ui.save()
	ui.refreshList()
	ui.status.SetText("Deleted")
}

func (ui *stashApp) refreshList() {
	ui.list.Objects = nil

	if len(ui.snippets) == 0 {
		empty := widget.NewLabel("No saved snippets yet")
		empty.Alignment = fyne.TextAlignCenter
		ui.list.Add(container.NewPadded(empty))
	} else {
		for _, snippet := range ui.snippets {
			ui.list.Add(ui.snippetRow(snippet))
		}
	}

	ui.count.SetText(fmt.Sprintf("%d saved", len(ui.snippets)))
	ui.list.Refresh()
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

	line := canvas.NewLine(color.NRGBA{R: 220, G: 225, B: 232, A: 255})
	line.StrokeWidth = 1

	return container.NewVBox(
		container.NewBorder(nil, nil, nil, container.NewHBox(copyButton, deleteButton), text),
		line,
	)
}

func (ui *stashApp) save() {
	if err := ui.store.Save(store.File{
		Snippets: ui.snippets,
		Settings: ui.settings,
	}); err != nil {
		ui.status.SetText("Save failed")
	}
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
			ui.save()
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
		ui.save()
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
