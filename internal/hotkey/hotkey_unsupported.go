//go:build !darwin

package hotkey

import (
	"errors"

	"stash/internal/keybind"
)

func Register(binding keybind.Binding, onPressed func()) error {
	return errors.New("global shortcuts are only implemented on macOS")
}

func Unregister() {}
