package xcf

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

type header struct {
	version     int
	width       int
	height      int
	compression byte
	layers      *pointers
}

func newHeader(rs io.ReadSeeker) (*header, error) {
	hdr := struct {
		Fti        [9]byte
		Version    [4]byte
		VersionSep byte
		Width      uint32
		Height     uint32
		ColorMode  uint32
		Precision  uint32
	}{}
	if err := binary.Read(rs, binary.BigEndian, &hdr); err != nil {
		return nil, err
	}

	if fti := string(hdr.Fti[:]); fti != "gimp xcf " {
		return nil, fmt.Errorf("xcf: header: unsupported file type identifier: %q", fti)
	}

	version := 0
	if hdr.Version[0] == 'v' {
		v, err := strconv.ParseInt(string(hdr.Version[1:]), 10, 32)
		if err != nil {
			return nil, err
		}
		version = int(v)
	}
	if version < 7 {
		return nil, fmt.Errorf("xcf: header: unsupported version: %d (must be at least 7)", version)
	}

	if hdr.ColorMode != 0 {
		return nil, fmt.Errorf("xcf: header: unsupported image color mode (must be RGB): %d", hdr.ColorMode)
	}

	if hdr.Precision != 150 {
		return nil, fmt.Errorf("xcf: header: unsupported image precision (must be 8-bit gamma integer): %d", hdr.Precision)
	}

	rv := header{
		version: version,
		width:   int(hdr.Width),
		height:  int(hdr.Height),
	}

	if err := forEachProperty(rs, func(p *property) error {
		if p.ti == 17 {
			if err := p.decode(&rv.compression); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if rv.compression != 1 {
		return nil, fmt.Errorf("xcf: header: unsupported tile compression (must be RLE): %d", rv.compression)
	}

	l, err := newPointers(rs, version)
	if err != nil {
		return nil, err
	}
	l.reverse()
	rv.layers = l

	return &rv, nil
}
