package xcf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
)

type tile struct {
	x    int
	y    int
	w    int
	h    int
	data []byte
}

type layer struct {
	name    string
	opacity byte
	visible bool
	bounds  image.Rectangle
	bpp     int
	tiles   []*tile
}

func (h *header) newLayer(rs io.ReadSeeker) (*layer, error) {
	hdr := struct {
		Width   uint32
		Height  uint32
		Type    uint32
		NameLen uint32
	}{}
	if err := binary.Read(rs, binary.BigEndian, &hdr); err != nil {
		return nil, err
	}

	if hdr.Type != 0 && hdr.Type != 1 {
		return nil, fmt.Errorf("xcf: layer: unsupported type (must be RGB/RGBA): %d", hdr.Type)
	}

	name := make([]byte, hdr.NameLen)
	if err := binary.Read(rs, binary.BigEndian, &name); err != nil {
		return nil, err
	}
	if n := bytes.IndexByte(name, 0); n >= 0 {
		name = name[:n]
	}

	rv := layer{
		name:    string(name),
		opacity: 0xff,
	}

	if err := forEachProperty(rs, func(p *property) error {
		switch p.ti {
		case 6:
			v := uint32(0)
			if err := p.decode(&v); err != nil {
				return err
			}
			rv.opacity = byte(v)

		case 8:
			v := uint32(0)
			if err := p.decode(&v); err != nil {
				return err
			}
			rv.visible = v != 0

		case 15:
			v := struct {
				X int32
				Y int32
			}{}
			if err := p.decode(&v); err != nil {
				return err
			}
			rv.bounds = image.Rect(int(v.X), int(v.Y), int(v.X)+int(hdr.Width), int(v.Y)+int(hdr.Height))

		case 33:
			v := float32(0)
			if err := p.decode(&v); err != nil {
				return err
			}
			rv.opacity = byte(v * 0xff)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	hptr, err := newPointer(rs, h.version)
	if err != nil {
		return nil, err
	}

	if err := hptr.dereference(rs); err != nil {
		return nil, err
	}

	hhdr := struct {
		Width  uint32
		Height uint32
		Bpp    uint32
	}{}
	if err := binary.Read(rs, binary.BigEndian, &hhdr); err != nil {
		return nil, err
	}

	if hhdr.Width != uint32(rv.bounds.Dx()) || hhdr.Height != uint32(rv.bounds.Dy()) {
		return nil, fmt.Errorf("xcf: %s: invalid hierarchy dimensions: %dx%d != %dx%d", rv.name, hhdr.Width, hhdr.Height, rv.bounds.Dx(), rv.bounds.Dy())
	}
	if hhdr.Bpp < 3 || hhdr.Bpp > 4 {
		return nil, fmt.Errorf("xcf: %s: invalid bits per pixel: %d", rv.name, hhdr.Bpp)
	}
	rv.bpp = int(hhdr.Bpp)

	lptr, err := newPointer(rs, h.version)
	if err != nil {
		return nil, err
	}

	if err := lptr.dereference(rs); err != nil {
		return nil, err
	}

	lhdr := struct {
		Width  uint32
		Height uint32
	}{}
	if err := binary.Read(rs, binary.BigEndian, &lhdr); err != nil {
		return nil, err
	}

	if lhdr.Width != uint32(rv.bounds.Dx()) || lhdr.Height != uint32(rv.bounds.Dy()) {
		return nil, fmt.Errorf("xcf: %s: invalid layer dimensions: %dx%d != %dx%d", rv.name, lhdr.Width, lhdr.Height, rv.bounds.Dx(), rv.bounds.Dy())
	}

	tptrs, err := newPointers(rs, h.version)
	if err != nil {
		return nil, err
	}

	tx := (rv.bounds.Dx() + 63) / 64
	ty := (rv.bounds.Dy() + 63) / 64
	if err := tptrs.forEach(func(i int) error {
		t := tile{
			x: i % tx,
			y: i / tx,
			w: 64,
			h: 64,
		}
		if t.x >= tx || t.y >= ty {
			return nil
		}

		if (t.x+1)*64 > rv.bounds.Dx() {
			t.w = rv.bounds.Dx() - t.x*64
		}
		if (t.y+1)*64 > rv.bounds.Dy() {
			t.h = rv.bounds.Dy() - t.y*64
		}

		t.data = make([]byte, t.w*t.h*rv.bpp)
		if err := rleDecode(rs, t.data); err != nil {
			return err
		}
		rv.tiles = append(rv.tiles, &t)
		return nil
	}); err != nil {
		return nil, err
	}
	return &rv, nil
}

func (l *layer) toImage() (*image.NRGBA, error) {
	rv := image.NewNRGBA(l.bounds)
	for _, tile := range l.tiles {
		for yy := range tile.h {
			for xx := range tile.w {
				c := color.NRGBA{
					R: tile.data[yy*tile.w+xx],
					G: tile.data[yy*tile.w+xx+tile.w*tile.h],
					B: tile.data[yy*tile.w+xx+2*tile.w*tile.h],
					A: 0xff,
				}
				if l.bpp == 4 {
					c.A = tile.data[yy*tile.w+xx+3*tile.w*tile.h]
				}
				rv.Set(l.bounds.Min.X+tile.x*64+xx, l.bounds.Min.Y+tile.y*64+yy, c)
			}
		}
	}
	return rv, nil
}
