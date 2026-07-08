//go:build darwin

package imageclipboard

/*
#cgo LDFLAGS: -framework AppKit -framework Foundation
#include <stdlib.h>

int stash_read_clipboard_png(unsigned char **data, long *length);
int stash_write_clipboard_png(const unsigned char *data, long length);
*/
import "C"

import (
	"errors"
	"unsafe"
)

func ReadPNG() ([]byte, error) {
	var data *C.uchar
	var length C.long
	if C.stash_read_clipboard_png(&data, &length) == 0 {
		return nil, errors.New("clipboard does not contain an image")
	}
	defer C.free(unsafe.Pointer(data))
	return C.GoBytes(unsafe.Pointer(data), C.int(length)), nil
}

func WritePNG(data []byte) error {
	if len(data) == 0 {
		return errors.New("image is empty")
	}
	if C.stash_write_clipboard_png((*C.uchar)(unsafe.Pointer(&data[0])), C.long(len(data))) == 0 {
		return errors.New("could not copy image to clipboard")
	}
	return nil
}
