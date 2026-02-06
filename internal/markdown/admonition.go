package markdown

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type nodeAdmonition struct {
	ast.BaseBlock
	Level string
}

var kindAdmonition = ast.NewNodeKind("Admonition")

func (n *nodeAdmonition) Kind() ast.NodeKind {
	return kindAdmonition
}

func (n *nodeAdmonition) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Level": n.Level,
	}, nil)
}

func newAdmonition(level string) *nodeAdmonition {
	return &nodeAdmonition{
		BaseBlock: ast.BaseBlock{},
		Level:     level,
	}
}

type admonitionParser struct {
	parser.BlockParser
}

func newAdmonitionParser() parser.BlockParser {
	return &admonitionParser{
		BlockParser: parser.NewBlockquoteParser(),
	}
}

var admonitionPattern = regexp.MustCompile(`^>\s*\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*$`)

func (b *admonitionParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	match := admonitionPattern.FindSubmatch(line)
	if match != nil {
		reader.AdvanceToEOL()
		return newAdmonition(string(match[1])), parser.HasChildren
	}
	return b.BlockParser.Open(parent, reader, pc)
}

type admonitionHTMLRenderer struct{}

func newAdmonitionHTMLRenderer() renderer.NodeRenderer {
	return &admonitionHTMLRenderer{}
}

func (r *admonitionHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindAdmonition, r.renderAdmonition)
}

func (r *admonitionHTMLRenderer) renderAdmonition(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*nodeAdmonition)
	levelLower := strings.ToLower(n.Level)
	title := strings.ToUpper(levelLower[:1]) + strings.ToLower(levelLower[1:])

	if entering {
		w.WriteString("<div class=\"admonition admonition-")
		w.WriteString(levelLower)
		w.WriteString("\">\n")
		w.WriteString("<p class=\"admonition-title\">")

		switch n.Level {
		case "NOTE":
			w.WriteString("<i class=\"fas fa-circle-info\"></i>")

		case "TIP":
			w.WriteString("<i class=\"fas fa-lightbulb\"></i>")

		case "IMPORTANT":
			w.WriteString("<i class=\"fas fa-circle-exclamation\"></i>")

		case "WARNING":
			w.WriteString("<i class=\"fas fa-triangle-exclamation\"></i>")

		case "CAUTION":
			w.WriteString("<i class=\"fas fa-circle-xmark\"></i>")
		}

		w.WriteString(title)
		w.WriteString("</p>\n")
		return ast.WalkContinue, nil
	}

	w.WriteString("</div>\n")
	return ast.WalkContinue, nil
}

type admonitions struct{}

func (e *admonitions) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(newAdmonitionParser(), 100),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(newAdmonitionHTMLRenderer(), 100),
		),
	)
}
