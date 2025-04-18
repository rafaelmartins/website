package ogimage

import (
	"errors"
	"image"
	"math"
	"strings"

	"golang.org/x/image/font"
)

var (
	errTitleTooLongWidth  = errors.New("ogimage: title is too long (width)")
	errTitleTooLongHeight = errors.New("ogimage: title is too long (height)")
)

func titleFaceHeight(face font.Face) int {
	return face.Metrics().Ascent.Ceil()
}

func titleFaceSpacing(face font.Face) int {
	return int(math.Ceil(float64(titleFaceHeight(face)) * 0.125))
}

func titleSplit(text string, face font.Face, mask image.Rectangle) ([]string, int, error) {
	height := 0
	lines := []string{}
	if text == "" {
		return lines, height, nil
	}

	line := ""
	for part := range strings.SplitSeq(text, " ") {
		if l := font.MeasureString(face, part).Ceil(); l > mask.Dx() {
			return nil, 0, errTitleTooLongWidth
		}

		tline := ""
		if line == "" {
			tline = part
		} else {
			tline = line + " " + part
		}

		if l := font.MeasureString(face, tline).Ceil(); l > mask.Dx() {
			lines = append(lines, line)
			line = part
		} else {
			line = tline
		}
	}
	if line != "" {
		lines = append(lines, line)
	}

	height = len(lines)*titleFaceHeight(face) + (len(lines)-1)*titleFaceSpacing(face)
	if height > mask.Dy() {
		return nil, 0, errTitleTooLongHeight
	}

	return lines, height, nil
}
