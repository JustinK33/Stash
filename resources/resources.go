// Package resources embeds the app icon and bundled fonts.
package resources

import _ "embed"

// AppIcon is the full macOS icon, with the standard transparent margin.
//
//go:embed icons/stash.png
var AppIcon []byte

// Logo is the icon cropped to its rounded square, for in-app branding.
//
//go:embed icons/logo.png
var Logo []byte

//go:embed fonts/Inter-Regular.ttf
var InterRegular []byte

//go:embed fonts/Inter-SemiBold.ttf
var InterSemiBold []byte
