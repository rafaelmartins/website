package github

import (
	"fmt"
	"io"
	"path"
	"path/filepath"
)

var getRepository = `
query GetRepository($owner: String!, $repo: String!, $headersref: String){
	repository(owner: $owner, name: $repo) {
		description
		homepageUrl
		defaultBranchRef {
			name
			target {
				oid
			}
		}
		head: object(expression: "HEAD") {
			... on Commit {
				oid
			}
		}
		forkCount
		stargazerCount
		watchers {
			totalCount
		}
		licenseInfo {
			spdxId
		}
		license: object(expression: "HEAD:LICENSE") {
			... on Blob {
				text
				isBinary
				isTruncated
			}
		}
		copying: object(expression: "HEAD:COPYING") {
			... on Blob {
				text
				isBinary
				isTruncated
			}
		}
		latestRelease {
			name
			tagName
			url
			description
			releaseAssets(first: 100) {
				nodes {
					name
					downloadUrl
				}
				totalCount
			}
		}
		releases(first: 100, orderBy:{field: NAME, direction: DESC}) {
			nodes {
				tagName
				isDraft
				isPrerelease
			}
			totalCount
		}
		readme: object(expression: "HEAD:README.md") {
			... on Blob {
				text
				isBinary
				isTruncated
			}
		}
		docs: object(expression: "HEAD:docs") {
			... on Tree {
				entries {
					name
					type
					object {
						... on Blob {
							text
							isBinary
							isTruncated
						}
					}
				}
			}
		}
		headers: object(expression: $headersref) {
			... on Tree {
				entries {
					name
					type
					object {
						... on Blob {
							text
							isBinary
							isTruncated
						}
						... on Tree {
							entries {
								name
								type
								object {
									... on Blob {
										text
										isBinary
										isTruncated
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
`

type RepositoryLatestReleaseAsset struct {
	Name        string
	DownloadUrl string
}

type RepositoryLatestRelease struct {
	Name        string
	Tag         string
	Url         string
	Description string
	Assets      []RepositoryLatestReleaseAsset
}

type RepositoryFile struct {
	Name string

	owner string
	repo  string
	ref   string

	data []byte
}

type Repository struct {
	Description   string
	HomepageUrl   string
	DefaultBranch string
	Head          string
	Forks         int
	Stars         int
	Watchers      int
	LicenseSpdx   string
	LicenseData   *RepositoryFile
	LatestRelease *RepositoryLatestRelease
	Releases      []string
	Readme        *RepositoryFile
	Docs          []*RepositoryFile
	Headers       []*RepositoryFile
}

type repositoryBlob struct {
	Text        *string `json:"text"`
	IsBinary    bool    `json:"isBinary"`
	IsTruncated bool    `json:"isTruncated"`
}

func newRepositoryFile(owner string, repo string, path string, ref string, blob *repositoryBlob) *RepositoryFile {
	rv := &RepositoryFile{
		Name:  path,
		owner: owner,
		repo:  repo,
		ref:   ref,
	}

	if blob.Text != nil && !blob.IsBinary && !blob.IsTruncated {
		rv.data = []byte(*blob.Text)
	}
	return rv
}

func (r *RepositoryFile) Read() ([]byte, error) {
	if r.data != nil {
		return r.data, nil
	}

	v, err := GetRepositoryFile(r.owner, r.repo, r.Name, r.ref)
	if err != nil {
		return nil, err
	}
	defer v.Close()

	rv, err := io.ReadAll(v)
	if err != nil {
		return nil, err
	}
	r.data = rv
	return r.data, nil
}

func GetRepository(owner string, repo string, headersDir *string) (*Repository, error) {
	o := struct {
		Repository struct {
			Description      string `json:"description"`
			HomepageUrl      string `json:"homepageUrl"`
			DefaultBranchRef struct {
				Name   string `json:"name"`
				Target struct {
					Oid string `json:"oid"`
				} `json:"target"`
			} `json:"defaultBranchRef"`
			Head struct {
				Oid string `json:"oid"`
			} `json:"head"`
			ForkCount      int `json:"forkCount"`
			StargazerCount int `json:"stargazerCount"`
			Watchers       struct {
				TotalCount int `json:"totalCount"`
			} `json:"watchers"`
			LicenseInfo *struct {
				SpdxId string `json:"spdxId"`
			} `json:"licenseInfo"`
			License       *repositoryBlob `json:"license"`
			Copying       *repositoryBlob `json:"copying"`
			LatestRelease *struct {
				Name          string `json:"name"`
				TagName       string `json:"tagName"`
				Url           string `json:"url"`
				Description   string `json:"description"`
				ReleaseAssets struct {
					Nodes []struct {
						Name        string `json:"name"`
						DownloadUrl string `json:"downloadUrl"`
					} `json:"nodes"`
					TotalCount int `json:"totalCount"`
				} `json:"releaseAssets"`
			} `json:"latestRelease"`
			Releases struct {
				Nodes []struct {
					TagName      string `json:"tagName"`
					IsDraft      bool   `json:"isDraft"`
					IsPrerelease bool   `json:"isPrerelease"`
				} `json:"nodes"`
				TotalCount int `json:"totalCount"`
			} `json:"releases"`
			Readme *repositoryBlob `json:"readme"`
			Docs   *struct {
				Entries []struct {
					Name   string         `json:"name"`
					Type   string         `json:"type"`
					Object repositoryBlob `json:"object"`
				} `json:"entries"`
			} `json:"docs"`
			Headers *struct {
				Entries []struct {
					Name   string `json:"name"`
					Type   string `json:"type"`
					Object struct {
						repositoryBlob
						Entries []struct {
							Name   string         `json:"name"`
							Type   string         `json:"type"`
							Object repositoryBlob `json:"object"`
						} `json:"entries"`
					} `json:"object"`
				} `json:"entries"`
			} `json:"headers"`
		} `json:"repository"`
	}{}

	variables := map[string]any{
		"owner": owner,
		"repo":  repo,
	}
	if headersDir != nil {
		if *headersDir == "." {
			variables["headersref"] = "HEAD:"
		} else {
			variables["headersref"] = "HEAD:" + filepath.ToSlash(filepath.Clean(*headersDir))
		}
	}

	if err := GraphqlRequest(getRepository, variables, &o); err != nil {
		return nil, err
	}

	if o.Repository.Head.Oid != o.Repository.DefaultBranchRef.Target.Oid {
		return nil, fmt.Errorf("github: repository: %s/%s: HEAD is not %s: %s != %s",
			owner, repo, o.Repository.DefaultBranchRef.Name, o.Repository.Head.Oid,
			o.Repository.DefaultBranchRef.Target.Oid,
		)
	}

	if o.Repository.LatestRelease != nil && o.Repository.LatestRelease.ReleaseAssets.TotalCount > 100 {
		return nil, fmt.Errorf("github: repository: %s/%s: latest release with more than 100 assets: %d",
			owner, repo, o.Repository.LatestRelease.ReleaseAssets.TotalCount,
		)
	}

	if o.Repository.Releases.TotalCount > 100 {
		return nil, fmt.Errorf("github: repository: %s/%s: more than 100 releases: %d",
			owner, repo, o.Repository.Releases.TotalCount,
		)
	}

	rv := &Repository{
		Description:   o.Repository.Description,
		HomepageUrl:   o.Repository.HomepageUrl,
		DefaultBranch: o.Repository.DefaultBranchRef.Name,
		Head:          o.Repository.Head.Oid,
		Forks:         o.Repository.ForkCount,
		Stars:         o.Repository.StargazerCount,
		Watchers:      o.Repository.Watchers.TotalCount,
	}

	if o.Repository.LicenseInfo != nil && o.Repository.LicenseInfo.SpdxId != "NOASSERTION" {
		rv.LicenseSpdx = o.Repository.LicenseInfo.SpdxId
	} else if o.Repository.License != nil {
		rv.LicenseData = newRepositoryFile(owner, repo, "LICENSE", o.Repository.Head.Oid, o.Repository.License)
	} else if o.Repository.Copying != nil {
		rv.LicenseData = newRepositoryFile(owner, repo, "COPYING", o.Repository.Head.Oid, o.Repository.Copying)
	} else {
		return nil, fmt.Errorf("github: repository: %s/%s: failed to find license", owner, repo)
	}

	if o.Repository.LatestRelease != nil {
		rv.LatestRelease = &RepositoryLatestRelease{
			Name:        o.Repository.LatestRelease.Name,
			Tag:         o.Repository.LatestRelease.TagName,
			Url:         o.Repository.LatestRelease.Url,
			Description: o.Repository.LatestRelease.Description,
		}
		for _, asset := range o.Repository.LatestRelease.ReleaseAssets.Nodes {
			rv.LatestRelease.Assets = append(rv.LatestRelease.Assets, RepositoryLatestReleaseAsset(asset))
		}
	}

	for _, release := range o.Repository.Releases.Nodes {
		if !release.IsDraft && !release.IsPrerelease {
			rv.Releases = append(rv.Releases, release.TagName)
		}
	}

	if o.Repository.Readme != nil {
		rv.Readme = newRepositoryFile(owner, repo, "README.md", o.Repository.Head.Oid, o.Repository.Readme)
	}

	if o.Repository.Docs != nil {
		for _, doc := range o.Repository.Docs.Entries {
			if doc.Type != "blob" || (path.Ext(doc.Name) != ".md" && path.Ext(doc.Name) != ".markdown") {
				continue
			}
			rv.Docs = append(rv.Docs, newRepositoryFile(owner, repo, path.Join("docs", doc.Name), o.Repository.Head.Oid, &doc.Object))
		}
	}

	if o.Repository.Headers != nil {
		prefix := ""
		if headersDir != nil {
			prefix = *headersDir
		}
		for _, entry := range o.Repository.Headers.Entries {
			if entry.Type == "tree" {
				for _, subEntry := range entry.Object.Entries {
					if subEntry.Type != "blob" || path.Ext(subEntry.Name) != ".h" {
						continue
					}
					rv.Headers = append(rv.Headers, newRepositoryFile(owner, repo, path.Join(prefix, entry.Name, subEntry.Name), o.Repository.Head.Oid, &subEntry.Object))
				}
				continue
			}

			if entry.Type != "blob" || path.Ext(entry.Name) != ".h" {
				continue
			}
			rv.Headers = append(rv.Headers, newRepositoryFile(owner, repo, path.Join(prefix, entry.Name), o.Repository.Head.Oid, &entry.Object.repositoryBlob))
		}
	}
	return rv, nil
}

func GetRepositoryFile(owner string, repo string, ppath string, ref string) (io.ReadCloser, error) {
	// this semi-documented "raw.githubusercontent.com" service is know for being problematic
	// try it first if we have a ref, since it is "free", but failback to the rest api
	if ref != "" {
		if body, err := Request("GET", "https://raw.githubusercontent.com/"+owner+"/"+repo+"/"+ref+"/"+ppath, nil, nil); err == nil {
			return body, nil
		}
	}

	headers := map[string]string{
		"accept": "application/vnd.github.raw+json",
	}
	return Request("GET", path.Join("repos", owner, repo, "contents", ppath), headers, nil)
}
