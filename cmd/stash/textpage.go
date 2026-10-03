package main

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"stash/internal/store"
)

const (
	contentMaxWidth = 820
	pagePadding     = 28
	previewLines    = 6
	previewChars    = 400
)

func (ui *stashApp) buildTextPage() fyne.CanvasObject {
	ui.input.SetPlaceHolder("Paste text here to save it...")
	ui.input.Wrapping = fyne.TextWrapWord
	ui.input.SetMinRowsVisible(5)
	ui.input.OnChanged = func(text string) {
		ui.charCount.Text = characterCount(text)
		ui.charCount.Refresh()
	}
	ui.charCount.Text = characterCount("")
	input := container.NewStack(ui.input, container.NewBorder(nil, container.NewHBox(layout.NewSpacer(), inset(ui.charCount, 10)), nil, nil))

	saveButton := widget.NewButtonWithIcon("Save Text", icon("plus", colorWhite), ui.saveSnippet)
	saveButton.Importance = widget.HighImportance
	clearButton := widget.NewButtonWithIcon("Clear Input", icon("trash", colorText), func() {
		ui.input.SetText("")
		ui.window.Canvas().Focus(ui.input)
	})
	clearAllButton := widget.NewButtonWithIcon("Clear All", icon("clearAll", colorText), ui.confirmClearText)
	actions := container.NewHBox(saveButton, outlined(clearButton), layout.NewSpacer(), outlined(clearAllButton))

	top := container.New(layout.NewCustomPaddedVBoxLayout(18),
		pageHeader("Text", "Quickly save and organize text snippets.", ui.shortcutBadge),
		input,
		actions,
		sectionHeader("Saved Text", ui.textCount),
	)
	return container.NewBorder(
		container.NewVBox(gap(pagePadding), maxWidth(contentMaxWidth, pagePadding, top), gap(4)),
		nil, nil, nil,
		container.NewVScroll(maxWidth(contentMaxWidth, pagePadding, container.NewVBox(ui.textList, gap(pagePadding)))),
	)
}

func (ui *stashApp) confirmClearText() {
	if len(ui.snippets) == 0 {
		return
	}
	dialog.ShowConfirm("Clear all saved text?", fmt.Sprintf("This removes %d saved snippets.", len(ui.snippets)), func(ok bool) {
		if !ok {
			return
		}
		ui.snippets = nil
		if err := ui.save(); err != nil {
			return
		}
		ui.refreshTextList()
		ui.toast("Text cleared")
	}, ui.window)
}

func (ui *stashApp) refreshTextList() {
	query := ui.search.Text
	visible := filterSnippets(ui.snippets, query)

	ui.textList.Objects = nil
	switch {
	case len(ui.snippets) == 0:
		ui.textList.Add(emptyState("text", "No saved text yet", "Paste text above and press Save Text."))
	case len(visible) == 0:
		ui.textList.Add(emptyState("search", "No matches", fmt.Sprintf("Nothing saved matches \"%s\".", strings.TrimSpace(query))))
	default:
		for _, snippet := range visible {
			ui.textList.Add(ui.snippetCard(snippet))
		}
	}

	if strings.TrimSpace(query) != "" {
		ui.textCount.Text = fmt.Sprintf("%d of %d", len(visible), len(ui.snippets))
	} else {
		ui.textCount.Text = fmt.Sprintf("%d saved", len(ui.snippets))
	}
	ui.textCount.Refresh()
	ui.textList.Refresh()
}

func (ui *stashApp) snippetCard(snippet store.Snippet) fyne.CanvasObject {
	preview, truncated := truncateSnippet(snippet.Text, previewLines, previewChars)
	expanded := ui.expanded[snippet.ID]

	text := widget.NewLabel(preview)
	if expanded {
		text.SetText(snippet.Text)
	}
	text.Wrapping = fyne.TextWrapWord
	timestamp := widget.NewLabel(formatTimestamp(snippet.CreatedAt))
	timestamp.Importance = widget.LowImportance
	timestamp.SizeName = theme.SizeNameCaptionText

	var copyButton *hoverButton
	copyButton = newHoverButton("", "copy", func() {
		ui.copySnippet(snippet.Text)
		copyButton.flashConfirm()
	})
	deleteButton := newHoverButton("", "trash", func() {
		ui.deleteSnippet(snippet.Text)
	})

	body := container.New(layout.NewCustomPaddedVBoxLayout(-16), text, timestamp)
	var link *hoverLink
	if truncated {
		label := "Show more"
		if expanded {
			label = "Show less"
		}
		link = newHoverLink(label, func() {
			ui.expanded[snippet.ID] = !expanded
			ui.refreshTextList()
		})
		body.Add(container.NewHBox(link))
	}
	actions := container.NewVBox(layout.NewSpacer(), container.NewHBox(copyButton, deleteButton), layout.NewSpacer())
	c := newCard(container.NewBorder(nil, nil, nil, actions, body), 8, copyButton, deleteButton)
	if link != nil {
		link.onHover = c.setHovered
	}
	return c
}

// hoverLink is a hyperlink that keeps its parent card highlighted.
type hoverLink struct {
	widget.Hyperlink
	onHover func(bool)
}

func newHoverLink(text string, tapped func()) *hoverLink {
	l := &hoverLink{}
	l.Text = text
	l.OnTapped = tapped
	l.SizeName = theme.SizeNameCaptionText
	l.ExtendBaseWidget(l)
	return l
}

func (l *hoverLink) MouseIn(e *desktop.MouseEvent) {
	l.Hyperlink.MouseIn(e)
	if l.onHover != nil {
		l.onHover(true)
	}
}

func (l *hoverLink) MouseOut() {
	l.Hyperlink.MouseOut()
	if l.onHover != nil {
		l.onHover(false)
	}
}

func filterSnippets(snippets []store.Snippet, query string) []store.Snippet {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return snippets
	}
	var matches []store.Snippet
	for _, snippet := range snippets {
		if strings.Contains(strings.ToLower(snippet.Text), query) {
			matches = append(matches, snippet)
		}
	}
	return matches
}

// truncateSnippet shortens text to at most maxLines lines and maxChars runes.
func truncateSnippet(text string, maxLines, maxChars int) (string, bool) {
	truncated := false
	if lines := strings.SplitN(text, "\n", maxLines+1); len(lines) > maxLines {
		text = strings.Join(lines[:maxLines], "\n")
		truncated = true
	}
	if utf8.RuneCountInString(text) > maxChars {
		text = string([]rune(text)[:maxChars])
		truncated = true
	}
	if truncated {
		text = strings.TrimRight(text, " \n\t") + "…"
	}
	return text, truncated
}

func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("Jan 2, 2006, 3:04 PM")
}

func characterCount(text string) string {
	n := utf8.RuneCountInString(text)
	if n == 1 {
		return "1 character"
	}
	return fmt.Sprintf("%s characters", groupThousands(n))
}

func groupThousands(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
