package generic

import (
	"imagelnk2/internal/core"

	"github.com/go-rod/rod"
)

type Generic struct {
	browser *rod.Browser
}

func New(browser *rod.Browser) Generic {
	return Generic{
		browser: browser,
	}
}

func (g Generic) GetCanonicalURL(url string) string {
	return url
}

func (g Generic) OpenPage(url string) (*rod.Page, string, []byte, error) {
	return core.OpenPage(g.browser, url, core.OpenPageOptions{
		Cookies:         nil,
		SupportRawImage: true,
	})
}

func (g Generic) GetImageURLs(page *rod.Page, canonicalURL string) (*core.Result, error) {
	result := core.NewResult()

	result.Title = core.GetOpenGraphTitle(page)
	if result.Title == "" {
		result.Title = core.GetTitle(page)
	}

	imageURL := core.GetOpenGraphImage(page)
	if imageURL != "" {
		result.AppendImageURL(imageURL)
	}

	return result, nil
}
