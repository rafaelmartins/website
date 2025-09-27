package kicad

import (
	"image"
	"image/png"
	"io"

	"golang.org/x/image/draw"
)

func resize(w io.Writer, r io.ReadSeeker, scale int) error {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return err
	}

	src, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	srect := src.Bounds()

	dwidth := 0
	dheight := 0
	if srect.Dx() > srect.Dy() {
		dwidth = scale * srect.Dx() / srect.Dy()
		dheight = scale
	} else {
		dwidth = scale
		dheight = scale * srect.Dy() / srect.Dx()
	}
	drect := image.Rect(0, 0, dwidth, dheight)

	dst := image.NewRGBA(drect)
	draw.BiLinear.Scale(dst, drect, src, srect, draw.Src, nil)
	return png.Encode(w, dst)
}
