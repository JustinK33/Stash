package imageclipboard

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestClipboardRoundTrip(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	src.SetNRGBA(1, 1, color.NRGBA{G: 255, A: 255})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, src); err != nil {
		t.Fatal(err)
	}

	if err := WritePNG(encoded.Bytes()); err != nil {
		t.Fatal(err)
	}
	data, err := ReadPNG()
	if err != nil {
		t.Fatal(err)
	}
	got, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			if r1, g1, b1, a1 := got.At(x, y).RGBA(); [4]uint32{r1, g1, b1, a1} != rgba(src.At(x, y)) {
				t.Fatalf("pixel (%d,%d) = %v, want %v", x, y, got.At(x, y), src.At(x, y))
			}
		}
	}
}

func rgba(c color.Color) [4]uint32 {
	r, g, b, a := c.RGBA()
	return [4]uint32{r, g, b, a}
}
