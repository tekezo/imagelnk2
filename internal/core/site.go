package core

import "github.com/go-rod/rod"

type Site interface {
	GetCanonicalURL(url string) string
	OpenPage(url string) (page *rod.Page, mime string, body []byte, err error)
	WaitReady(page *rod.Page, canonicalURL string) error
	Extract(page *rod.Page, canonicalURL string) (*Result, error)
}
