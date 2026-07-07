package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type stashTheme struct{}

func (stashTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 47, G: 111, B: 237, A: 255}
	case theme.ColorNameBackground:
		return color.NRGBA{R: 248, G: 249, B: 251, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 232, G: 236, B: 242, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 132, G: 143, B: 158, A: 255}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 47, G: 111, B: 237, A: 90}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 32, G: 36, B: 42, A: 255}
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 237, G: 241, B: 247, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 196, G: 205, B: 218, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 112, G: 122, B: 136, A: 255}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 221, G: 229, B: 240, A: 255}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 174, G: 185, B: 199, A: 255}
	case theme.ColorNameScrollBarBackground:
		return color.NRGBA{R: 238, G: 242, B: 247, A: 255}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 218, G: 224, B: 232, A: 255}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 15, G: 23, B: 42, A: 28}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
}

func (stashTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (stashTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (stashTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
