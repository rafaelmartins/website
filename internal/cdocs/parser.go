package cdocs

import (
	"io"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
)

var (
	lex = lexer.MustStateful(lexer.Rules{
		"Root": {
			{Name: "CommentOpenDoc", Pattern: `/\*\* ?`, Action: lexer.Push("Comment")},
			{Name: "CommentOpen", Pattern: `/\* ?`, Action: lexer.Push("Comment")},
			{Name: "CommentClose", Pattern: ` \*/`},
			{Name: "CommentContinuation", Pattern: ` \* ?`, Action: lexer.Push("Comment")},
			{Name: "PreProcessor", Pattern: `\#`, Action: lexer.Push("PreProcessor")},
			{Name: "Ident", Pattern: `[a-zA-Z0-9_]+`, Action: lexer.Push("Declaration")},
			{Name: "comment", Pattern: `//[^/\n\r]*`},
			{Name: "NewLine", Pattern: `\n|\r\n`},
			{Name: "whitespace", Pattern: `\s+`},
		},

		"Comment": {
			{Name: "CommentCommand", Pattern: `@[a-z\[\]\{\},]+`},
			{Name: "CommentValue", Pattern: `[^\s]+`},
			{Name: "CommentNewLine", Pattern: `\n|\r\n`, Action: lexer.Pop()},
			{Name: "Whitespace", Pattern: `\s+`},
		},

		"PreProcessor": {
			{Name: "PreprocessorNameInclude", Pattern: `include`, Action: lexer.Push("Include")},
			{Name: "PreprocessorNameDefine", Pattern: `define`, Action: lexer.Push("Define")},
			{Name: "PreProcessorName", Pattern: `pragma|ifdef|ifndef|elif|if|error`},
			{Name: "PreProcessorNameNoValue", Pattern: `else|endif`},
			{Name: "PreProcessorContinuation", Pattern: `\\(\n|\r\n)`},
			{Name: "PreProcessorValue", Pattern: `[^\n\r\\]+`},
			{Name: "PreProcessorNewLine", Pattern: `\n|\r\n`, Action: lexer.Pop()},
		},

		"Include": {
			{Name: "IncludeLocal", Pattern: `"`},
			{Name: "IncludeSystemOpen", Pattern: `<`},
			{Name: "IncludeSystemClose", Pattern: `>`},
			{Name: "IncludeWhitespace", Pattern: `[ \t]+`},
			{Name: "IncludeFile", Pattern: `[^\n\r\<\>"]+`},
			lexer.Return(),
		},

		"Define": {
			{Name: "DefineName", Pattern: `[a-zA-Z0-9_]+`, Action: lexer.Pop()},
			{Name: "DefineWhitespace", Pattern: `[ \t]+`},
		},

		"Declaration": {
			{Name: "Ident", Pattern: `[a-zA-Z0-9_]+`},
			{Name: "Dot", Pattern: `\.`},
			{Name: "Star", Pattern: `\*`},
			{Name: "Comma", Pattern: `,`},
			{Name: "Semicolon", Pattern: `;`, Action: lexer.Pop()},
			{Name: "ArgsOpen", Pattern: `\(`},
			{Name: "ArgsClose", Pattern: `\)`},
			{Name: "BracesOpen", Pattern: `\{`, Action: lexer.Push("Members")},
			{Name: "NewLine", Pattern: `\n|\r\n`},
			{Name: "Whitespace", Pattern: `\s+`},
		},

		"Members": {
			{Name: "Members", Pattern: `[^\}]+`},
			{Name: "BracesClose", Pattern: `\}`, Action: lexer.Pop()},
		},
	})
	parser = participle.MustBuild[Header](participle.Lexer(lex), participle.UseLookahead(100))
)

type Header struct {
	Pos lexer.Position

	Entries []Entry `parser:"@@*"`
}

func (h *Header) Dump(w io.Writer) {
	repr.New(w).Println(h)
}

type Entry struct {
	Pos lexer.Position

	Comment      *Comment      `parser:"  @@"`
	Include      *Include      `parser:"| @@"`
	Define       *Define       `parser:"| @@"`
	PreProcessor *PreProcessor `parser:"| @@"`
	Declaration  *Declaration  `parser:"| @@"`
	NewLine      *string       `parser:"| @NewLine"`
}

type Comment struct {
	Pos lexer.Position

	Lines []CommentLine `parser:"@@+"`
}

type CommentLine struct {
	Pos lexer.Position

	OpenDoc      *string  `parser:"( (  @CommentOpenDoc"`
	Open         *string  `parser:"   | @CommentOpen"`
	Continuation *string  `parser:"   | @CommentContinuation"`
	ValueTokens  []string `parser:"  ) (@CommentCommand|@CommentValue|Whitespace)* CommentNewLine"`
	Close        *string  `parser:") | @CommentClose"`
}

type Include struct {
	Pos lexer.Position

	Local  *string `parser:"'#' 'include' IncludeWhitespace (  @IncludeLocal      (?= IncludeFile IncludeLocal )"`
	System *string `parser:"                                 | @IncludeSystemOpen (?= IncludeFile IncludeSystemClose ) )"`
	File   string  `parser:"@IncludeFile (IncludeLocal | IncludeSystemClose) PreProcessorNewLine"`
}

type Define struct {
	Pos lexer.Position

	Name  string   `parser:"'#' 'define' DefineWhitespace @DefineName"`
	Lines []string `parser:"(@PreProcessorValue? (PreProcessorNewLine|PreProcessorContinuation))*"`
}

type PreProcessor struct {
	Pos lexer.Position

	Name  string   `parser:"'#' (@PreProcessorNameNoValue PreProcessorNewLine|(@PreProcessorName"`
	Lines []string `parser:"(@PreProcessorValue (PreProcessorNewLine|PreProcessorContinuation))+))"`
}

type Declaration struct {
	Pos lexer.Position

	Struct       *Struct       `parser:"  @@"`
	Enum         *Enum         `parser:"| @@"`
	FunctionType *FunctionType `parser:"| @@"`
	Function     *Function     `parser:"| @@"`
}

type Struct struct {
	Pos lexer.Position

	PreMembers  string `parser:"@'typedef' @Whitespace+ @'struct' (@Ident|@Whitespace|@ArgsOpen|@ArgsClose|@Comma)* @BracesOpen"`
	Members     string `parser:"@Members"`
	PostMembers string `parser:"@BracesClose @Whitespace+"`
	Name        string `parser:"@Ident"`
	PostName    string `parser:"@Whitespace* @Semicolon"`
}

type Enum struct {
	Pos lexer.Position

	PreMembers  string `parser:"@'typedef' @Whitespace+ @'enum' (@Ident|@Whitespace|@ArgsOpen|@ArgsClose|@Comma)* @BracesOpen"`
	Members     string `parser:"@Members"`
	PostMembers string `parser:"@BracesClose @Whitespace+"`
	Name        string `parser:"@Ident?"`
	PostName    string `parser:"@Whitespace* @Semicolon"`
}

type Member struct {
	Pos lexer.Position

	Member  string  `parser:"@NewLine? @Member"`
	Comment *string `parser:"(MemberComment @CommentValue (?= NewLine))?"`
}

type FunctionType struct {
	Pos lexer.Position

	PreName  string `parser:"@'typedef' @Whitespace+ (@Ident @Star* @Whitespace+ @Star* @Whitespace* @Dot*)+ @ArgsOpen @Star"`
	Name     string `parser:"@Ident"`
	PostName string `parser:"@ArgsClose @Whitespace*"`
	Args     string `parser:"@ArgsOpen (@Ident|@Whitespace|@NewLine|@Star|@Comma|@Dot)* @ArgsClose"`
	PostArgs string `parser:"(@Ident|@Whitespace|@NewLine|@Star|@ArgsOpen|@ArgsClose|@Comma)* @Semicolon"`
}

type Function struct {
	Pos lexer.Position

	PreName  string `parser:"(?= (Ident Star* Whitespace+ Star* Whitespace*)+ Ident ArgsOpen) (@Ident @Star* @Whitespace+ @Star* @Whitespace* @Dot*)+"`
	Name     string `parser:"@Ident"`
	Args     string `parser:"@ArgsOpen (@Ident|@Whitespace|@NewLine|@Star|@Comma|@Dot)* @ArgsClose"`
	PostArgs string `parser:"(@Ident|@Whitespace|@NewLine|@Star|@ArgsOpen|@ArgsClose|@Comma)* @Semicolon"`
}

func Parse(filename string, r io.ReadCloser) (*Header, error) {
	defer r.Close()

	ast, err := parser.Parse(filename, r)
	return ast, err
}
