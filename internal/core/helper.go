package core

import (
	"log"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/spf13/viper"
)

var (
	spaceRegexp       = regexp.MustCompile(`\s+`)
	contentTypeRegexp = regexp.MustCompile(`^([^;]+)`)
	elementTimeout    = 100 * time.Millisecond
	pageReadyTimeout  = 5 * time.Second
)

func ReadConfig() error {
	viper.SetConfigName("imagelnk2")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.BindEnv("hostname")
	if err != nil {
		return err
	}

	err = viper.BindEnv("port")
	if err != nil {
		return err
	}

	viper.Unmarshal(&Config)

	return nil
}

type OpenPageOptions struct {
	Cookies         []*proto.NetworkCookieParam
	SupportRawImage bool
}

func OpenPage(
	browser *rod.Browser,
	url string,
	options OpenPageOptions,
) (page *rod.Page, mime string, body []byte, err error) {
	page = stealth.MustPage(browser)
	router := page.HijackRequests()
	defer router.MustStop()

	var blockedErr error
	var blockedErrMutex sync.Mutex
	router.MustAdd("*", func(ctx *rod.Hijack) {
		if ctx.Request.Type() == proto.NetworkResourceTypeDocument {
			requestURL := ctx.Request.URL()
			if requestURL.Scheme == "http" || requestURL.Scheme == "https" {
				if validationErr := ValidatePublicURL(ctx.Request.Req().Context(), requestURL.String()); validationErr != nil {
					blockedErrMutex.Lock()
					blockedErr = validationErr
					blockedErrMutex.Unlock()
					ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
					return
				}
			}
		}

		ctx.Skip = true
	})

	if options.SupportRawImage {
		router.MustAdd(url, func(ctx *rod.Hijack) {
			_ = ctx.LoadResponse(newPublicHTTPClient(), true)

			if contentType, ok := ctx.Response.Headers()["Content-Type"]; ok {
				if len(contentType) > 0 {
					m := contentTypeRegexp.FindStringSubmatch(contentType[0])
					if len(m) > 0 {
						mime = m[0]
					}
				}
			}

			body = []byte(ctx.Response.Body())

			ctx.Response.SetBody(body)
		})
	}

	router.MustAdd("*", func(ctx *rod.Hijack) {
		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})
	go router.Run()

	if options.Cookies != nil {
		page.SetCookies(options.Cookies)
	}

	navigationErr := page.Navigate(url)

	blockedErrMutex.Lock()
	err = blockedErr
	blockedErrMutex.Unlock()
	if err != nil {
		return
	}
	if navigationErr != nil {
		err = navigationErr
		return
	}

	return
}

func WaitForPageStable(page *rod.Page) error {
	return page.Timeout(pageReadyTimeout).WaitStable(800 * time.Millisecond)
}

func WaitForElement(page *rod.Page, selector string, name string) (*rod.Element, error) {
	element, err := page.Timeout(pageReadyTimeout).Element(selector)
	if err != nil {
		return nil, NewErrMandatoryElementNotFound(name)
	}
	return element, nil
}

func StartRequestReduceRouter(page *rod.Page) *rod.HijackRouter {
	router := page.HijackRequests()

	router.MustAdd("*", func(ctx *rod.Hijack) {
		if slices.Contains([]proto.NetworkResourceType{
			proto.NetworkResourceTypeFont,
			proto.NetworkResourceTypeImage,
			proto.NetworkResourceTypeMedia,
			proto.NetworkResourceTypePing,
			proto.NetworkResourceTypeStylesheet,
		}, ctx.Request.Type()) {
			ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			return
		}

		log.Printf(
			"request: %s %s",
			ctx.Request.Type(),
			ctx.Request.URL(),
		)

		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})

	return router
}

func GetAttribute(element *rod.Element, name string) string {
	value, err := element.Attribute(name)
	if err != nil {
		return ""
	}

	if value == nil {
		return ""
	}

	return *value
}

func FindElementInPage(page *rod.Page, selector string) *rod.Element {
	e, err := page.Timeout(elementTimeout).Element(selector)
	if err != nil {
		return nil
	}
	return e
}

func FindElement(element *rod.Element, selector string) *rod.Element {
	e, err := element.Timeout(elementTimeout).Element(selector)
	if err != nil {
		return nil
	}
	return e
}

func GetTitle(page *rod.Page) string {
	title := FindElementInPage(page, "title")
	if title != nil {
		return FormatTitle(title.MustText())
	}

	return ""
}

func FormatTitle(title string) string {
	return spaceRegexp.ReplaceAllString(strings.TrimSpace(title), " ")
}

func GetOpenGraphTitle(page *rod.Page) string {
	meta := FindElementInPage(page, `meta[property="og:title"]`)
	if meta != nil {
		return GetAttribute(meta, "content")
	}

	return ""
}

func GetOpenGraphImage(page *rod.Page) string {
	meta := FindElementInPage(page, `meta[property="og:image"]`)
	if meta != nil {
		return GetAttribute(meta, "content")
	}

	return ""
}

func GetOpenGraphURL(page *rod.Page) string {
	meta := FindElementInPage(page, `meta[property="og:url"]`)
	if meta != nil {
		return GetAttribute(meta, "content")
	}

	return ""
}
