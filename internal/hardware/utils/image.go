package utils

import (
	"errors"
	"image"
	"image/png"
	"io"

	"golang.org/x/image/draw"
)

func Resize(w io.Writer, src image.Image, scale int) error {
	dwidth := 0
	dheight := 0
	srect := src.Bounds()
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

func Montage(imgs []image.Image) (image.Image, error) {
	if len(imgs) == 0 {
		return nil, errors.New("kicad: montage: no images found")
	}
	if len(imgs) == 1 {
		return imgs[0], nil
	}

	b1 := imgs[0].Bounds()
	b2 := imgs[1].Bounds()

	if b1.Dx() != b2.Dx() || b1.Dy() != b2.Dy() {
		return nil, errors.New("kicad: montage: image dimensions must match")
	}

	type srcType struct {
		img  image.Image
		orig image.Point
	}

	srcs := []*srcType{}
	dstRect := image.Rect(0, 0, 0, 0)
	for _, img := range imgs {
		if b1.Dy() >= b1.Dx() {
			srcs = append(srcs, &srcType{
				img: img,
				orig: image.Point{
					X: dstRect.Max.X,
				},
			})
			if dstRect.Max.Y == 0 {
				dstRect.Max.Y = img.Bounds().Dy()
			}
			dstRect.Max.X += img.Bounds().Dx()
			continue
		}

		srcs = append(srcs, &srcType{
			img: img,
			orig: image.Point{
				Y: dstRect.Max.Y,
			},
		})
		if dstRect.Max.X == 0 {
			dstRect.Max.X = img.Bounds().Dx()
		}
		dstRect.Max.Y += img.Bounds().Dy()
	}

	rv := image.NewRGBA(dstRect)
	for _, src := range srcs {
		draw.Copy(rv, src.orig, src.img, src.img.Bounds(), draw.Src, nil)
	}
	return rv, nil
}
