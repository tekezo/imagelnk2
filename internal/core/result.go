package core

import (
	"log"
	"net/http"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"golang.org/x/exp/slices"
)

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

func (r *Result) SaveImageCache(page *rod.Page, canonicalURL string) {
	// When a request is made for an image, the request may be recursive with different types (Document and Image) for the same URL.
	// In such cases, only the first result is used, so remember the URLs that have already been added.
	savedImageURLs := []string{}

	page = page.CancelTimeout().Timeout(30 * time.Second)

	router := page.HijackRequests()
	defer router.MustStop()
	go router.Run()

	router.MustAdd("*", func(ctx *rod.Hijack) {
		url := ctx.Request.URL().String()

		if slices.Contains(r.ImageURLs, url) {
			log.Printf("hijack %s %s", ctx.Request.Type(), ctx.Request.URL())

			if err := ctx.LoadResponse(http.DefaultClient, true); err == nil {
				if !slices.Contains(savedImageURLs, url) {
					savedImageURLs = append(savedImageURLs, url)

					url, err := SaveImageCache(
						[]byte(ctx.Response.Body()),
					)
					if err != nil {
						log.Print(err)
					} else {
						r.AppendImageCacheURL(url)
					}
				}
			}
		}

		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	})

	cleanup := page.MustSetExtraHeaders("Referer", canonicalURL)
	defer cleanup()

	for _, imageURL := range r.ImageURLs {
		log.Printf("save image cache: %s", imageURL)

		waitNavigation := page.Timeout(5 * time.Second).MustWaitNavigation()
		page.Timeout(5 * time.Second).Navigate(imageURL)
		waitNavigation()

		if len(r.ImageCacheURLs) > 4 {
			break
		}
	}
}
