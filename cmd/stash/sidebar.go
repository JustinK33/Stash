package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"stash/internal/keybind"
	"stash/resources"
)

const (
	sidebarWidth        = 244
	sidebarCompactWidth = 66
	compactBelow        = 760
)

// navItem is a sidebar row: icon plus label, blue when selected.
type navItem struct {
	widget.BaseWidget
	icon     string
	bg       *canvas.Rectangle
	iconView *widget.Icon
	label    *canvas.Text
	onTap    func()
	anim     *fyne.Animation

	selected, hovered, focused, compact bool
}

func newNavItem(label string, iconName string, tapped func()) *navItem {
	n := &navItem{
		icon:     iconName,
		bg:       canvas.NewRectangle(color.Transparent),
		iconView: widget.NewIcon(icon(iconName, colorText)),
		label:    newText(label, 15, colorText, false),
		onTap:    tapped,
	}
	n.bg.CornerRadius = 8
	n.ExtendBaseWidget(n)
	n.apply(false)
	return n
}

func (n *navItem) CreateRenderer() fyne.WidgetRenderer {
	row := container.NewHBox(container.NewGridWrap(fyne.NewSize(20, 20), n.iconView), n.label)
	return widget.NewSimpleRenderer(container.NewStack(n.bg, container.New(&navLayout{item: n}, row)))
}

// navLayout insets the row, or centers just the icon when compact.
type navLayout struct{ item *navItem }

func (l *navLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	row := objects[0]
	rowMin := row.MinSize()
	x := float32(12)
	if l.item.compact {
		x = (size.Width - 20) / 2
	}
	row.Resize(fyne.NewSize(size.Width-x, rowMin.Height))
	row.Move(fyne.NewPos(x, (size.Height-rowMin.Height)/2))
}

func (l *navLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(40, 42) }

func (n *navItem) setSelected(selected bool) {
	n.selected = selected
	n.apply(true)
}

func (n *navItem) setCompact(compact bool) {
	n.compact = compact
	if compact {
		n.label.Hide()
	} else {
		n.label.Show()
	}
	n.Refresh()
}

func (n *navItem) apply(animated bool) {
	fill := color.Color(color.Transparent)
	switch {
	case n.selected:
		fill = colorSelection
	case n.hovered:
		fill = colorHover
	}
	if n.focused {
		n.bg.StrokeColor = color.NRGBA{R: 37, G: 99, B: 235, A: 120}
		n.bg.StrokeWidth = 1.5
	} else {
		n.bg.StrokeWidth = 0
	}
	if animated {
		animateColor(&n.anim, n.bg.FillColor, fill, func(c color.Color) { n.bg.FillColor = c; n.bg.Refresh() })
	} else {
		n.bg.FillColor = fill
	}
	n.bg.Refresh()

	if n.selected {
		n.iconView.SetResource(icon(n.icon, colorPrimary))
		n.label.Color = colorPrimary
	} else {
		n.iconView.SetResource(icon(n.icon, colorText))
		n.label.Color = colorText
	}
	n.label.Refresh()
}

func (n *navItem) Tapped(*fyne.PointEvent)        { n.onTap() }
func (n *navItem) Cursor() desktop.Cursor         { return desktop.PointerCursor }
func (n *navItem) MouseIn(*desktop.MouseEvent)    { n.hovered = true; n.apply(true) }
func (n *navItem) MouseMoved(*desktop.MouseEvent) {}
func (n *navItem) MouseOut()                      { n.hovered = false; n.apply(true) }
func (n *navItem) FocusGained()                   { n.focused = true; n.apply(false) }
func (n *navItem) FocusLost()                     { n.focused = false; n.apply(false) }
func (n *navItem) TypedRune(r rune) {
	if r == ' ' {
		n.onTap()
	}
}
func (n *navItem) TypedKey(e *fyne.KeyEvent) {
	if e.Name == fyne.KeyReturn || e.Name == fyne.KeyEnter {
		n.onTap()
	}
}

type sidebar struct {
	root       *fyne.Container
	wordmark   *canvas.Text
	searchBox  fyne.CanvasObject
	searchIcon *navItem
	search     *widget.Entry
	items      []*navItem
	compact    bool
	// expanded keeps the full sidebar open while searching in a narrow window.
	expanded bool
}

func newSidebar(search *widget.Entry, nav []*navItem, settings *navItem, onSearchIcon func()) *sidebar {
	s := &sidebar{search: search, items: append(append([]*navItem{}, nav...), settings)}

	logo := canvas.NewImageFromResource(fyne.NewStaticResource("logo.png", resources.Logo))
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(30, 30))
	s.wordmark = newText("Stash", 20, colorText, true)
	brand := container.NewHBox(logo, container.NewCenter(s.wordmark))

	search.SetPlaceHolder("Search saved text...")
	searchIcon := widget.NewIcon(icon("search", colorMuted))
	hint := container.NewStack(roundedRect(colorSidebar, colorBorder, 5), inset(newText(keybind.Labels().Shortcut+" K", 11, colorMuted, false), 3))
	searchBg := roundedRect(colorSurface, colorBorder, 8)
	searchBg.SetMinSize(fyne.NewSize(0, 36))
	s.searchBox = container.NewStack(
		searchBg,
		container.NewBorder(nil, nil,
			container.NewHBox(gap(0), container.NewCenter(container.NewGridWrap(fyne.NewSize(16, 16), searchIcon))),
			container.NewHBox(container.NewCenter(hint), gap(0)),
			container.NewVBox(layout.NewSpacer(), container.NewThemeOverride(search, borderless{}), layout.NewSpacer()),
		),
	)
	s.searchIcon = newNavItem("Search", "search", onSearchIcon)
	s.searchIcon.setCompact(true)
	s.searchIcon.Hide()

	navList := container.New(layout.NewCustomPaddedVBoxLayout(4))
	navList.Add(s.searchIcon)
	for _, item := range nav {
		navList.Add(item)
	}

	top := container.New(layout.NewCustomPaddedVBoxLayout(14), inset(brand, 4), s.searchBox, navList)
	content := inset(container.NewBorder(top, settings, nil, nil), 14)

	divider := canvas.NewRectangle(colorBorder)
	divider.SetMinSize(fyne.NewSize(1, 1))
	s.root = container.NewStack(canvas.NewRectangle(colorSidebar), container.NewBorder(nil, nil, nil, divider, content))
	return s
}

func (s *sidebar) width() float32 {
	if s.compact && !s.expanded {
		return sidebarCompactWidth
	}
	return sidebarWidth
}

func (s *sidebar) setCompact(compact bool) {
	if s.compact == compact {
		return
	}
	s.compact = compact
	s.apply()
}

func (s *sidebar) setExpanded(expanded bool) {
	if s.expanded == expanded {
		return
	}
	s.expanded = expanded
	s.apply()
}

func (s *sidebar) apply() {
	narrow := s.compact && !s.expanded
	for _, item := range s.items {
		item.setCompact(narrow)
	}
	if narrow {
		s.wordmark.Hide()
		s.searchBox.Hide()
		s.searchIcon.Show()
	} else {
		s.wordmark.Show()
		s.searchBox.Show()
		s.searchIcon.Hide()
	}
	s.root.Refresh()
}

// shellLayout places the sidebar and page side by side and collapses the
// sidebar to an icon rail in narrow windows.
type shellLayout struct{ sidebar *sidebar }

func (l *shellLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.sidebar.setCompact(size.Width < compactBelow)
	w := l.sidebar.width()
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(w, size.Height))
	objects[1].Move(fyne.NewPos(w, 0))
	objects[1].Resize(fyne.NewSize(size.Width-w, size.Height))
}

func (l *shellLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	page := objects[1].MinSize()
	return fyne.NewSize(sidebarCompactWidth+page.Width, max(objects[0].MinSize().Height, page.Height))
}
