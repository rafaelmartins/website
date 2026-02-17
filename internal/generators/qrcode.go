package generators

import (
	"bytes"
	"errors"
	"image/png"
	"io"
	"os"

	"github.com/skip2/go-qrcode"
	"rafaelmartins.com/p/website/internal/hexcolor"
	"rafaelmartins.com/p/website/internal/runner"
)

type QRCode struct {
	File            string
	Content         string
	Size            int
	ForegroundColor *string
	BackgroundColor *string
	WithoutBorders  bool
}

func (QRCode) GetID() string {
	return "QRCODE"
}

func (s *QRCode) GetReader() (io.ReadCloser, error) {
	if s.File != "" && s.Content != "" {
		return nil, errors.New("qrcode: can't set file and content at the same time")
	}

	c := s.Content
	if s.File != "" {
		fp, err := os.Open(s.File)
		if err != nil {
			return nil, err
		}
		defer fp.Close()

		cb, err := io.ReadAll(fp)
		if err != nil {
			return nil, err
		}
		c = string(cb)
	}

	qr, err := qrcode.New(c, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	if s.ForegroundColor != nil {
		cc, err := hexcolor.ToRGBA(*s.ForegroundColor)
		if err != nil {
			return nil, err
		}
		qr.ForegroundColor = cc
	}

	if s.BackgroundColor != nil {
		cc, err := hexcolor.ToRGBA(*s.BackgroundColor)
		if err != nil {
			return nil, err
		}
		qr.BackgroundColor = cc
	}

	qr.DisableBorder = s.WithoutBorders

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, qr.Image(s.Size)); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (s *QRCode) GetPaths() ([]string, error) {
	if s.File == "" {
		return nil, nil
	}

	return []string{s.File}, nil
}

func (QRCode) GetImmutable() bool {
	return false
}

func (QRCode) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
