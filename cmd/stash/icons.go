package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"fyne.io/fyne/v2"
)

// Outline icons drawn on a 24px grid. Fyne only recolors SVG fills, so each
// icon is rendered per stroke color and cached.
var iconPaths = map[string]string{
	"text":     `<path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z"/><path d="M14 3v5h5M9 13h6M9 17h6M9 9h2"/>`,
	"image":    `<rect x="3" y="3" width="18" height="18" rx="2.5"/><circle cx="9" cy="9" r="1.75"/><path d="M21 15.5l-5-5L5 21"/>`,
	"settings": gearPath() + `<circle cx="12" cy="12" r="3"/>`,
	"search":   `<circle cx="11" cy="11" r="6.5"/><path d="M16 16l4.5 4.5"/>`,
	"copy":     `<rect x="8.5" y="8.5" width="12" height="12" rx="2"/><path d="M15.5 8.5V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7.5a2 2 0 0 0 2 2h2.5"/>`,
	"trash":    `<path d="M4 7h16M10 11v6M14 11v6M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12M9 7V4.5a.5.5 0 0 1 .5-.5h5a.5.5 0 0 1 .5.5V7"/>`,
	"plus":     `<path d="M12 5v14M5 12h14"/>`,
	"clearAll": `<path d="M4 6h16M4 12h9M4 18h6M15 14.5l5 5M20 14.5l-5 5"/>`,
	"check":    `<path d="M5 12.5l4.5 4.5L19 7.5"/>`,
	"paste":    `<rect x="8.5" y="2.5" width="7" height="4" rx="1"/><path d="M15.5 4.5H17a2 2 0 0 1 2 2V20a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6.5a2 2 0 0 1 2-2h1.5"/>`,
}

var iconCache = map[string]fyne.Resource{}

func icon(name string, col color.NRGBA) fyne.Resource {
	hex := fmt.Sprintf("#%02x%02x%02x", col.R, col.G, col.B)
	key := name + hex
	if res, ok := iconCache[key]; ok {
		return res
	}
	body, ok := iconPaths[name]
	if !ok {
		panic("unknown icon " + name)
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="%s" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">%s</svg>`, hex, body)
	res := fyne.NewStaticResource(key+".svg", []byte(svg))
	iconCache[key] = res
	return res
}

// gearPath traces an eight-tooth cog outline.
func gearPath() string {
	const teeth = 8
	var b strings.Builder
	b.WriteString(`<path d="`)
	for i := 0; i < teeth*4; i++ {
		radius := 9.0
		if i%4 >= 2 {
			radius = 7.0
		}
		angle := (float64(i) + 0.5) * math.Pi * 2 / (teeth * 4)
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&b, "%s%.2f %.2f", cmd, 12+radius*math.Cos(angle), 12+radius*math.Sin(angle))
	}
	b.WriteString(`Z"/>`)
	return b.String()
}
