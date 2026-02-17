package hexcolor

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"strings"
)

func ToRGBA(s string) (color.RGBA, error) {
	if !strings.HasPrefix(s, "#") {
		return color.RGBA{}, fmt.Errorf("hexcolor: must start with #: %s", s)
	}
	s = strings.TrimPrefix(s, "#")

	ss := strings.Builder{}
	switch len(s) {
	case 3, 4:
		for _, c := range s {
			fmt.Fprintf(&ss, "%c%c", c, c)
		}

	case 6, 8:
		ss.WriteString(s)

	default:
		return color.RGBA{}, fmt.Errorf("hexcolor: invalid color: %s: size mismatch", s)
	}

	b, err := hex.DecodeString(ss.String())
	if err != nil {
		return color.RGBA{}, fmt.Errorf("hexcolor: invalid color: %s: %w", s, err)
	}

	if len(b) == 3 {
		b = append(b, 0xff)
	}
	return color.RGBA{R: b[0], G: b[1], B: b[2], A: b[3]}, nil
}
