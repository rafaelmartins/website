package xcf

import (
	"encoding/binary"
	"io"
	"slices"
)

type pointer uint64

func newPointer(rs io.ReadSeeker, version int) (pointer, error) {
	ptr := any(new(uint32))
	if version >= 11 {
		ptr = new(uint64)
	}
	if err := binary.Read(rs, binary.BigEndian, ptr); err != nil {
		return 0, err
	}

	if version < 11 {
		return pointer(*ptr.(*uint32)), nil
	}
	return pointer(*ptr.(*uint64)), nil
}

func (p pointer) dereference(rs io.ReadSeeker) error {
	_, err := rs.Seek(int64(p), io.SeekStart)
	return err
}

type pointers struct {
	rs io.ReadSeeker
	p  []pointer
}

func newPointers(rs io.ReadSeeker, version int) (*pointers, error) {
	rv := &pointers{
		rs: rs,
	}

	for {
		ptr, err := newPointer(rs, version)
		if err != nil {
			return nil, err
		}
		if ptr == 0 {
			return rv, nil
		}

		rv.p = append(rv.p, ptr)
	}
}

func (p *pointers) reverse() {
	slices.Reverse(p.p)
}

func (p *pointers) forEach(f func(i int) error) error {
	for i, ptr := range p.p {
		if err := ptr.dereference(p.rs); err != nil {
			return err
		}

		if err := f(i); err != nil {
			return err
		}
	}
	return nil
}
