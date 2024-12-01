package processor

import (
	"context"
	"fmt"
	"imagelnk2/internal/apiclients/bluesky"
	"imagelnk2/internal/core"
	"imagelnk2/internal/site/amazoncojp"
	"imagelnk2/internal/site/amazoncom"
	"imagelnk2/internal/site/generic"
	"imagelnk2/internal/site/pixiv"
	"imagelnk2/internal/site/xcom"
	"imagelnk2/internal/site/zht"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/yosssi/gohtml"
)

var (
	twitterURLRegexp = regexp.MustCompile(`^https://twitter\.com/`)
)

type Processor struct {
	browser    *rod.Browser
	apiClients []core.APIClient
	sites      []core.Site
}

func New(browser *rod.Browser) Processor {
	return Processor{
		browser: browser,
		apiClients: []core.APIClient{
			bluesky.New(),
		},
		sites: []core.Site{
			amazoncojp.New(browser),
			amazoncom.New(browser),
			pixiv.New(browser),
			xcom.New(browser),
			zht.New(browser),
			generic.New(browser),
		},
	}
}

func (p Processor) GetImageURLs(ctx context.Context, url string) (*core.Result, error) {
	url = p.redirectedURL(url)

	apiClient, canonicalURL := p.findAPIClient(url)
	if apiClient != nil {
		result, err := (*apiClient).GetImageURLs(ctx, canonicalURL)
		if err != nil {
			return nil, err
		}

		page := rod.New().MustConnect().MustPage()

		result.SaveImageCache(page, canonicalURL)
		return result, nil
	}

	site, canonicalURL := p.findSite(url)
	if site != nil {
		page, mime, body, err := (*site).OpenPage(canonicalURL)
		if err != nil {
			return nil, err
		}

		log.Printf("page is opened: %s %s", canonicalURL, mime)

		if strings.HasPrefix(mime, "image/") {
			result := core.NewResult()

			url, err := core.SaveImageCache(body)
			if err != nil {
				log.Print(err)
			} else {
				result.AppendImageCacheURL(url)
			}

			return result, nil
		}

		// Set timeout after page is prepared to avoid deadline exceeded error in OpenPage.
		page = page.Timeout(5 * time.Second)

		result, err := (*site).GetImageURLs(page, canonicalURL)
		if err != nil {
			return nil, err
		}

		result.SaveImageCache(page, canonicalURL)
		return result, nil
	}

	return nil, fmt.Errorf("no matching apiClient and site found")
}

func (p Processor) SaveHTML(url string, filename string) error {
	url = p.redirectedURL(url)

	site, canonicalURL := p.findSite(url)
	if site != nil {
		log.Printf("open: %s", canonicalURL)

		page, mime, _, err := (*site).OpenPage(canonicalURL)
		if err != nil {
			return err
		}
		defer page.MustClose()

		log.Printf("page is opened: %s %s", canonicalURL, mime)

		// Set timeout after page is prepared to avoid deadline exceeded error in OpenPage.
		page = page.Timeout(5 * time.Second)

		removeElementSelectors := []string{
			"iframe",
			"link",
			"noscript",
			"script",
			"style",
		}
		for _, selector := range removeElementSelectors {
			if elements, err := page.Elements(selector); err == nil {
				for _, e := range elements {
					err := e.Remove()
					if err != nil {
						return err
					}
				}
			}
		}

		html := page.MustHTML()
		html = gohtml.Format(html)
		outputPath := filepath.Join("testdata", "html", filename)
		log.Printf("save %s", outputPath)

		err = os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
		if err != nil {
			return err
		}

		err = os.WriteFile(outputPath, []byte(html), 0600)
		if err != nil {
			return err
		}

		// Check whether mandatory elements exist
		_, err = p.Debug(canonicalURL, filename)
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("no site is matched")
}

func (p Processor) Debug(url string, path string) (*core.Result, error) {
	url = p.redirectedURL(url)

	apiClient, _ := p.findAPIClient(url)
	if apiClient != nil {
		result, err := (*apiClient).GetImageURLs(context.Background(), url)
		if err != nil {
			return nil, err
		}

		return result, nil
	}

	site, _ := p.findSite(url)
	if site != nil {
		testdataURL := fmt.Sprintf("http://%s:%d/testdata?filename=%s", core.Config.Hostname, core.Config.Port, path)

		page := p.browser.MustPage()

		router := core.StartRequestReduceRouter(page)
		defer router.MustStop()
		go router.Run()

		page.MustNavigate(testdataURL)
		page.MustWaitStable()

		// Set timeout after page is prepared to avoid deadline exceeded error.
		page = page.Timeout(5 * time.Second)

		return (*site).GetImageURLs(page, url)
	}

	return nil, fmt.Errorf("no matching apiClient and site found")
}

func (p Processor) redirectedURL(url string) string {
	url = twitterURLRegexp.ReplaceAllString(url, "https://x.com/")
	return url
}

func (p Processor) findAPIClient(url string) (apiClinet *core.APIClient, canonicalURL string) {
	for _, c := range p.apiClients {
		canonicalURL = c.GetCanonicalURL(url)
		if canonicalURL != "" {
			apiClinet = &c
			return
		}
	}

	return
}

func (p Processor) findSite(url string) (site *core.Site, canonicalURL string) {
	for _, s := range p.sites {
		canonicalURL = s.GetCanonicalURL(url)
		if canonicalURL != "" {
			site = &s
			return
		}
	}

	return
}
