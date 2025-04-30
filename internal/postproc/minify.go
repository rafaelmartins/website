package postproc

import (
	"fmt"
	"io"
	"path/filepath"
	"slices"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

type mimeType struct {
	name       string
	extensions []string
}

var mimeTypes = []*mimeType{
	{"application/json", []string{".json"}},
	{"image/svg+xml", []string{".svg"}},
	{"text/css", []string{".css"}},
	{"text/html", []string{".html", ".htm"}},
	{"text/javascript", []string{".js", ".jsm"}},
	{"text/xml", []string{".xml", ".xsl", ".rss", ".xslt", ".xsd", ".wsdl", ".wsf", ".atom"}},
}

type Minify struct {
	m *minify.M
}

func (Minify) Supported(ext string) bool {
	for _, mt := range mimeTypes {
		if slices.Contains(mt.extensions, ext) {
			return true
		}
	}
	return false
}

func (m *Minify) Run(dstFn string, dst io.Writer, src io.Reader) error {
	if m.m == nil {
		m.m = minify.New()

		m.m.Add("application/json", &json.Minifier{})
		m.m.Add("image/svg+xml", &svg.Minifier{
			KeepComments: false,
			Precision:    0,
		})
		m.m.Add("text/css", &css.Minifier{
			Precision: 0,
			KeepCSS2:  true,
		})
		m.m.Add("text/html", &html.Minifier{
			KeepDocumentTags:    true,
			KeepSpecialComments: true,
			KeepEndTags:         true,
			KeepDefaultAttrVals: true,
			KeepWhitespace:      false,
		})
		m.m.Add("text/javascript", &js.Minifier{
			Version: 2022,
		})
		m.m.Add("text/xml", &xml.Minifier{
			KeepWhitespace: false,
		})
	}

	ext := filepath.Ext(dstFn)

	mtt := ""
	if ext != "" {
		for _, mt := range mimeTypes {
			if slices.Contains(mt.extensions, ext) {
				mtt = mt.name
				break
			}
		}
	}

	if mtt == "" {
		return fmt.Errorf("postproc: minify: invalid extension: %s", ext)
	}
	return m.m.Minify(mtt, dst, src)
}
