package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type nodeSpoiler struct {
	ast.BaseInline
}

var kindSpoiler = ast.NewNodeKind("Spoiler")

func (n *nodeSpoiler) Kind() ast.NodeKind {
	return kindSpoiler
}

func (n *nodeSpoiler) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

type spoilerDelimiterProcessor struct{}

func (p *spoilerDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '|'
}

func (p *spoilerDelimiterProcessor) CanOpenCloser(opener *parser.Delimiter, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *spoilerDelimiterProcessor) OnMatch(consumes int) ast.Node {
	return &nodeSpoiler{}
}

type spoilerParser struct{}

func (s *spoilerParser) Trigger() []byte {
	return []byte{'|'}
}

func (s *spoilerParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 2, &spoilerDelimiterProcessor{})
	if node == nil || node.OriginalLength != 2 || before == '|' {
		return nil
	}

	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

type spoilerHTMLRenderer struct{}

func (r *spoilerHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindSpoiler, r.renderSpoiler)
}

func (r *spoilerHTMLRenderer) renderSpoiler(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<span class=\"spoiler\">")
		return ast.WalkContinue, nil
	}

	w.WriteString("</span>")
	return ast.WalkContinue, nil
}

type spoilers struct{}

func (e *spoilers) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&spoilerParser{}, 500),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&spoilerHTMLRenderer{}, 500),
		),
	)
}
