package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	PcTocEnable = parser.NewContextKey()
	pcTocItems  = parser.NewContextKey()
	pcTocError  = parser.NewContextKey()
)

func tocRender(pc parser.Context) (string, error) {
	if err := pc.Get(pcTocError); err != nil {
		return "", err.(error)
	}

	itemsV := pc.Get(pcTocItems)
	if itemsV == nil {
		return "", nil
	}
	items := itemsV.([]*tocItem)
	if len(items) == 0 {
		return "", nil
	}

	baselevel := items[0].level
	for _, item := range items[1:] {
		if item.level < baselevel {
			baselevel = item.level
		}
	}

	rv := strings.Builder{}
	rv.WriteString("<details id=\"__toc__\">\n<summary>Table of contents</summary>\n<ul>\n")

	currentlevel := baselevel
	for idx, item := range items {
		if idx == 0 && item.level > baselevel {
			for range item.level - baselevel {
				rv.WriteString("<li>\n<ul>\n")
			}
		} else if idx > 0 {
			if item.level > currentlevel {
				diff := item.level - currentlevel
				for i := range diff {
					rv.WriteString("\n<ul>\n")
					if i < diff-1 {
						rv.WriteString("<li>")
					}
				}
			} else if item.level < currentlevel {
				rv.WriteString("</li>\n")
				for range currentlevel - item.level {
					rv.WriteString("</ul>\n</li>\n")
				}
			} else {
				rv.WriteString("</li>\n")
			}
		}
		fmt.Fprintf(&rv, "<li><a href=\"#%s\">%s</a>", item.id, item.title)
		currentlevel = item.level
	}
	rv.WriteString("</li>\n")
	for range currentlevel - baselevel {
		rv.WriteString("</ul>\n</li>\n")
	}
	rv.WriteString("</ul>\n</details>\n")

	return rv.String(), nil
}

type tocItem struct {
	level int
	id    string
	title string
}

type toc struct{}

func (t *toc) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(t, 100)))
}

func (*toc) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	enable, _ := pc.Get(PcTocEnable).(*bool)
	if enable == nil || !*enable {
		return
	}

	items := []*tocItem{}
	pc.Set(pcTocError, ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindHeading {
			h := n.(*ast.Heading)
			id, ok := h.AttributeString("id")
			if !ok {
				return ast.WalkContinue, nil
			}

			buf := &bytes.Buffer{}
			for c := n.FirstChild(); c != nil; c = c.NextSibling() {
				if err := goldmark.DefaultRenderer().Render(buf, reader.Source(), c); err != nil {
					return ast.WalkStop, err
				}
			}

			faLink := ast.NewLink()
			faLink.SetAttributeString("class", []byte("toc-anchor has-text-link-light fa-solid fa-paragraph"))
			faLink.Destination = append([]byte{'#'}, id.([]byte)...)
			h.AppendChild(h, faLink)

			faLink = ast.NewLink()
			faLink.SetAttributeString("class", []byte("toc-anchor has-text-link-light fa-solid fa-arrow-turn-up"))
			faLink.Destination = []byte("#__toc__")
			h.AppendChild(h, faLink)

			items = append(items, &tocItem{
				level: h.Level,
				id:    string(id.([]byte)),
				title: buf.String(),
			})
		}
		return ast.WalkContinue, nil
	}))
	pc.Set(pcTocItems, items)
}
