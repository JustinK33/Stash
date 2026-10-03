package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"stash/internal/store"
)

var imageCardSize = fyne.NewSize(200, 218)

func (ui *stashApp) buildImagePage() fyne.CanvasObject {
	pasteButton := widget.NewButtonWithIcon("Paste Image", icon("paste", colorWhite), ui.pasteImage)
	pasteButton.Importance = widget.HighImportance
	clearButton := widget.NewButtonWithIcon("Clear All", icon("clearAll", colorText), ui.confirmClearImages)

	dropIcon := container.NewGridWrap(fyne.NewSize(36, 36), widget.NewIcon(icon("image", colorPrimary)))
	title := newText("Drop images here", 15, colorText, true)
	formats := newText("PNG, JPEG, or GIF", 13, colorMuted, false)
	hint := newText("or press ⌘ V", 12, colorMuted, false)
	dropZone := surface(container.NewCenter(container.New(layout.NewCustomPaddedVBoxLayout(8),
		container.NewCenter(dropIcon),
		container.NewCenter(title),
		container.NewCenter(formats),
		gap(2),
		container.NewCenter(container.NewHBox(pasteButton, container.NewCenter(hint))),
	)), 22)

	top := container.New(layout.NewCustomPaddedVBoxLayout(18),
		pageHeader("Images", "Save and quickly reuse images.", ui.imageCount),
		dropZone,
		sectionHeader("Saved Images", outlined(clearButton)),
	)
	return container.NewBorder(
		container.NewVBox(gap(pagePadding), maxWidth(contentMaxWidth, pagePadding, top), gap(4)),
		nil, nil, nil,
		container.NewVScroll(maxWidth(contentMaxWidth, pagePadding, container.NewVBox(ui.imageList, gap(pagePadding)))),
	)
}

func (ui *stashApp) confirmClearImages() {
	if len(ui.images) == 0 {
		return
	}
	dialog.ShowConfirm("Clear all saved images?", fmt.Sprintf("This removes %d saved images.", len(ui.images)), func(ok bool) {
		if ok {
			ui.clearImages()
		}
	}, ui.window)
}

func (ui *stashApp) refreshImageList() {
	ui.imageList.Objects = nil
	if len(ui.images) == 0 {
		ui.imageList.Add(emptyState("image", "No saved images yet", "Paste or drop an image here to get started."))
	} else {
		grid := container.New(&responsiveGrid{cell: imageCardSize, gap: 10})
		for _, savedImage := range ui.images {
			grid.Add(ui.imageCard(savedImage))
		}
		ui.imageList.Add(grid)
	}
	ui.imageCount.Text = fmt.Sprintf("%d saved", len(ui.images))
	ui.imageCount.Refresh()
	ui.imageList.Refresh()
}

func (ui *stashApp) imageCard(savedImage store.Image) fyne.CanvasObject {
	thumbnail := canvas.NewImageFromFile(ui.store.ImagePath(savedImage))
	thumbnail.FillMode = canvas.ImageFillContain
	thumbnail.ScaleMode = canvas.ImageScaleSmooth
	tile := container.NewStack(roundedRect(colorBackground, colorBorder, 7), inset(thumbnail, 6))

	name := widget.NewLabelWithStyle(savedImage.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	name.Truncation = fyne.TextTruncateEllipsis
	details := widget.NewLabel(fmt.Sprintf("%d × %d  •  %s", savedImage.Width, savedImage.Height, formatBytes(savedImage.Size)))
	details.Importance = widget.LowImportance
	details.SizeName = theme.SizeNameCaptionText
	details.Truncation = fyne.TextTruncateEllipsis

	var copyButton *hoverButton
	copyButton = newHoverButton("", "copy", func() {
		if ui.copyImage(savedImage) {
			copyButton.flashConfirm()
		}
	})
	deleteButton := newHoverButton("", "trash", func() {
		ui.deleteImage(savedImage)
	})

	footer := container.New(layout.NewCustomPaddedVBoxLayout(-20),
		container.NewBorder(nil, nil, nil, container.NewHBox(copyButton, deleteButton), name),
		details,
	)
	return newCard(container.NewBorder(nil, footer, nil, nil, tile), 8, copyButton, deleteButton)
}

func formatBytes(size int64) string {
	if size >= 1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.1f KB", float64(size)/1024)
}

// responsiveGrid fits as many columns as the width allows and stretches them
// to fill the row, so the grid never leaves a ragged gap on the right.
type responsiveGrid struct {
	cell fyne.Size
	gap  float32
	cols int // from the last Layout, like widget GridWrap, since MinSize gets no width
}

func (g *responsiveGrid) columns(width float32) int {
	return max(1, int((width+g.gap)/(g.cell.Width+g.gap)))
}

func (g *responsiveGrid) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	cols := g.columns(size.Width)
	g.cols = cols
	cellWidth := (size.Width - g.gap*float32(cols-1)) / float32(cols)
	for i, o := range objects {
		row, col := i/cols, i%cols
		o.Move(fyne.NewPos(float32(col)*(cellWidth+g.gap), float32(row)*(g.cell.Height+g.gap)))
		o.Resize(fyne.NewSize(cellWidth, g.cell.Height))
	}
}

func (g *responsiveGrid) MinSize(objects []fyne.CanvasObject) fyne.Size {
	cols := max(1, g.cols)
	rows := (len(objects) + cols - 1) / cols
	return fyne.NewSize(g.cell.Width, float32(rows)*g.cell.Height+float32(max(0, rows-1))*g.gap)
}
