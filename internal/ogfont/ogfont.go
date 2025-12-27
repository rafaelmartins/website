package ogfont

import (
	"io"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"rafaelmartins.com/p/website/internal/github"
)

const (
	fontOwner = "googlefonts"
	fontRepo  = "atkinson-hyperlegible-next"
	fontFile  = "fonts/ttf/AtkinsonHyperlegibleNext-ExtraBold.ttf"
	fontRef   = "main"
)

type Font struct {
	m    sync.Mutex
	font *sfnt.Font
}

func New() (*Font, error) {
	resp, err := github.GetRepositoryFile(fontOwner, fontRepo, fontFile, fontRef)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	fontBytes, err := io.ReadAll(resp)
	if err != nil {
		return nil, err
	}

	fnt, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	return &Font{
		font: fnt,
	}, nil
}

func (f *Font) GetFace(size float64, dpi float64) (font.Face, error) {
	f.m.Lock()
	defer f.m.Unlock()

	return opentype.NewFace(f.font, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingNone,
	})
}
