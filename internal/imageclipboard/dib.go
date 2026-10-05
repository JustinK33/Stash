package imageclipboard

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/draw"
)

// Windows clipboard bitmaps are DIBs: a BITMAPINFOHEADER (or a V4/V5 header)
// followed by optional color masks and the pixel rows. These helpers live
// outside the windows build so their tests run everywhere.

const (
	biRGB       = 0
	biBitfields = 3
)

func decodeDIB(data []byte) (image.Image, error) {
	if len(data) < 40 {
		return nil, errors.New("bitmap header is truncated")
	}
	headerSize := int(binary.LittleEndian.Uint32(data[0:]))
	width := int(int32(binary.LittleEndian.Uint32(data[4:])))
	height := int(int32(binary.LittleEndian.Uint32(data[8:])))
	bitCount := int(binary.LittleEndian.Uint16(data[14:]))
	compression := binary.LittleEndian.Uint32(data[16:])
	if headerSize < 40 || headerSize > len(data) {
		return nil, errors.New("bitmap header is invalid")
	}
	if bitCount != 24 && bitCount != 32 {
		return nil, errors.New("bitmap bit depth is unsupported")
	}
	if compression != biRGB && compression != biBitfields {
		return nil, errors.New("bitmap compression is unsupported")
	}

	topDown := height < 0
	if topDown {
		height = -height
	}
	if width <= 0 || height <= 0 || width > 1<<15 || height > 1<<15 {
		return nil, errors.New("bitmap size is invalid")
	}

	offset := headerSize
	masks := [4]uint32{0x00ff0000, 0x0000ff00, 0x000000ff, 0}
	if compression == biBitfields {
		if bitCount != 32 {
			return nil, errors.New("bitmap bit fields need 32 bits per pixel")
		}
		if headerSize == 40 {
			// A plain info header is followed by three RGB masks.
			if len(data) < 52 {
				return nil, errors.New("bitmap masks are truncated")
			}
			offset += 12
		}
		maskStart := 40
		for i := 0; i < 3; i++ {
			masks[i] = binary.LittleEndian.Uint32(data[maskStart+4*i:])
		}
		if headerSize >= 56 {
			masks[3] = binary.LittleEndian.Uint32(data[52:])
		}
	}

	stride := (width*bitCount/8 + 3) &^ 3
	if len(data)-offset < stride*height {
		return nil, errors.New("bitmap pixels are truncated")
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	hasAlpha := false
	for y := 0; y < height; y++ {
		row := data[offset+stride*y:]
		dstY := y
		if !topDown {
			dstY = height - 1 - y
		}
		for x := 0; x < width; x++ {
			var c color.NRGBA
			if bitCount == 24 {
				p := row[x*3:]
				c = color.NRGBA{R: p[2], G: p[1], B: p[0], A: 0xff}
			} else {
				p := binary.LittleEndian.Uint32(row[x*4:])
				alphaMask := masks[3]
				if compression == biRGB {
					alphaMask = 0xff000000
				}
				c = color.NRGBA{R: channel(p, masks[0]), G: channel(p, masks[1]), B: channel(p, masks[2]), A: channel(p, alphaMask)}
				if alphaMask == 0 {
					c.A = 0xff
				}
				hasAlpha = hasAlpha || c.A != 0
			}
			img.SetNRGBA(x, dstY, c)
		}
	}
	// Most apps leave the fourth byte of a 32-bit BI_RGB bitmap at zero, which
	// means "no alpha", not "fully transparent".
	if bitCount == 32 && !hasAlpha {
		for i := 3; i < len(img.Pix); i += 4 {
			img.Pix[i] = 0xff
		}
	}
	return img, nil
}

func channel(pixel, mask uint32) uint8 {
	if mask == 0 {
		return 0
	}
	shift := 0
	for mask&1 == 0 {
		mask >>= 1
		shift++
	}
	value := (pixel >> shift) & mask
	return uint8(value * 255 / mask)
}

// encodeDIB writes a bottom-up 32-bit BI_RGB DIB, the most widely read layout.
func encodeDIB(src image.Image) []byte {
	bounds := src.Bounds()
	img := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(img, img.Bounds(), src, bounds.Min, draw.Src)
	width, height := img.Bounds().Dx(), img.Bounds().Dy()

	data := make([]byte, 40+width*height*4)
	binary.LittleEndian.PutUint32(data[0:], 40)
	binary.LittleEndian.PutUint32(data[4:], uint32(width))
	binary.LittleEndian.PutUint32(data[8:], uint32(height))
	binary.LittleEndian.PutUint16(data[12:], 1)
	binary.LittleEndian.PutUint16(data[14:], 32)
	binary.LittleEndian.PutUint32(data[20:], uint32(width*height*4))

	pixels := data[40:]
	for y := 0; y < height; y++ {
		row := pixels[(height-1-y)*width*4:]
		for x := 0; x < width; x++ {
			s := img.Pix[y*img.Stride+x*4:]
			row[x*4+0], row[x*4+1], row[x*4+2], row[x*4+3] = s[2], s[1], s[0], s[3]
		}
	}
	return data
}
