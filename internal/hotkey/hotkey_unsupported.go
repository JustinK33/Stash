//go:build !darwin

package hotkey

import "errors"

func Register(onPressed func()) error {
	return errors.New("global shortcuts are only implemented on macOS")
}

func Unregister() {}
