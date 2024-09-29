package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Title  string `yaml:"title"`
	Footer string `yaml:"footer"`
	URL    string `yaml:"url"`

	Author struct {
		Name  string `yaml:"name"`
		Email string `yaml:"email"`
	} `yaml:"author"`

	Assets struct {
		BaseDestination string `yaml:"base-destination"`
		Cdnjs           []*struct {
			Name    string   `yaml:"name"`
			Version string   `yaml:"version"`
			Files   []string `yaml:"files"`
		} `yaml:"cdnjs"`
	} `yaml:"assets"`

	Files []*struct {
		BaseDestination string   `yaml:"base-destination"`
		Paths           []string `yaml:"paths"`
	} `yaml:"files"`

	Menu []*struct {
		Title string `yaml:"title"`
		URL   string `yaml:"url"`
	} `yaml:"menu"`

	SocialLinks []*struct {
		URL   string `yaml:"url"`
		Label string `yaml:"label"`
		Icon  string `yaml:"icon"`
	} `yaml:"social-links"`

	Projects []*struct {
		Owner           string `yaml:"owner"`
		Repo            string `yaml:"repo"`
		BaseDestination string `yaml:"base-destination"`
		Template        string `yaml:"template"`
		Immutable       *bool  `yaml:"immutable"`
		WithSidebar     bool   `yaml:"with-sidebar"`
	} `yaml:"projects"`

	Pages []*struct {
		Sources []*struct {
			Slug string `yaml:"slug"`
			File string `yaml:"file"`
		} `yaml:"sources"`
		ExtraDependencies []string               `yaml:"extra-dependencies"`
		HighlightStyle    string                 `yaml:"highlight-style"`
		BaseDestination   string                 `yaml:"base-destination"`
		Template          string                 `yaml:"template"`
		TemplateCtx       map[string]interface{} `yaml:"template-context"`
		WithSidebar       bool                   `yaml:"with-sidebar"`
	} `yaml:"pages"`

	Posts []*struct {
		Title              string                 `yaml:"title"`
		SourceDir          string                 `yaml:"source-dir"`
		PostsPerPage       *int                   `yaml:"posts-per-page"`
		PostsPerPageAtom   *int                   `yaml:"posts-per-page-atom"`
		SortReverse        *bool                  `yaml:"sort-reverse"`
		HighlightStyle     string                 `yaml:"highlight-style"`
		BaseDestination    string                 `yaml:"base-destination"`
		Template           string                 `yaml:"template"`
		TemplateAtom       string                 `yaml:"template-atom"`
		TemplatePagination string                 `yaml:"template-pagination"`
		TemplateCtx        map[string]interface{} `yaml:"template-context"`
		WithSidebar        bool                   `yaml:"with-sidebar"`
	} `yaml:"posts"`

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

	rv := &Config{
		file: file,
		ts:   st.ModTime().UTC(),
	}
	if err := yaml.NewDecoder(f).Decode(rv); err != nil {
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
	if err != nil {
		return false
	}
	return ts.Compare(c.ts) <= 0
}
