package core

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

var (
	spaceRegexp       = regexp.MustCompile(`\s+`)
	contentTypeRegexp = regexp.MustCompile(`^([^;]+)`)
	elementTimeout    = 100 * time.Millisecond
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
	page = browser.MustPage()

	if options.SupportRawImage {
		router := page.HijackRequests()
		defer router.MustStop()
		go router.Run()

		router.MustAdd(url, func(ctx *rod.Hijack) {
			_ = ctx.LoadResponse(http.DefaultClient, true)

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

	if options.Cookies != nil {
		page.SetCookies(options.Cookies)
	}

	//waitNavigation := page.Timeout(5 * time.Second).MustWaitNavigation()
	//waitRequestIdle := page.Timeout(5 * time.Second).MustWaitRequestIdle()

	page.MustNavigate(url)

	err = rod.Try(func() {
		//waitNavigation()
		//waitRequestIdle()

		page.MustWaitStable()
	})
	if err != nil {
		return
	}

	return
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
		return spaceRegexp.ReplaceAllString(strings.TrimSpace(title.MustText()), " ")
	}

	return ""
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
