//go:build darwin

// Package appearance pins the native window chrome to the light look, so the
// title bar matches Stash's light-only theme even when macOS is in dark mode.
package appearance

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>

static void stash_use_light_appearance(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		NSApp.appearance = [NSAppearance appearanceNamed:NSAppearanceNameAqua];
	});
}
*/
import "C"

func UseLight() {
	C.stash_use_light_appearance()
}
