package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"

	"rafaelmartins.com/p/website/internal/dfuse"
	"rafaelmartins.com/p/website/internal/github"
	"rafaelmartins.com/p/website/internal/http"
	"rafaelmartins.com/p/website/internal/runner"
)

type dfuFile struct {
	name    string
	version string
	typ     string
	data    *dfuse.DfuSe
}

type dfu struct {
	proj *Project

	rollingCtx http.RequestContext
	latestCtx  http.RequestContext

	files []*dfuFile
}

func (d *dfu) GetDestination() string {
	return filepath.Join(d.proj.Repo, d.proj.dfuDestination, "index.json")
}

func (d *dfu) GetGenerator() (runner.Generator, error) {
	return d, nil
}

func (*dfu) GetID() string {
	return "DFU"
}

func (d *dfu) processAssets(ctx *http.RequestContext, assets []github.RepositoryReleaseAsset) ([]*dfuFile, error) {
	p, err := regexp.Compile(d.proj.DfuReleaseAssetsPattern)
	if err != nil {
		return nil, err
	}

	rv := []*dfuFile{}
	for _, asset := range assets {
		matches := p.FindStringSubmatch(asset.Name)
		if len(matches) != 3 {
			continue
		}

		v := &dfuFile{
			name:    matches[1],
			version: matches[2],
		}
		resp, err := http.RequestWithContext(ctx, "GET", asset.DownloadUrl, nil, nil)
		if err != nil {
			return nil, err
		}

		v.data, err = dfuse.NewFromArchive(asset.Name, resp)
		if err != nil {
			return nil, err
		}

		rv = append(rv, v)
	}
	if len(rv) == 0 {
		return nil, fmt.Errorf("dfu: pattern not matched: %s", d.proj.DfuReleaseAssetsPattern)
	}
	return rv, nil
}

type dfuOutFirmware struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
	Url     string `json:"url"`
}

type dfuOut struct {
	Name      string           `json:"name"`
	Firmwares []dfuOutFirmware `json:"firmwares"`
}

func (d *dfu) GetReader() (io.ReadCloser, error) {
	if d.proj.proj.LatestRelease != nil {
		ff, err := d.processAssets(&d.latestCtx, d.proj.proj.LatestRelease.Assets)
		if err != nil {
			return nil, err
		}

		for _, f := range ff {
			if f.version == "" {
				f.version = d.proj.proj.LatestRelease.Tag
			}
			f.typ = "Latest Release"
			d.files = append(d.files, f)
		}
	}
	if d.proj.proj.RollingRelease != nil {
		ff, err := d.processAssets(&d.rollingCtx, d.proj.proj.RollingRelease.Assets)
		if err != nil {
			return nil, err
		}

		for _, f := range ff {
			if f.version == "" {
				f.version = d.proj.RollingTag
			}
			f.typ = "Rolling Release"
			d.files = append(d.files, f)
		}
	}

	v := &dfuOut{
		Name: d.proj.Repo,
	}
	for _, file := range d.files {
		u, err := url.JoinPath(d.proj.dfuUrl, file.name+"-"+file.version+".json")
		if err != nil {
			return nil, err
		}

		v.Firmwares = append(v.Firmwares, dfuOutFirmware{
			Name:    file.name,
			Version: file.version,
			Type:    file.typ,
			Url:     u,
		})
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (d *dfu) GetPaths() ([]string, error) {
	return nil, nil
}

func (d *dfu) GetImmutable() bool {
	return d.proj.Immutable
}

func (d *dfu) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch == nil {
		return
	}

	for _, file := range d.files {
		if file.data == nil {
			continue
		}

		rd, err := file.data.ToJson()
		if err != nil {
			ch <- &runner.GeneratorByProduct{
				Err: err,
			}
			return
		}

		ch <- &runner.GeneratorByProduct{
			Filename: file.name + "-" + file.version + ".json",
			Reader:   rd,
		}
	}
	close(ch)
}
