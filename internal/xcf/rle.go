package xcf

import (
	"fmt"
	"io"
)

func readByte(r io.Reader) (byte, error) {
	b := [1]byte{}
	if _, err := r.Read(b[:]); err != nil {
		return 0, err
	}
	return b[0], nil
}

func rleDecode(r io.Reader, v []byte) error {
	l := len(v)
	idx := 0
	for idx < l {
		n, err := readByte(r)
		if err != nil {
			return err
		}

		switch {
		case n <= 126:
			vv, err := readByte(r)
			if err != nil {
				return err
			}

			for range n + 1 {
				v[idx] = vv
				idx++
			}

		case n == 127:
			p, err := readByte(r)
			if err != nil {
				return err
			}
			q, err := readByte(r)
			if err != nil {
				return err
			}
			vv, err := readByte(r)
			if err != nil {
				return err
			}

			for range int(p)*256 + int(q) {
				v[idx] = vv
				idx++
			}

		case n == 128:
			p, err := readByte(r)
			if err != nil {
				return err
			}
			q, err := readByte(r)
			if err != nil {
				return err
			}

			nn, err := r.Read(v[idx : idx+int(p)*256+int(q)])
			if err != nil {
				return err
			}
			idx += nn

		default:
			nn, err := r.Read(v[idx : idx+256-int(n)])
			if err != nil {
				return err
			}
			idx += nn
		}
	}

	if idx != l {
		return fmt.Errorf("xcf: rle: failed to decompress: %d != %d", idx, l)
	}
	return nil
}
