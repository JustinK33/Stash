//go:build !darwin

package imageclipboard

import "errors"

func ReadPNG() ([]byte, error) {
	return nil, errors.New("image clipboard is only supported on macOS")
}

func WritePNG([]byte) error {
	return errors.New("image clipboard is only supported on macOS")
}
