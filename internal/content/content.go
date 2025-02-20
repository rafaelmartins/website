package content

import (
	"fmt"
	"io"
	"time"
)

type contentProvider interface {
	IsSupported(f string) bool
	Render(f string, style string, baseurl string) (string, *Metadata, error)
	ListAssets(f string) ([]string, error)
	ListAssetTimeStamps(f string) ([]time.Time, error)
	OpenAsset(f string, a string) (string, io.ReadCloser, error)
}

var providers = []contentProvider{
	&markdown{},
	&textBundle{},
	&textPack{},
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

func Render(f string, style string, baseurl string) (string, *Metadata, error) {
	p := getProvider(f)
	if p == nil {
		return "", nil, fmt.Errorf("content: render: no provider found: %s", f)
	}
	return p.Render(f, style, baseurl)
}

func ListAssets(f string) ([]string, error) {
	p := getProvider(f)
	if p == nil {
		return nil, fmt.Errorf("content: render: no provider found: %s", f)
	}
	return p.ListAssets(f)
}

func ListAssetTimeStamps(f string) ([]time.Time, error) {
	p := getProvider(f)
	if p == nil {
		return nil, fmt.Errorf("content: render: no provider found: %s", f)
	}
	return p.ListAssetTimeStamps(f)
}

func OpenAsset(f string, a string) (string, io.ReadCloser, error) {
	p := getProvider(f)
	if p == nil {
		return "", nil, fmt.Errorf("content: render: no provider found: %s", f)
	}
	return p.OpenAsset(f, a)
}

func GetMetadata(f string) (*Metadata, error) {
	_, md, err := Render(f, "", "")
	return md, err
}
