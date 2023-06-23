package core

import "github.com/go-rod/rod"

type Site interface {
	GetCanonicalURL(url string) string
	OpenPage(url string) (page *rod.Page, mime string, body []byte, err error)
	GetImageURLs(page *rod.Page, canonicalURL string) (*Result, error)
}
