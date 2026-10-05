package imageclipboard

import (
	"bytes"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	cfDIB         = 8
	cfHDROP       = 15
	gmemMoveable  = 0x0002
	openRetries   = 10
	openRetryWait = 20 * time.Millisecond
)

var (
	user32                         = syscall.NewLazyDLL("user32.dll")
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	shell32                        = syscall.NewLazyDLL("shell32.dll")
	procOpenClipboard              = user32.NewProc("OpenClipboard")
	procCloseClipboard             = user32.NewProc("CloseClipboard")
	procEmptyClipboard             = user32.NewProc("EmptyClipboard")
	procGetClipboardData           = user32.NewProc("GetClipboardData")
	procSetClipboardData           = user32.NewProc("SetClipboardData")
	procRegisterClipboardFormatW   = user32.NewProc("RegisterClipboardFormatW")
	procIsClipboardFormatAvailable = user32.NewProc("IsClipboardFormatAvailable")
	procGlobalAlloc                = kernel32.NewProc("GlobalAlloc")
	procGlobalFree                 = kernel32.NewProc("GlobalFree")
	procGlobalLock                 = kernel32.NewProc("GlobalLock")
	procGlobalUnlock               = kernel32.NewProc("GlobalUnlock")
	procGlobalSize                 = kernel32.NewProc("GlobalSize")
	procRtlMoveMemory              = kernel32.NewProc("RtlMoveMemory")
	procDragQueryFileW             = shell32.NewProc("DragQueryFileW")
)

func pngFormat() uintptr {
	name, _ := syscall.UTF16PtrFromString("PNG")
	format, _, _ := procRegisterClipboardFormatW.Call(uintptr(unsafe.Pointer(name)))
	return format
}

// openClipboard retries because another app may hold the clipboard briefly.
func openClipboard() error {
	for i := 0; i < openRetries; i++ {
		if ok, _, _ := procOpenClipboard.Call(0); ok != 0 {
			return nil
		}
		time.Sleep(openRetryWait)
	}
	return errors.New("clipboard is busy")
}

func ReadPNG() ([]byte, error) {
	if err := openClipboard(); err != nil {
		return nil, err
	}
	defer procCloseClipboard.Call()

	// Same order as macOS: native PNG, then a bitmap, then a copied file.
	if data := clipboardBytes(pngFormat()); data != nil {
		if _, err := png.DecodeConfig(bytes.NewReader(data)); err == nil {
			return data, nil
		}
	}
	if data := clipboardBytes(cfDIB); data != nil {
		if img, err := decodeDIB(data); err == nil {
			return encodePNG(img)
		}
	}
	if path := firstDroppedFile(); path != "" {
		if file, err := os.Open(path); err == nil {
			defer file.Close()
			if img, _, err := image.Decode(file); err == nil {
				return encodePNG(img)
			}
		}
	}
	return nil, errors.New("clipboard does not contain an image")
}

func WritePNG(data []byte) error {
	if len(data) == 0 {
		return errors.New("image is empty")
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return errors.New("could not copy image to clipboard")
	}

	if err := openClipboard(); err != nil {
		return err
	}
	defer procCloseClipboard.Call()

	if ok, _, _ := procEmptyClipboard.Call(); ok == 0 {
		return errors.New("could not copy image to clipboard")
	}
	// PNG keeps transparency for apps that read it; DIB is what Paint and
	// older apps understand.
	pngOK := setClipboardBytes(pngFormat(), data)
	dibOK := setClipboardBytes(cfDIB, encodeDIB(img))
	if !pngOK && !dibOK {
		return errors.New("could not copy image to clipboard")
	}
	return nil
}

func clipboardBytes(format uintptr) []byte {
	if format == 0 {
		return nil
	}
	if ok, _, _ := procIsClipboardFormatAvailable.Call(format); ok == 0 {
		return nil
	}
	handle, _, _ := procGetClipboardData.Call(format)
	if handle == 0 {
		return nil
	}
	size, _, _ := procGlobalSize.Call(handle)
	pointer, _, _ := procGlobalLock.Call(handle)
	if pointer == 0 || size == 0 {
		return nil
	}
	defer procGlobalUnlock.Call(handle)
	data := make([]byte, size)
	procRtlMoveMemory.Call(uintptr(unsafe.Pointer(&data[0])), pointer, size)
	return data
}

func setClipboardBytes(format uintptr, data []byte) bool {
	if format == 0 {
		return false
	}
	if len(data) == 0 {
		return false
	}
	handle, _, _ := procGlobalAlloc.Call(gmemMoveable, uintptr(len(data)))
	if handle == 0 {
		return false
	}
	pointer, _, _ := procGlobalLock.Call(handle)
	if pointer == 0 {
		procGlobalFree.Call(handle)
		return false
	}
	procRtlMoveMemory.Call(pointer, uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)))
	procGlobalUnlock.Call(handle)

	// On success the clipboard owns the memory; on failure we still do.
	if ok, _, _ := procSetClipboardData.Call(format, handle); ok == 0 {
		procGlobalFree.Call(handle)
		return false
	}
	return true
}

func firstDroppedFile() string {
	if ok, _, _ := procIsClipboardFormatAvailable.Call(cfHDROP); ok == 0 {
		return ""
	}
	handle, _, _ := procGetClipboardData.Call(cfHDROP)
	if handle == 0 {
		return ""
	}
	length, _, _ := procDragQueryFileW.Call(handle, 0, 0, 0)
	if length == 0 {
		return ""
	}
	buffer := make([]uint16, length+1)
	procDragQueryFileW.Call(handle, 0, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	return syscall.UTF16ToString(buffer)
}

func encodePNG(img image.Image) ([]byte, error) {
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
