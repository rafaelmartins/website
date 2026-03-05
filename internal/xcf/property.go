package xcf

import (
	"encoding/binary"
	"io"
)

type property struct {
	ti int
	v  []byte
}

func newProperty(rs io.ReadSeeker) (*property, error) {
	hdr := struct {
		Ti  uint32
		Len uint32
	}{}
	if err := binary.Read(rs, binary.BigEndian, &hdr); err != nil {
		return nil, err
	}

	v := make([]byte, hdr.Len)
	if _, err := rs.Read(v); err != nil {
		return nil, err
	}

	return &property{
		ti: int(hdr.Ti),
		v:  v,
	}, nil
}

func (p *property) decode(v any) error {
	_, err := binary.Decode(p.v, binary.BigEndian, v)
	return err
}

func forEachProperty(rs io.ReadSeeker, f func(*property) error) error {
	for {
		prop, err := newProperty(rs)
		if err != nil {
			return err
		}
		if prop.ti == 0 {
			return nil
		}

		if err := f(prop); err != nil {
			return err
		}
	}
}
