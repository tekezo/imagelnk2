package core

import "golang.org/x/exp/slices"

type (
	Result struct {
		Title          string   `json:"title"`
		ImageURLs      []string `json:"imageURLs"`
		ExtraURLs      []string `json:"extraURLs"`
		ImageCacheURLs []string `json:"imageCacheURLs"`
	}
)

func NewResult() *Result {
	return &Result{
		Title:          "",
		ImageURLs:      []string{},
		ExtraURLs:      []string{},
		ImageCacheURLs: []string{},
	}
}

func (r *Result) AppendImageURL(url string) {
	if !slices.Contains(r.ImageURLs, url) {
		r.ImageURLs = append(r.ImageURLs, url)
	}
}

func (r *Result) AppendExtraURL(url string) {
	if !slices.Contains(r.ExtraURLs, url) {
		r.ExtraURLs = append(r.ExtraURLs, url)
	}
}

func (r *Result) AppendImageCacheURL(url string) {
	if !slices.Contains(r.ImageCacheURLs, url) {
		r.ImageCacheURLs = append(r.ImageCacheURLs, url)
	}
}
