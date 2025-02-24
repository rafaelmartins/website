package content

import (
	"fmt"
	"io"
	"time"

	"github.com/rafaelmartins/website/internal/content/frontmatter"
)

type contentProvider interface {
	IsSupported(f string) bool
	Render(f string, style string, baseurl string) (string, *frontmatter.FrontMatter, error)
	GetTimeStamps(f string) ([]time.Time, error)
	ListAssets(f string) ([]string, error)
	OpenAsset(f string, a string) (string, io.ReadCloser, error)
}

var providers = []contentProvider{
	&markdown{},
	&textBundle{},
	&textPack{},
	&html{},
}

func getProvider(f string) contentProvider {
	for _, p := range providers {
		if p.IsSupported(f) {
			return p
		}
	}
	return nil
}

func IsSupported(f string) bool {
	return getProvider(f) != nil
}

func Render(f string, style string, baseurl string) (string, *frontmatter.FrontMatter, error) {
	p := getProvider(f)
	if p == nil {
		return "", nil, fmt.Errorf("content: no provider found: %s", f)
	}
	return p.Render(f, style, baseurl)
}

func GetTimeStamps(f string) ([]time.Time, error) {
	p := getProvider(f)
	if p == nil {
		return nil, fmt.Errorf("content: no provider found: %s", f)
	}
	return p.GetTimeStamps(f)
}

func ListAssets(f string) ([]string, error) {
	p := getProvider(f)
	if p == nil {
		return nil, fmt.Errorf("content: no provider found: %s", f)
	}
	return p.ListAssets(f)
}

func OpenAsset(f string, a string) (string, io.ReadCloser, error) {
	p := getProvider(f)
	if p == nil {
		return "", nil, fmt.Errorf("content: no provider found: %s", f)
	}
	return p.OpenAsset(f, a)
}

func GetMetadata(f string) (*frontmatter.FrontMatter, error) {
	_, md, err := Render(f, "", "")
	return md, err
}
