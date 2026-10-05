package imageclipboard

import (
	"encoding/binary"
	"image"
	"image/color"
	"testing"
)

func TestDIBRoundTrip(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	src.SetNRGBA(2, 1, color.NRGBA{B: 255, A: 128})

	decoded, err := decodeDIB(encodeDIB(src))
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			if got, want := decoded.At(x, y), src.NRGBAAt(x, y); got != want {
				t.Fatalf("pixel (%d,%d) = %v, want %v", x, y, got, want)
			}
		}
	}
}

func TestDecodeDIBZeroAlphaIsOpaque(t *testing.T) {
	data := encodeDIB(image.NewNRGBA(image.Rect(0, 0, 1, 1)))
	copy(data[40:], []byte{10, 20, 30, 0})

	decoded, err := decodeDIB(data)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := decoded.At(0, 0), (color.NRGBA{R: 30, G: 20, B: 10, A: 255}); got != want {
		t.Fatalf("pixel = %v, want %v", got, want)
	}
}

func TestDecodeDIB24BitTopDownWithPadding(t *testing.T) {
	// 1x2, 24-bit rows padded to 4 bytes, negative height means top-down.
	data := make([]byte, 40+8)
	binary.LittleEndian.PutUint32(data[0:], 40)
	binary.LittleEndian.PutUint32(data[4:], 1)
	binary.LittleEndian.PutUint32(data[8:], uint32(0xffffffff-1)) // -2
	binary.LittleEndian.PutUint16(data[14:], 24)
	copy(data[40:], []byte{0, 0, 255, 0, 255, 0, 0, 0})

	decoded, err := decodeDIB(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := decoded.At(0, 0); got != (color.NRGBA{R: 255, A: 255}) {
		t.Fatalf("top pixel = %v, want red", got)
	}
	if got := decoded.At(0, 1); got != (color.NRGBA{B: 255, A: 255}) {
		t.Fatalf("bottom pixel = %v, want blue", got)
	}
}

func TestDecodeDIBBitfields(t *testing.T) {
	data := make([]byte, 40+12+4)
	binary.LittleEndian.PutUint32(data[0:], 40)
	binary.LittleEndian.PutUint32(data[4:], 1)
	binary.LittleEndian.PutUint32(data[8:], 1)
	binary.LittleEndian.PutUint16(data[14:], 32)
	binary.LittleEndian.PutUint32(data[16:], biBitfields)
	binary.LittleEndian.PutUint32(data[40:], 0x000000ff) // red in the low byte
	binary.LittleEndian.PutUint32(data[44:], 0x0000ff00)
	binary.LittleEndian.PutUint32(data[48:], 0x00ff0000)
	binary.LittleEndian.PutUint32(data[52:], 0x00000011)

	decoded, err := decodeDIB(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := decoded.At(0, 0); got != (color.NRGBA{R: 0x11, A: 255}) {
		t.Fatalf("pixel = %v, want R=0x11 opaque", got)
	}
}

func TestDecodeDIBRejectsTruncated(t *testing.T) {
	data := encodeDIB(image.NewNRGBA(image.Rect(0, 0, 4, 4)))
	if _, err := decodeDIB(data[:len(data)-1]); err == nil {
		t.Fatal("expected error for truncated pixels")
	}
}
