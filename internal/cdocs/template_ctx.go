package cdocs

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
)

type commentParam struct {
	ident       string
	description string
}

type commentLine struct {
	pos lexer.Position

	file string

	brief  string
	normal string

	sectionName  string
	sectionOpen  bool
	sectionClose bool

	warning string
	returns string

	paramIn    commentParam
	paramOut   commentParam
	paramInOut commentParam
}

func newCommentLine(vtoks []string, pos lexer.Position) (*commentLine, error) {
	rv := &commentLine{
		pos: pos,
	}

	prependNext := ""
	appendNext := ""

	captureFile := false
	captureBrief := false
	captureName := false
	captureWarning := false
	captureReturns := false
	captureParamIn := false
	captureParamOut := false
	captureParamInOut := false
	captureRef := false

	lineTokens := []string{}

	for i, vtok := range vtoks {
		switch vtok {
		case "@file":
			if i != 0 {
				return nil, fmt.Errorf("cdocs: %s: @file not at start of line", pos)
			}
			if rv.file != "" || captureFile {
				return nil, fmt.Errorf("cdocs: %s: duplicated @file", pos)
			}
			captureFile = true

		case "@brief":
			if i != 0 {
				return nil, fmt.Errorf("cdocs: %s: @brief not at start of line", pos)
			}
			if rv.brief != "" || captureBrief {
				return nil, fmt.Errorf("cdocs: %s: duplicated @brief", pos)
			}
			captureBrief = true

		case "@name":
			if i != 0 {
				return nil, fmt.Errorf("cdocs: %s: @name not at start of line", pos)
			}
			if rv.sectionName != "" || captureName {
				return nil, fmt.Errorf("cdocs: %s: duplicated @name", pos)
			}
			captureName = true

		case "@warning":
			if i != 0 {
				return nil, fmt.Errorf("cdocs: %s: @warning not at start of line", pos)
			}
			if rv.warning != "" || captureWarning {
				return nil, fmt.Errorf("cdocs: %s: duplicated @warning", pos)
			}
			captureWarning = true

		case "@returns":
			if i != 0 {
				return nil, fmt.Errorf("cdocs: %s: @returns not at start of line", pos)
			}
			if rv.returns != "" || captureReturns {
				return nil, fmt.Errorf("cdocs: %s: duplicated @returns", pos)
			}
			captureReturns = true

		case "@param[in]":
			if i != 0 {
				return nil, fmt.Errorf("cdocs: %s: @param[in] not at start of line", pos)
			}
			if rv.paramIn.description != "" || captureParamIn {
				return nil, fmt.Errorf("cdocs: %s: duplicated @param[in]", pos)
			}
			captureParamIn = true

		case "@param[out]":
			if i != 0 {
				return nil, fmt.Errorf("cdocs: %s: @param[out] not at start of line", pos)
			}
			if rv.paramOut.description != "" || captureParamOut {
				return nil, fmt.Errorf("cdocs: %s: duplicated @param[out]", pos)
			}
			captureParamOut = true

		case "@param[in,out]":
			if i != 0 {
				return nil, fmt.Errorf("cdocs: %s: @param[in,out] not at start of line", pos)
			}
			if rv.paramInOut.description != "" || captureParamInOut {
				return nil, fmt.Errorf("cdocs: %s: duplicated @param[in,out]", pos)
			}
			captureParamInOut = true

		case "@ref":
			captureRef = true

		case "@b":
			prependNext = "<b>"
			appendNext = "</b>"

		case "@c":
			prependNext = "<code>"
			appendNext = "</code>"

		case "@{":
			rv.sectionOpen = true

		case "@}":
			rv.sectionClose = true

		default:
			if strings.HasPrefix(vtok, "@") {
				return nil, fmt.Errorf("cdocs: %s: unsupported command: %s", pos, vtok)
			}

			if captureParamIn && rv.paramIn.ident == "" {
				rv.paramIn.ident = vtok
				continue
			}
			if captureParamOut && rv.paramOut.ident == "" {
				rv.paramOut.ident = vtok
				continue
			}
			if captureParamInOut && rv.paramInOut.ident == "" {
				rv.paramInOut.ident = vtok
				continue
			}
			if captureRef {
				// split punctuation similar to what doxygen does
				r := vtok
				end := ""
				if strings.HasSuffix(r, ".") {
					r = strings.TrimRight(r, ".")
					end = "."
				} else if strings.HasSuffix(r, ",") {
					r = strings.TrimRight(r, ",")
					end = ","
				} else if strings.HasSuffix(r, ";") {
					r = strings.TrimRight(r, ";")
					end = ";"
				}

				lineTokens = append(lineTokens, `<a href="#`+id(r)+`">`+r+"</a>"+end)
				captureRef = false
				continue
			}

			lineTokens = append(lineTokens, prependNext+vtok+appendNext)
			prependNext = ""
			appendNext = ""
		}
	}

	if len(lineTokens) == 0 {
		return rv, nil
	}

	if captureFile {
		rv.file = strings.Join(lineTokens, " ")
		return rv, nil
	}

	if captureBrief {
		rv.brief = "<p>" + strings.Join(lineTokens, " ") + "</p>\n"
		return rv, nil
	}

	if captureName {
		rv.sectionName = strings.Join(lineTokens, " ")
		return rv, nil
	}

	if captureWarning {
		rv.warning = strings.Join(lineTokens, " ")
		return rv, nil
	}

	if captureReturns {
		rv.returns = "<p>" + strings.Join(lineTokens, " ") + "</p>\n"
		return rv, nil
	}

	if captureParamIn {
		rv.paramIn.description = "<p>" + strings.Join(lineTokens, " ") + "</p>\n"
		return rv, nil
	}

	if captureParamOut {
		rv.paramOut.description = "<p>" + strings.Join(lineTokens, " ") + "</p>\n"
		return rv, nil
	}

	if captureParamInOut {
		rv.paramInOut.description = "<p>" + strings.Join(lineTokens, " ") + "</p>\n"
		return rv, nil
	}

	rv.normal = strings.Join(lineTokens, " ")
	return rv, nil
}

type TemplateCtx struct {
	Headers []HeaderCtx
}

func (t *TemplateCtx) Dump(w io.Writer) {
	repr.New(w).Println(t)
}

type HeaderCtx struct {
	ID          string
	Name        string
	Description string

	Includes []string

	Sections []*SectionCtx

	Defines       []*EntryCtx
	Structs       []*EntryCtx
	Enums         []*EntryCtx
	Functions     []*EntryCtx
	FunctionTypes []*EntryCtx
}

type SectionCtx struct {
	ID          string
	Name        string
	Description string

	Defines       []*EntryCtx
	Structs       []*EntryCtx
	Enums         []*EntryCtx
	Functions     []*EntryCtx
	FunctionTypes []*EntryCtx
}

type EntryCtx struct {
	ID          string
	Type        string
	Name        string
	Proto       string
	Description string
	Link        string
}

type TemplateCtxHeader struct {
	Filename  string
	Header    *Header
	GithubUrl string
}

func NewTemplateCtx(headers []*TemplateCtxHeader) (*TemplateCtx, error) {
	rv := &TemplateCtx{}

	for _, hdr := range headers {
		hctx := HeaderCtx{
			Name: hdr.Filename,
		}

		link := func(line int) string {
			return fmt.Sprintf("%s#L%d", hdr.GithubUrl, line)
		}

		section := (*SectionCtx)(nil)
		pendingSection := (*SectionCtx)(nil)
		pendingDescription := ""

		for _, entry := range hdr.Header.Entries {
			if c := entry.Comment; c != nil {
				paragraphs := [][]*commentLine{}
				paragraph := []*commentLine{}

				for i, line := range c.Lines {
					if line.Open != nil {
						break
					}

					if i == 0 && line.OpenDoc == nil {
						return nil, fmt.Errorf("cdocs: %s: invalid start of documentation block", line.Pos)
					}
					if i == len(c.Lines)-1 && line.Close == nil {
						return nil, fmt.Errorf("cdocs: %s: invalid end of documentation block", line.Pos)
					}

					if l := len(line.ValueTokens); line.Continuation != nil && l == 0 {
						paragraphs = append(paragraphs, paragraph)
						paragraph = []*commentLine{}
					} else if l > 0 {
						cl, err := newCommentLine(line.ValueTokens, line.Pos)
						if err != nil {
							return nil, err
						}
						paragraph = append(paragraph, cl)
					}
				}
				if len(paragraph) > 0 {
					paragraphs = append(paragraphs, paragraph)
				}

				isFile := false
				isSection := false
				description := ""

				// brief is always the first thing
				for _, paragraph := range paragraphs {
					for _, line := range paragraph {
						if line.brief != "" {
							if description != "" {
								return nil, fmt.Errorf("cdocs: %s: @brief must be at start of comment block", line.pos)
							}
							description += line.brief + "\n"
						}
					}
				}

				for _, paragraph := range paragraphs {
					paragraphOpen := false
					paramsOpen := false
					paramsDone := false
					warningOpen := false

					for _, line := range paragraph {
						if paragraphOpen && line.normal == "" {
							description += "</p>\n"
							paragraphOpen = false
						}

						if paramsOpen && line.paramIn.ident == "" && line.paramOut.ident == "" && line.paramInOut.ident == "" {
							description += "</table></dd></dl>\n"
							paramsOpen = false
							paramsDone = true
						} else if !paramsOpen && (line.paramIn.ident != "" || line.paramOut.ident != "" || line.paramInOut.ident != "") {
							if paramsDone {
								return nil, fmt.Errorf("cdocs: %s: @param lines must be grouped together", line.pos)
							}
							description += `<dl class="arguments"><dt><b>Arguments</b></dt><dd><table class="arguments">`
							paramsOpen = true
						}

						if line.file != "" {
							hctx.ID = id(line.file)
							hctx.Name = line.file
							isFile = true
						}

						if line.sectionName != "" {
							pendingSection = &SectionCtx{
								ID:   id(line.sectionName),
								Name: line.sectionName,
							}
							isSection = true
						}
						if line.sectionOpen {
							section = pendingSection
							hctx.Sections = append(hctx.Sections, section)
							pendingSection = nil
						}
						if line.sectionClose {
							section = nil
							pendingSection = nil
						}

						if line.normal != "" {
							if !paragraphOpen && !warningOpen {
								description += "<p>\n"
								paragraphOpen = true
							}
							description += line.normal + "\n"
							continue
						}

						if line.paramIn.ident != "" {
							description += "<tr><th><code>[in] " + line.paramIn.ident + "</code></th><td>" + line.paramIn.description + "</td></tr>\n"
							continue
						}
						if line.paramOut.ident != "" {
							description += "<tr><th><code>[out] " + line.paramOut.ident + "</code></th><td>" + line.paramOut.description + "</td></tr>\n"
							continue
						}
						if line.paramInOut.ident != "" {
							description += "<tr><th><code>[in,out] " + line.paramInOut.ident + "</code></th><td>" + line.paramInOut.description + "</td></tr>\n"
							continue
						}

						if line.returns != "" {
							description += `<dl class="return"><dt><b>Returns</b></dt><dd>` + line.returns + "</dd></dl>\n"
							continue
						}

						if line.warning != "" {
							description += `<div class="notification is-warning is-light"><p>` + "\n" + line.warning + "\n"
							warningOpen = true
							continue
						}
					}

					if paragraphOpen {
						description += "</p>\n"
					}
					if paramsOpen {
						description += "</table></dd></dl>\n"
					}
					if warningOpen {
						description += "</p></div>\n"
					}
				}

				if isFile {
					hctx.Description = description
					continue
				}

				if isSection {
					if pendingSection != nil {
						pendingSection.Description = description
					} else if section != nil {
						section.Description = description
					}
					continue
				}

				pendingDescription = description
				continue
			}

			if inc := entry.Include; inc != nil {
				include := "#include "
				if inc.System != nil {
					include += "&lt;"
				} else {
					include += "\""
				}

				found := false
				for _, i := range headers {
					if i.Filename == inc.File {
						found = true
						break
					}
				}
				if found {
					include += `<a href="#` + id(inc.File) + `">` + inc.File + "</a>"
				} else {
					include += inc.File
				}

				if inc.System != nil {
					include += "&gt;"
				} else {
					include += "\""
				}
				hctx.Includes = append(hctx.Includes, include)
			}

			if def := entry.Define; def != nil {
				proto, err := highlight("#define " + def.Name + strings.Join(def.Lines, "\\\n"))
				if err != nil {
					return nil, err
				}

				ectx := &EntryCtx{
					ID:          id(def.Name),
					Type:        "Macro",
					Name:        def.Name,
					Proto:       proto,
					Description: pendingDescription,
					Link:        link(def.Pos.Line),
				}
				pendingDescription = ""
				if section != nil {
					section.Defines = append(section.Defines, ectx)
				} else {
					hctx.Defines = append(hctx.Defines, ectx)
				}
			}

			if pp := entry.PreProcessor; pp != nil {
				switch pp.Name {
				case "ifdef":
				case "ifndef":
				case "elif":
				case "if":
					// do nothing (for now)

				case "pragma":
				case "error":
					// do nothing
				}
				continue
			}

			if d := entry.Declaration; d != nil {
				if st := d.Struct; st != nil {
					proto, err := highlight(st.PreMembers + st.Members + st.PostMembers + st.Name + st.PostName)
					if err != nil {
						return nil, err
					}

					ectx := &EntryCtx{
						ID:          id(st.Name),
						Type:        "Struct",
						Name:        st.Name,
						Proto:       proto,
						Description: pendingDescription,
						Link:        link(st.Pos.Line),
					}
					pendingDescription = ""
					if section != nil {
						section.Structs = append(section.Structs, ectx)
					} else {
						hctx.Structs = append(hctx.Structs, ectx)
					}
					continue
				}

				if en := d.Enum; en != nil {
					proto, err := highlight(en.PreMembers + en.Members + en.PostMembers + en.Name + en.PostName)
					if err != nil {
						return nil, err
					}

					ectx := &EntryCtx{
						ID:          id(en.Name),
						Type:        "Enumeration",
						Name:        en.Name,
						Proto:       proto,
						Description: pendingDescription,
						Link:        link(en.Pos.Line),
					}
					pendingDescription = ""
					if section != nil {
						section.Enums = append(section.Enums, ectx)
					} else {
						hctx.Enums = append(hctx.Enums, ectx)
					}
					continue
				}

				if fn := d.Function; fn != nil {
					proto, err := highlight(fn.PreName + fn.Name + fn.Args + fn.PostArgs)
					if err != nil {
						return nil, err
					}

					ectx := &EntryCtx{
						ID:          id(fn.Name),
						Type:        "Function",
						Name:        fn.Name,
						Proto:       proto,
						Description: pendingDescription,
						Link:        link(fn.Pos.Line),
					}
					pendingDescription = ""
					if section != nil {
						section.Functions = append(section.Functions, ectx)
					} else {
						hctx.Functions = append(hctx.Functions, ectx)
					}
					continue
				}

				if fnt := d.FunctionType; fnt != nil {
					proto, err := highlight(fnt.PreName + fnt.Name + fnt.PostName + fnt.Args + fnt.PostArgs)
					if err != nil {
						return nil, err
					}

					ectx := &EntryCtx{
						ID:          id(fnt.Name),
						Type:        "Function type",
						Name:        fnt.Name,
						Proto:       proto,
						Description: pendingDescription,
						Link:        link(fnt.Pos.Line),
					}
					pendingDescription = ""
					if section != nil {
						section.FunctionTypes = append(section.FunctionTypes, ectx)
					} else {
						hctx.FunctionTypes = append(hctx.FunctionTypes, ectx)
					}
					continue
				}
			}
		}

		rv.Headers = append(rv.Headers, hctx)
	}

	return rv, nil
}
