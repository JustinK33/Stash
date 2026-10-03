package main

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	cardRadius    = 10
	hoverDuration = 120 * time.Millisecond
)

// animateColor tweens a color, cancelling whatever tween was running before.
func animateColor(running **fyne.Animation, from, to color.Color, set func(color.Color)) {
	if *running != nil {
		(*running).Stop()
	}
	anim := canvas.NewColorRGBAAnimation(from, to, hoverDuration, set)
	anim.Curve = fyne.AnimationEaseOut
	*running = anim
	anim.Start()
}

func roundedRect(fill, stroke color.Color, radius float32) *canvas.Rectangle {
	rect := canvas.NewRectangle(fill)
	rect.StrokeColor = stroke
	rect.StrokeWidth = 1
	rect.CornerRadius = radius
	return rect
}

// inset pads content by the given amount on every side.
func inset(content fyne.CanvasObject, amount float32) fyne.CanvasObject {
	return container.New(&insetLayout{amount}, content)
}

type insetLayout struct{ amount float32 }

func (l *insetLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(l.amount, l.amount))
		o.Resize(size.SubtractWidthHeight(l.amount*2, l.amount*2).Max(fyne.NewSize(0, 0)))
	}
}

func (l *insetLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	min := fyne.NewSize(0, 0)
	for _, o := range objects {
		min = min.Max(o.MinSize())
	}
	return min.AddWidthHeight(l.amount*2, l.amount*2)
}

// surface is a static white rounded panel with a hairline border.
func surface(content fyne.CanvasObject, padding float32) fyne.CanvasObject {
	return container.NewStack(roundedRect(colorSurface, colorBorder, cardRadius), inset(content, padding))
}

// card is a surface that tints on hover and tells its children about it.
type card struct {
	widget.BaseWidget
	bg       *canvas.Rectangle
	content  fyne.CanvasObject
	hovered  bool
	fill     *fyne.Animation
	stroke   *fyne.Animation
	OnHover  func(bool)
	children []*hoverButton
}

func newCard(content fyne.CanvasObject, padding float32, children ...*hoverButton) *card {
	c := &card{bg: roundedRect(colorSurface, colorBorder, cardRadius), content: inset(content, padding), children: children}
	for _, child := range children {
		child.onHover = c.setHovered
	}
	c.ExtendBaseWidget(c)
	return c
}

func (c *card) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(c.bg, c.content))
}

func (c *card) MouseIn(*desktop.MouseEvent)    { c.setHovered(true) }
func (c *card) MouseMoved(*desktop.MouseEvent) {}
func (c *card) MouseOut()                      { c.setHovered(false) }

// setHovered is shared with child buttons, because Fyne only hovers the
// deepest object under the pointer and the card would otherwise flicker.
func (c *card) setHovered(hovered bool) {
	if c.hovered == hovered {
		return
	}
	c.hovered = hovered
	fill, stroke := colorSurface, colorBorder
	if hovered {
		fill, stroke = colorSurfaceHov, colorBorderHov
	}
	animateColor(&c.fill, c.bg.FillColor, fill, func(col color.Color) { c.bg.FillColor = col; c.bg.Refresh() })
	animateColor(&c.stroke, c.bg.StrokeColor, stroke, func(col color.Color) { c.bg.StrokeColor = col; c.bg.Refresh() })
	for _, child := range c.children {
		child.setEmphasis(hovered)
	}
	if c.OnHover != nil {
		c.OnHover(hovered)
	}
}

// hoverButton is a quiet icon button that brightens when its card is hovered.
type hoverButton struct {
	widget.Button
	base    string
	onHover func(bool)
	flash   *time.Timer
}

func newHoverButton(label string, iconName string, tapped func()) *hoverButton {
	b := &hoverButton{base: iconName}
	b.Text = label
	b.OnTapped = tapped
	b.Importance = widget.LowImportance
	b.ExtendBaseWidget(b)
	b.setEmphasis(false)
	return b
}

func (b *hoverButton) MouseIn(e *desktop.MouseEvent) {
	b.Button.MouseIn(e)
	if b.onHover != nil {
		b.onHover(true)
	}
}

func (b *hoverButton) MouseOut() {
	b.Button.MouseOut()
	if b.onHover != nil {
		b.onHover(false)
	}
}

func (b *hoverButton) Cursor() desktop.Cursor { return desktop.PointerCursor }

func (b *hoverButton) setEmphasis(on bool) {
	if b.base == "" || b.flash != nil {
		return
	}
	if on {
		b.SetIcon(icon(b.base, colorText))
	} else {
		b.SetIcon(icon(b.base, colorFaint))
	}
}

// flashConfirm swaps the icon for a blue checkmark for a moment.
func (b *hoverButton) flashConfirm() {
	if b.flash != nil {
		b.flash.Stop()
	}
	b.SetIcon(icon("check", colorPrimary))
	b.flash = time.AfterFunc(1200*time.Millisecond, func() {
		fyne.Do(func() {
			b.flash = nil
			b.setEmphasis(true)
		})
	})
}

// shortcutBadge looks like a keycap and opens the shortcut settings.
type shortcutBadge struct {
	widget.BaseWidget
	bg    *canvas.Rectangle
	text  *canvas.Text
	onTap func()
	anim  *fyne.Animation
}

func newShortcutBadge(text string, tapped func()) *shortcutBadge {
	label := canvas.NewText(text, colorText)
	label.TextSize = 13
	label.TextStyle.Bold = true
	b := &shortcutBadge{bg: roundedRect(colorSurface, colorBorder, 7), text: label, onTap: tapped}
	b.ExtendBaseWidget(b)
	return b
}

func (b *shortcutBadge) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(b.bg, container.NewCenter(b.text)))
}

func (b *shortcutBadge) MinSize() fyne.Size {
	return b.text.MinSize().AddWidthHeight(26, 16)
}

func (b *shortcutBadge) SetText(text string) {
	b.text.Text = text
	b.text.Refresh()
	b.Refresh()
}

func (b *shortcutBadge) Tapped(*fyne.PointEvent) { b.onTap() }
func (b *shortcutBadge) Cursor() desktop.Cursor  { return desktop.PointerCursor }
func (b *shortcutBadge) MouseIn(*desktop.MouseEvent) {
	animateColor(&b.anim, b.bg.FillColor, colorSidebar, func(c color.Color) { b.bg.FillColor = c; b.bg.Refresh() })
}
func (b *shortcutBadge) MouseMoved(*desktop.MouseEvent) {}
func (b *shortcutBadge) MouseOut() {
	animateColor(&b.anim, b.bg.FillColor, colorSurface, func(c color.Color) { b.bg.FillColor = c; b.bg.Refresh() })
}

// outlined draws a hairline border around a secondary button.
func outlined(button *widget.Button) fyne.CanvasObject {
	return container.NewStack(button, roundedRect(color.Transparent, colorBorder, 8))
}

// gap is fixed empty space for vertical rhythm.
func gap(height float32) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, height))
	return spacer
}

func newText(text string, size float32, col color.Color, bold bool) *canvas.Text {
	t := canvas.NewText(text, col)
	t.TextSize = size
	t.TextStyle.Bold = bold
	return t
}

func pageHeader(title, subtitle string, trailing fyne.CanvasObject) fyne.CanvasObject {
	heading := newText(title, 28, colorText, true)
	sub := newText(subtitle, 14, colorMuted, false)
	titles := container.New(layout.NewCustomPaddedVBoxLayout(4), heading, sub)
	if trailing == nil {
		return titles
	}
	return container.NewBorder(nil, nil, nil, container.NewVBox(trailing), titles)
}

func sectionHeader(title string, trailing fyne.CanvasObject) fyne.CanvasObject {
	return container.NewBorder(nil, nil, newText(title, 15, colorText, true), trailing)
}

func emptyState(iconName, title, body string) fyne.CanvasObject {
	image := widget.NewIcon(icon(iconName, colorFaint))
	sizedIcon := container.NewGridWrap(fyne.NewSize(40, 40), image)
	heading := newText(title, 15, colorText, true)
	heading.Alignment = fyne.TextAlignCenter
	sub := newText(body, 13, colorMuted, false)
	sub.Alignment = fyne.TextAlignCenter
	return container.NewPadded(container.NewCenter(container.New(layout.NewCustomPaddedVBoxLayout(6),
		layout.NewSpacer(),
		container.NewCenter(sizedIcon),
		heading,
		sub,
		layout.NewSpacer(),
	)))
}

// maxWidthLayout centers its content and stops it growing past max.
type maxWidthLayout struct {
	max     float32
	padding float32
}

func maxWidth(max, padding float32, content fyne.CanvasObject) fyne.CanvasObject {
	return container.New(&maxWidthLayout{max: max, padding: padding}, content)
}

func (l *maxWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	width := max(0, min(size.Width-l.padding*2, l.max))
	for _, o := range objects {
		o.Resize(fyne.NewSize(width, size.Height))
		o.Move(fyne.NewPos((size.Width-width)/2, 0))
	}
}

func (l *maxWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	min := fyne.NewSize(0, 0)
	for _, o := range objects {
		min = min.Max(o.MinSize())
	}
	return min.AddWidthHeight(l.padding*2, 0)
}

// toaster shows short-lived messages in a dark pill near the bottom edge.
type toaster struct {
	bg    *canvas.Rectangle
	text  *canvas.Text
	layer *fyne.Container
	anim  *fyne.Animation
	timer *time.Timer
}

func newToaster() *toaster {
	t := &toaster{bg: canvas.NewRectangle(colorToast), text: newText("", 13, color.White, false)}
	t.bg.CornerRadius = 9
	pill := container.NewStack(t.bg, inset(t.text, 10))
	t.layer = container.NewBorder(nil, inset(container.NewCenter(pill), 20), nil, nil)
	t.layer.Hide()
	return t
}

func (t *toaster) show(message string) {
	if t.timer != nil {
		t.timer.Stop()
	}
	t.text.Text = message
	t.layer.Show()
	t.fade(0, 1)
	t.layer.Refresh()
	t.timer = time.AfterFunc(1600*time.Millisecond, func() {
		fyne.Do(func() { t.fade(1, 0) })
	})
}

func (t *toaster) fade(from, to float32) {
	if t.anim != nil {
		t.anim.Stop()
	}
	t.anim = fyne.NewAnimation(160*time.Millisecond, func(progress float32) {
		alpha := uint8(255 * (from + (to-from)*progress))
		t.bg.FillColor = color.NRGBA{R: colorToast.R, G: colorToast.G, B: colorToast.B, A: alpha}
		t.text.Color = color.NRGBA{R: 255, G: 255, B: 255, A: alpha}
		t.bg.Refresh()
		t.text.Refresh()
		if progress == 1 && to == 0 {
			t.layer.Hide()
		}
	})
	t.anim.Start()
}
