package ogfont

import (
	"io"
	"sync"

	"github.com/rafaelmartins/website/internal/github"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

const (
	fontOwner = "vernnobile"
	fontRepo  = "NunitoFont"
	fontFile  = "version-2.0/Nunito-ExtraBold.ttf"
)

type Font struct {
	m    sync.Mutex
	font *sfnt.Font
}

func New() (*Font, error) {
	resp, _, err := github.Contents(nil, fontOwner, fontRepo, fontFile, true)
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
