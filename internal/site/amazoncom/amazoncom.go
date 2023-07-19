package amazoncom

import (
	"imagelnk2/internal/core"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Amazon struct {
	browser *rod.Browser
}

func New(browser *rod.Browser) Amazon {
	return Amazon{
		browser: browser,
	}
}

func (a Amazon) GetCanonicalURL(url string) string {
	if strings.HasPrefix(url, "https://www.amazon.com/") {
		return url
	}
	return ""
}

func (a Amazon) OpenPage(url string) (*rod.Page, string, []byte, error) {
	return core.OpenPage(a.browser, url, core.OpenPageOptions{
		Cookies: []*proto.NetworkCookieParam{
			{
				Name:   "session-token",
				Value:  core.Config.Amazoncom.SessionToken,
				Domain: ".amazon.com",
			},
			{
				Name:   "session-id",
				Value:  core.Config.Amazoncom.SessionID,
				Domain: ".amazon.com",
			},
		},
		SupportRawImage: false,
	})
}

func (a Amazon) GetImageURLs(page *rod.Page, canonicalURL string) (*core.Result, error) {
	result := core.NewResult()

	result.Title = core.GetTitle(page)
	if result.Title == "" {
		return nil, core.NewErrMandatoryElementNotFound("title")
	}

	img := core.FindElementInPage(page, `img[data-old-hires]`)
	if img != nil {
		src := core.GetAttribute(img, "data-old-hires")
		if src != "" {
			result.AppendImageURL(src)
		}
	}

	return result, nil
}
