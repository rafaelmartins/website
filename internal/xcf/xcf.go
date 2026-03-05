package xcf

import (
	"fmt"
	"image"
	"io"
)

/*
 * implementation fully based on information from:
 * https://developer.gimp.org/core/standards/xcf/
 */

func Decode(rs io.ReadSeeker) (image.Image, image.Rectangle, error) {
	hdr, err := newHeader(rs)
	if err != nil {
		return nil, image.Rectangle{}, err
	}

	var mask *image.Rectangle
	var layers []*layer

	if err := hdr.layers.forEach(func(_ int) error {
		if mask != nil {
			return fmt.Errorf("xcf: mask layer must be the top-most layer")
		}

		l, err := hdr.newLayer(rs)
		if err != nil {
			return err
		}

		if !l.visible {
			return nil
		}

		if l.name == "mask" {
			mask = &l.bounds
			return nil
		}
		layers = append(layers, l)
		return nil
	}); err != nil {
		return nil, image.Rectangle{}, err
	}

	if mask == nil {
		return nil, image.Rectangle{}, fmt.Errorf("xcf: no mask layer defined. must be the top-most layer, named \"mask\"")
	}

	rv, err := composite(layers, hdr.width, hdr.height)
	if err != nil {
		return nil, image.Rectangle{}, err
	}
	return rv, *mask, nil
}
