package xcf

import (
	"image"
	"math"
)

var srgbToLinear = func() []float64 {
	rv := []float64{}
	for i := range 256 {
		c := float64(i) / 0xff
		if c <= 0.04045 {
			rv = append(rv, c/12.92)
		} else {
			rv = append(rv, math.Pow((c+0.055)/1.055, 2.4))
		}
	}
	return rv
}()

func linearToSrgb(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

func clamp(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 255
	}
	return uint8(v*255.0 + 0.5)
}

func composite(layers []*layer, width int, height int) (image.Image, error) {
	// linear float buffer: [R, G, B, A] per pixel
	stride := width * 4
	buf := make([]float64, stride*height)

	for _, layer := range layers {
		img, err := layer.toImage()
		if err != nil {
			return nil, err
		}

		alpha := float64(layer.opacity) / 0xff
		b := layer.bounds.Intersect(image.Rect(0, 0, width, height))
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				sc := img.NRGBAAt(x, y)
				sR := srgbToLinear[sc.R]
				sG := srgbToLinear[sc.G]
				sB := srgbToLinear[sc.B]
				sA := float64(sc.A) / 0xff * alpha

				i := y*stride + x*4
				dR := buf[i+0]
				dG := buf[i+1]
				dB := buf[i+2]
				dA := buf[i+3]

				outA := sA + dA*(1-sA)
				if outA > 0 {
					buf[i+0] = (sR*sA + dR*dA*(1-sA)) / outA
					buf[i+1] = (sG*sA + dG*dA*(1-sA)) / outA
					buf[i+2] = (sB*sA + dB*dA*(1-sA)) / outA
				}
				buf[i+3] = outA
			}
		}
	}

	rv := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			fi := y*stride + x*4
			pi := y*rv.Stride + x*4
			rv.Pix[pi+0] = clamp(linearToSrgb(buf[fi+0]))
			rv.Pix[pi+1] = clamp(linearToSrgb(buf[fi+1]))
			rv.Pix[pi+2] = clamp(linearToSrgb(buf[fi+2]))
			rv.Pix[pi+3] = clamp(buf[fi+3])
		}
	}
	return rv, nil
}
