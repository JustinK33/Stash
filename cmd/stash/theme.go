package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"stash/resources"
)

var (
	colorBackground = rgb(247, 248, 250)
	colorSidebar    = rgb(242, 244, 247)
	colorSurface    = rgb(255, 255, 255)
	colorSurfaceHov = rgb(251, 252, 254)
	colorBorder     = rgb(227, 231, 237)
	colorBorderHov  = rgb(205, 213, 224)
	colorText       = rgb(17, 24, 39)
	colorMuted      = rgb(107, 114, 128)
	colorFaint      = rgb(148, 155, 168)
	colorWhite      = rgb(255, 255, 255)
	colorPrimary    = rgb(37, 99, 235)
	colorSelection  = rgb(230, 238, 252)
	colorHover      = rgb(234, 237, 242)
	colorToast      = rgb(17, 24, 39)

	interRegular  = fyne.NewStaticResource("Inter-Regular.ttf", resources.InterRegular)
	interSemiBold = fyne.NewStaticResource("Inter-SemiBold.ttf", resources.InterSemiBold)
)

func rgb(r, g, b uint8) color.NRGBA {
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

type stashTheme struct{}

func (stashTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return colorPrimary
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameButton:
		return colorSurface
	case theme.ColorNameDisabledButton:
		return rgb(232, 236, 242)
	case theme.ColorNameDisabled:
		return rgb(138, 146, 160)
	case theme.ColorNameFocus:
		return color.NRGBA{R: 37, G: 99, B: 235, A: 70}
	case theme.ColorNameForeground:
		return colorText
	case theme.ColorNameForegroundOnPrimary:
		return rgb(255, 255, 255)
	case theme.ColorNameHover:
		return color.NRGBA{R: 15, G: 23, B: 42, A: 14}
	case theme.ColorNameInputBackground:
		return colorSurface
	case theme.ColorNameInputBorder:
		return colorBorder
	case theme.ColorNamePlaceHolder:
		return rgb(148, 155, 168)
	case theme.ColorNamePressed:
		return color.NRGBA{R: 15, G: 23, B: 42, A: 26}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 15, G: 23, B: 42, A: 70}
	case theme.ColorNameScrollBarBackground:
		return color.NRGBA{A: 0}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 37, G: 99, B: 235, A: 50}
	case theme.ColorNameSeparator:
		return colorBorder
	case theme.ColorNameShadow:
		return color.NRGBA{R: 15, G: 23, B: 42, A: 18}
	case theme.ColorNameOverlayBackground, theme.ColorNameMenuBackground:
		return colorSurface
	default:
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
}

func (stashTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace || style.Symbol {
		return theme.DefaultTheme().Font(style)
	}
	if style.Bold {
		return interSemiBold
	}
	return interRegular
}

func (stashTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (stashTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameHeadingText:
		return 26
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameInputRadius:
		return 8
	case theme.SizeNameSelectionRadius:
		return 6
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameLineSpacing:
		return 6
	case theme.SizeNameScrollBarRadius:
		return 4
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// borderless hides the input chrome so an entry can sit inside a custom surface.
type borderless struct{ stashTheme }

func (t borderless) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameInputBackground, theme.ColorNameInputBorder, theme.ColorNameFocus, theme.ColorNameScrollBar:
		return color.Transparent
	}
	return t.stashTheme.Color(name, variant)
}

func (t borderless) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 13
	case theme.SizeNameInnerPadding:
		return 4
	}
	return t.stashTheme.Size(name)
}
