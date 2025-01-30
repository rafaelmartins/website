package generators

import (
	"bytes"
	"errors"
	"image/color"
	"image/png"
	"io"
	"os"
	"time"

	"github.com/rafaelmartins/website/internal/runner"
	"github.com/skip2/go-qrcode"
)

type QRCode struct {
	File            string
	Content         string
	Size            int
	ForegroundColor *uint32
	BackgroundColor *uint32
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
		qr.ForegroundColor = color.RGBA{
			R: byte(*s.ForegroundColor >> 24),
			G: byte(*s.ForegroundColor >> 16),
			B: byte(*s.ForegroundColor >> 8),
			A: byte(*s.ForegroundColor),
		}
	}

	if s.BackgroundColor != nil {
		qr.BackgroundColor = color.RGBA{
			R: byte(*s.BackgroundColor >> 24),
			G: byte(*s.BackgroundColor >> 16),
			B: byte(*s.BackgroundColor >> 8),
			A: byte(*s.BackgroundColor),
		}
	}

	qr.DisableBorder = s.WithoutBorders

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, qr.Image(s.Size)); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (s *QRCode) GetTimeStamps() ([]time.Time, error) {
	if s.File == "" {
		return nil, nil
	}

	st, err := os.Stat(s.File)
	if err != nil {
		return nil, err
	}
	return []time.Time{st.ModTime().UTC()}, nil
}

func (QRCode) GetImmutable() bool {
	return false
}

func (QRCode) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
