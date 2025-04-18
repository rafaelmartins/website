package ogimage

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rafaelmartins/website/internal/ogfont"
	"github.com/rafaelmartins/website/internal/runner"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	tmpl string
	fnt  *ogfont.Font
	img  image.Image
	mask image.Rectangle

	dColor color.Color
	dDPI   float64
	dSize  float64

	available bool
)

func SetGlobals(template string, minX *int, minY *int, maxX *int, maxY *int, defaultColor *uint32, defaultDPI *float64, defaultSize *float64) error {
	available = template != ""
	if !available {
		return nil
	}

	if fnt == nil {
		f, err := ogfont.New()
		if err != nil {
			return err
		}
		fnt = f
	}

	fp, err := os.Open(template)
	if err != nil {
		return err
	}
	defer fp.Close()

	iimg, _, err := image.Decode(fp)
	if err != nil {
		return err
	}
	tmpl = template

	mmin := iimg.Bounds().Min
	if minX != nil {
		mmin.X = *minX
	}
	if minY != nil {
		mmin.Y = *minY
	}

	mmax := iimg.Bounds().Max
	if maxX != nil {
		mmax.X = *maxX
	}
	if maxY != nil {
		mmax.Y = *maxY
	}

	mmask := image.Rectangle{
		Min: mmin,
		Max: mmax,
	}
	if mmask.Min.X > mmask.Max.X || mmask.Min.Y > mmask.Max.Y {
		return fmt.Errorf("ogimage: bad mask rectangle: %v", mask)
	}

	ddpi := float64(72)
	if defaultDPI != nil {
		ddpi = *defaultDPI
	}

	ssize := float64(96)
	if defaultSize != nil {
		ssize = *defaultSize
	}

	dcolor := color.Color(color.Black)
	if defaultColor != nil {
		dcolor = color.RGBA{
			R: byte(*defaultColor >> 24),
			G: byte(*defaultColor >> 16),
			B: byte(*defaultColor >> 8),
			A: byte(*defaultColor),
		}
	}

	img = iimg
	mask = mmask
	dColor = dcolor
	dDPI = ddpi
	dSize = ssize
	return nil
}

func GetTimeStamps() ([]time.Time, error) {
	rv := []time.Time{}
	if !available {
		return rv, nil
	}

	st, err := os.Stat(tmpl)
	if err != nil {
		return nil, err
	}
	rv = append(rv, st.ModTime().UTC())
	return rv, nil
}

func Generate(text string, c color.Color, dpi *float64, size *float64) (io.ReadCloser, error) {
	if !available || img == nil {
		return nil, errors.New("ogimage: not initialized")
	}

	if strings.ContainsAny(text, "\t\n\r") {
		return nil, errors.New("ogimage: invalid whitespace characters found in text")
	}

	ddpi := dDPI
	if dpi != nil {
		ddpi = *dpi
	}

	ssize := dSize
	if size != nil {
		ssize = *size
	}

	face, err := fnt.GetFace(ssize, ddpi)
	if err != nil {
		return nil, err
	}
	defer face.Close()

	dst := image.NewRGBA(img.Bounds())
	draw.Copy(dst, image.Pt(0, 0), img, img.Bounds(), draw.Src, nil)

	d := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(c),
		Face: face,
	}

	lines, height, err := titleSplit(text, face, mask)
	if err != nil {
		return nil, err
	}

	y := mask.Min.Y + titleFaceHeight(face) + (mask.Dy()-height)/2
	for _, line := range lines {
		x := mask.Min.X + (mask.Dx()-d.MeasureString(line).Ceil())/2
		d.Dot = fixed.P(x, y)
		d.DrawString(line)
		y += titleFaceHeight(face) + titleFaceSpacing(face)
	}

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, dst); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func GenerateByProduct(ch chan *runner.GeneratorByProduct, title string, generate bool, image string, c *uint32, dpi *float64, size *float64) {
	if ch == nil || !generate || !available {
		return
	}

	if image != "" {
		rd, err := os.Open(image)
		if err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
		} else {
			ch <- &runner.GeneratorByProduct{
				Filename: "opengraph.png",
				Reader:   rd,
			}
		}
		return
	}

	dcolor := dColor
	if c != nil {
		dcolor = color.RGBA{
			R: byte(*c >> 24),
			G: byte(*c >> 16),
			B: byte(*c >> 8),
			A: byte(*c),
		}
	}

	rd, err := Generate(html.UnescapeString(title), dcolor, dpi, size)
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
	} else {
		ch <- &runner.GeneratorByProduct{
			Filename: "opengraph.png",
			Reader:   rd,
		}
	}
}

func URL(baseurl string) string {
	if baseurl == "" || !available {
		return ""
	}

	if strings.HasSuffix(baseurl, "/") {
		return path.Join(baseurl, "opengraph.png")
	}

	if strings.HasSuffix(baseurl, "/index.html") {
		return path.Join(path.Dir(baseurl), "opengraph.png")
	}

	tmp := path.Base(baseurl)
	tmp = strings.TrimSuffix(tmp, path.Ext(tmp))
	return path.Join(path.Dir(baseurl), tmp, "opengraph.png")
}
