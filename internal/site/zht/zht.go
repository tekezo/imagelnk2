package zht

import (
	"fmt"
	"imagelnk2/internal/core"
	"regexp"

	"github.com/go-rod/rod"
)

var (
	urlRegexp = regexp.MustCompile(`^http://risky-safety.org/~zinnia/d/\d{4}/\d{2}/(#.+)`)
)

type ZHT struct {
	browser *rod.Browser
}

func New(browser *rod.Browser) ZHT {
	return ZHT{
		browser: browser,
	}
}

func (z ZHT) GetCanonicalURL(url string) string {
	m := urlRegexp.FindStringSubmatch(url)
	if m == nil {
		return ""
	}
	return url
}

func (z ZHT) OpenPage(url string) (*rod.Page, string, []byte, error) {
	return core.OpenPage(z.browser, url, core.OpenPageOptions{
		Cookies:         nil,
		SupportRawImage: false,
	})
}

func (z ZHT) GetImageURLs(page *rod.Page, canonicalURL string) (result *core.Result, err error) {
	result = core.NewResult()

	m := urlRegexp.FindStringSubmatch(canonicalURL)
	if m == nil {
		return
	}

	link := core.FindElementInPage(page, fmt.Sprintf(`a[href="%s"]`, m[1]))
	if link != nil {
		parent := link.MustParent()
		link.MustRemove()
		result.Title = parent.MustText()
	}

	if result.Title == "" {
		result.Title = core.GetTitle(page)
	}

	return
}
