package opengraph

import (
	"errors"
	"image"
	"math"
	"path"
	"strings"

	"golang.org/x/image/font"
)

var (
	errTitleTooLongWidth  = errors.New("opengraph: title is too long (width)")
	errTitleTooLongHeight = errors.New("opengraph: title is too long (height)")
)

func imageTitleFaceHeight(face font.Face) int {
	return face.Metrics().Ascent.Ceil()
}

func imageTitleFaceSpacing(face font.Face) int {
	return int(math.Ceil(float64(imageTitleFaceHeight(face)) * 0.125))
}

func imageTitleSplit(text string, face font.Face, mask image.Rectangle) ([]string, int, error) {
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

	height = len(lines)*imageTitleFaceHeight(face) + (len(lines)-1)*imageTitleFaceSpacing(face)
	if height > mask.Dy() {
		return nil, 0, errTitleTooLongHeight
	}

	return lines, height, nil
}

func url(baseurl string) string {
	if baseurl == "" {
		return ""
	}

	if strings.HasSuffix(baseurl, "/") {
		return path.Join(baseurl, "opengraph.png")
	}

	if strings.HasSuffix(baseurl, "/index.html") {
		return path.Join(path.Dir(baseurl), "opengraph.png")
	}

	tmp := path.Base(baseurl)
	tmp = strings.TrimSuffix(tmp, path.Ext(tmp))
	return path.Join(path.Dir(baseurl), tmp, "opengraph.png")
}
