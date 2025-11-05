package config

import (
	"errors"
	"io/fs"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Title  string `yaml:"title"`
	Footer string `yaml:"footer"`
	URL    string `yaml:"url"`

	Author struct {
		Name  string `yaml:"name"`
		Email string `yaml:"email"`
	} `yaml:"author"`

	TemplatePartials []string `yaml:"template-partials"`

	OpenGraphImageGen struct {
		Template string `yaml:"template"`
		Mask     struct {
			MinX *int `yaml:"min-x"`
			MinY *int `yaml:"min-y"`
			MaxX *int `yaml:"max-x"`
			MaxY *int `yaml:"max-y"`
		} `yaml:"mask"`
		DefaultColor *uint32  `yaml:"default-color"`
		DefaultDPI   *float64 `yaml:"default-dpi"`
		DefaultSize  *float64 `yaml:"default-size"`
	} `yaml:"opengraph-image-gen"`

	Assets struct {
		BaseDestination string `yaml:"base-destination"`
		Npm             []*struct {
			Name    string   `yaml:"name"`
			Version string   `yaml:"version"`
			Files   []string `yaml:"files"`
		} `yaml:"npm"`
	} `yaml:"assets"`

	Files []*struct {
		BaseDestination string   `yaml:"base-destination"`
		Paths           []string `yaml:"paths"`
	} `yaml:"files"`

	Menu []*struct {
		Title    string `yaml:"title"`
		URL      string `yaml:"url"`
		Dropdown []struct {
			Title   string `yaml:"title"`
			URL     string `yaml:"url"`
			Divider bool   `yaml:"divider"`
		} `yaml:"dropdown"`
	} `yaml:"menu"`

	SocialLinks []*struct {
		URL   string `yaml:"url"`
		Label string `yaml:"label"`
		Icon  string `yaml:"icon"`
	} `yaml:"social-links"`

	Projects []*struct {
		Repositories []*struct {
			Owner    string `yaml:"owner"`
			Repo     string `yaml:"repo"`
			SubPages []struct {
				SubPage     string `yaml:"subpage"`
				Template    string `yaml:"template"`
				WithSidebar *bool  `yaml:"with-sidebar"`
				OpenGraph   struct {
					Title       string `yaml:"title"`
					Description string `yaml:"description"`
					Image       string `yaml:"image"`
					ImageGen    struct {
						Color *uint32  `yaml:"color"`
						DPI   *float64 `yaml:"dpi"`
						Size  *float64 `yaml:"size"`
					} `yaml:"image-gen"`
				} `yaml:"opengraph"`
			} `yaml:"subpages"`
			CDocs struct {
				Destination    string   `yaml:"destination"`
				Headers        []string `yaml:"headers"`
				BaseDirectory  string   `yaml:"base-directory"`
				LocalDirectory string   `yaml:"local-directory"`
				Template       string   `yaml:"template"`
				WithSidebar    *bool    `yaml:"with-sidebar"`
				OpenGraph      struct {
					Title       string `yaml:"title"`
					Description string `yaml:"description"`
					Image       string `yaml:"image"`
					ImageGen    struct {
						Color *uint32  `yaml:"color"`
						DPI   *float64 `yaml:"dpi"`
						Size  *float64 `yaml:"size"`
					} `yaml:"image-gen"`
				} `yaml:"opengraph"`
			} `yaml:"c-docs"`
			KicadProjects []string `yaml:"kicad-projects"`
			DocLinks      []struct {
				URL   string `yaml:"url"`
				Label string `yaml:"label"`
				Icon  string `yaml:"icon"`
			} `yaml:"doc-links"`
			Go struct {
				Import string `yaml:"import"`
				Repo   string `yaml:"repo"`
			} `yaml:"go"`
			OpenGraph struct {
				Title       string `yaml:"title"`
				Description string `yaml:"description"`
				Image       string `yaml:"image"`
				ImageGen    struct {
					Color *uint32  `yaml:"color"`
					DPI   *float64 `yaml:"dpi"`
					Size  *float64 `yaml:"size"`
				} `yaml:"image-gen"`
			} `yaml:"opengraph"`
			Immutable *bool `yaml:"immutable"`
		} `yaml:"repositories"`
		BaseDestination string `yaml:"base-destination"`
		Template        string `yaml:"template"`
		WithSidebar     *bool  `yaml:"with-sidebar"`
	} `yaml:"projects"`

	Pages []*struct {
		Sources []*struct {
			Slug      string `yaml:"slug"`
			File      string `yaml:"file"`
			OpenGraph struct {
				Title       string `yaml:"title"`
				Description string `yaml:"description"`
				Image       string `yaml:"image"`
				ImageGen    struct {
					Generate *bool    `yaml:"generate"`
					Color    *uint32  `yaml:"color"`
					DPI      *float64 `yaml:"dpi"`
					Size     *float64 `yaml:"size"`
				} `yaml:"image-gen"`
			} `yaml:"opengraph"`
		} `yaml:"sources"`
		ExtraDependencies []string       `yaml:"extra-dependencies"`
		HighlightStyle    string         `yaml:"highlight-style"`
		PrettyURL         *bool          `yaml:"pretty-url"`
		BaseDestination   string         `yaml:"base-destination"`
		Template          string         `yaml:"template"`
		TemplateCtx       map[string]any `yaml:"template-context"`
		WithSidebar       bool           `yaml:"with-sidebar"`
	} `yaml:"pages"`

	Posts struct {
		Title              string         `yaml:"title"`
		Description        string         `yaml:"description"`
		PostsPerPage       int            `yaml:"posts-per-page"`
		PostsPerPageAtom   int            `yaml:"posts-per-page-atom"`
		SortReverse        *bool          `yaml:"sort-reverse"`
		HighlightStyle     string         `yaml:"highlight-style"`
		BaseDestination    string         `yaml:"base-destination"`
		TemplateAtom       string         `yaml:"template-atom"`
		TemplatePagination string         `yaml:"template-pagination"`
		TemplateCtx        map[string]any `yaml:"template-context"`
		WithSidebar        bool           `yaml:"with-sidebar"`
		OpenGraph          struct {
			Title       string `yaml:"title"`
			Description string `yaml:"description"`
			Image       string `yaml:"image"`
			ImageGen    struct {
				Color *uint32  `yaml:"color"`
				DPI   *float64 `yaml:"dpi"`
				Size  *float64 `yaml:"size"`
			} `yaml:"image-gen"`
		} `yaml:"opengraph"`

		Groups []*struct {
			Title              string         `yaml:"title"`
			Description        string         `yaml:"description"`
			SourceDir          string         `yaml:"source-dir"`
			PostsPerPage       int            `yaml:"posts-per-page"`
			PostsPerPageAtom   int            `yaml:"posts-per-page-atom"`
			SortReverse        *bool          `yaml:"sort-reverse"`
			HighlightStyle     string         `yaml:"highlight-style"`
			BaseDestination    string         `yaml:"base-destination"`
			Template           string         `yaml:"template"`
			TemplateAtom       string         `yaml:"template-atom"`
			TemplatePagination string         `yaml:"template-pagination"`
			TemplateCtx        map[string]any `yaml:"template-context"`
			WithSidebar        bool           `yaml:"with-sidebar"`
			OpenGraph          struct {
				Title       string `yaml:"title"`
				Description string `yaml:"description"`
				Image       string `yaml:"image"`
				ImageGen    struct {
					Color *uint32  `yaml:"color"`
					DPI   *float64 `yaml:"dpi"`
					Size  *float64 `yaml:"size"`
				} `yaml:"image-gen"`
			} `yaml:"opengraph"`
		} `yaml:"groups"`
	} `yaml:"posts"`

	QRCode []struct {
		SourceFile      string  `yaml:"source-file"`
		SourceContent   string  `yaml:"source-content"`
		DestinationFile string  `yaml:"destination-file"`
		Size            int     `yaml:"size"`
		ForegroundColor *uint32 `yaml:"foreground-color"`
		BackgroundColor *uint32 `yaml:"background-color"`
		WithoutBorders  bool    `yaml:"without-borders"`
	} `yaml:"qrcode"`

	file string
	ts   time.Time
}

func New(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	rv := &Config{
		file: file,
		ts:   st.ModTime().UTC(),
	}
	if err := dec.Decode(rv); err != nil {
		return nil, err
	}
	return rv, nil
}

func (c *Config) GetTimeStamp() (time.Time, error) {
	st, err := os.Stat(c.file)
	if err != nil {
		return time.Time{}, err
	}
	return st.ModTime().UTC(), nil
}

func (c *Config) IsUpToDate() bool {
	ts, err := c.GetTimeStamp()
	if err != nil || ts.Compare(c.ts) > 0 {
		return false
	}
	for _, pg := range c.Posts.Groups {
		st, err := os.Stat(pg.SourceDir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return false
		}
		if st.ModTime().UTC().Compare(c.ts) > 0 {
			return false
		}
	}
	return true
}
