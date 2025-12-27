package project

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"rafaelmartins.com/p/website/internal/markdown"
)

var (
	pcProjectKey     = parser.NewContextKey()
	pcBaseUrlKey     = parser.NewContextKey()
	pcCurrentPageKey = parser.NewContextKey()
	pcTitleKey       = parser.NewContextKey()
	pcImagesKey      = parser.NewContextKey()
	pcErrorKey       = parser.NewContextKey()

	gmMarkdown = markdown.New("github", &extension{})
)

type extension struct{}

func (e *extension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(e, 0)))
}

func (extension) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	if pc == nil || pc.Get(pcProjectKey) == nil {
		return
	}

	if fc := node.FirstChild(); fc != nil && fc.Kind() == ast.KindHeading {
		pc.Set(pcTitleKey, string(fc.(*ast.Heading).Lines().Value(reader.Source())))
		node.RemoveChild(node, fc)
	}

	proj := pc.Get(pcProjectKey).(*Project)
	baseurl := pc.Get(pcBaseUrlKey).(string)
	images := []string{}

	pc.Set(pcErrorKey, ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindImage:
			img := n.(*ast.Image)
			p, u, err := proj.handleImageUrl(string(img.Destination), pc.Get(pcCurrentPageKey).(string))
			if err != nil {
				return 0, err
			}

			if p != "" {
				images = append(images, p)
			}
			if u != "" {
				img.Destination = []byte(u)
			}

		case ast.KindLink:
			link := n.(*ast.Link)
			gh, u, err := proj.handleLinkUrl(string(link.Destination), pc.Get(pcCurrentPageKey).(string))
			if err != nil {
				return 0, err
			}

			if u != "" {
				if gh {
					link.Destination = []byte(baseurl + "/" + u)
				} else {
					link.Destination = []byte(u)
				}
			}
		}

		return ast.WalkContinue, nil
	}))

	pc.Set(pcImagesKey, images)
}
