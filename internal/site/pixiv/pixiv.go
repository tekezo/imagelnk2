package pixiv

import (
	"imagelnk2/internal/core"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Pixiv struct {
	browser *rod.Browser
}

func New(browser *rod.Browser) Pixiv {
	return Pixiv{
		browser: browser,
	}
}

func (p Pixiv) GetCanonicalURL(url string) string {
	if strings.HasPrefix(url, "https://www.pixiv.net/") {
		return url
	}
	return ""
}

func (p Pixiv) OpenPage(url string) (*rod.Page, string, []byte, error) {
	return core.OpenPage(p.browser, url, core.OpenPageOptions{
		Cookies: []*proto.NetworkCookieParam{
			{
				Name:   "PHPSESSID",
				Value:  core.Config.Pixiv.PHPSessionID,
				Domain: ".pixiv.net",
			},
			{
				Name:   "device_token",
				Value:  core.Config.Pixiv.DeviceToken,
				Domain: ".pixiv.net",
			},
		},
		SupportRawImage: false,
	})
}

func (p Pixiv) GetImageURLs(page *rod.Page, canonicalURL string) (*core.Result, error) {
	result := core.NewResult()

	result.Title = core.GetTitle(page)

	if links, err := page.Elements(`div[role="presentation"] a`); err == nil {
		for _, link := range links {
			href := core.GetAttribute(link, "href")
			if href != "" {
				result.AppendImageURL(href)
			}
		}
	}

	return result, nil
}
