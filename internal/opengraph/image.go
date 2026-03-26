package opengraph

import (
	"bytes"
	"errors"
	"html"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strings"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"rafaelmartins.com/p/website/internal/github"
	"rafaelmartins.com/p/website/internal/hexcolor"
	"rafaelmartins.com/p/website/internal/xcf"
)

const (
	fontGithubOwner = "googlefonts"
	fontGithubRepo  = "atkinson-hyperlegible-next"
	fontGithubFile  = "fonts/ttf/AtkinsonHyperlegibleNext-ExtraBold.ttf"
	fontGithubRef   = "main"
)

type ImageGenConfig struct {
	Template string `yaml:"template"`
	Font     struct {
		GithubOwner *string `yaml:"github-owner"`
		GithubRepo  *string `yaml:"github-repo"`
		GithubFile  *string `yaml:"github-file"`
		GithubRef   *string `yaml:"github-ref"`
	} `yaml:"font"`
	DefaultColor *string  `yaml:"default-color"`
	DefaultDPI   *float64 `yaml:"default-dpi"`
	DefaultSize  *float64 `yaml:"default-size"`
}

type OpenGraphImageGen struct {
	template string
	font     *sfnt.Font
	image    image.Image
	mask     image.Rectangle

	dcolor color.Color
	ddpi   float64
	dsize  float64
}

func NewImageGen(config *ImageGenConfig) (*OpenGraphImageGen, error) {
	if config == nil {
		return nil, nil
	}

	rv := &OpenGraphImageGen{
		template: config.Template,

		dcolor: color.Black,
		ddpi:   72,
		dsize:  96,
	}
	if config.Template == "" {
		return rv, nil
	}
	if config.DefaultColor != nil {
		c, err := hexcolor.ToRGBA(*config.DefaultColor)
		if err != nil {
			return nil, err
		}
		rv.dcolor = c
	}
	if config.DefaultDPI != nil {
		rv.ddpi = *config.DefaultDPI
	}
	if config.DefaultSize != nil {
		rv.dsize = *config.DefaultSize
	}

	owner := fontGithubOwner
	repo := fontGithubRepo
	file := fontGithubFile
	ref := fontGithubRef
	if config.Font.GithubOwner != nil {
		owner = *config.Font.GithubOwner
	}
	if config.Font.GithubRepo != nil {
		repo = *config.Font.GithubRepo
	}
	if config.Font.GithubFile != nil {
		file = *config.Font.GithubFile
	}
	if config.Font.GithubRef != nil {
		ref = *config.Font.GithubRef
	}

	resp, err := github.GetRepositoryFile(owner, repo, file, ref)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	fontBytes, err := io.ReadAll(resp)
	if err != nil {
		return nil, err
	}

	rv.font, err = opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	fp, err := os.Open(config.Template)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	rv.image, rv.mask, err = xcf.Decode(fp)
	if err != nil {
		return nil, err
	}
	return rv, nil
}

func (ogi *OpenGraphImageGen) GetPaths() []string {
	rv := []string{}
	if ogi.template != "" {
		rv = append(rv, ogi.template)
	}
	return rv
}

func (ogi *OpenGraphImageGen) Generate(og *OpenGraph) (io.ReadCloser, error) {
	if og == nil || ogi.template == "" || ogi.image == nil {
		return nil, errors.New("opengraph: image generator not initialized")
	}

	if strings.ContainsAny(og.title, "\t\n\r") {
		return nil, errors.New("opengraph: invalid whitespace characters found in text")
	}

	face, err := opentype.NewFace(ogi.font, &opentype.FaceOptions{
		Size:    og.size,
		DPI:     og.dpi,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return nil, err
	}
	defer face.Close()

	dst := image.NewRGBA(ogi.image.Bounds())
	draw.Copy(dst, image.Pt(0, 0), ogi.image, ogi.image.Bounds(), draw.Src, nil)

	d := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(og.c),
		Face: face,
	}

	lines, height, err := imageTitleSplit(html.UnescapeString(og.title), face, ogi.mask)
	if err != nil {
		return nil, err
	}

	y := ogi.mask.Min.Y + imageTitleFaceHeight(face) + (ogi.mask.Dy()-height)/2
	for _, line := range lines {
		x := ogi.mask.Min.X + (ogi.mask.Dx()-d.MeasureString(line).Ceil())/2
		d.Dot = fixed.P(x, y)
		d.DrawString(line)
		y += imageTitleFaceHeight(face) + imageTitleFaceSpacing(face)
	}

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, dst); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}
