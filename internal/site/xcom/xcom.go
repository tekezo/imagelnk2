package xcom

import (
	"fmt"
	"imagelnk2/internal/core"
	"regexp"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

var (
	urlRegexp               = regexp.MustCompile(`^(https://x.com/[^/]+/status/[^/]+)`)
	timeURLRegexp           = regexp.MustCompile(`^https://[^/]+(/[^/]+/status/\d+)`)
	imageURLNameQueryRegexp = regexp.MustCompile(`&name=[^&]+$`)
)

type Xcom struct {
	browser *rod.Browser
}

func New(browser *rod.Browser) Xcom {
	return Xcom{
		browser: browser,
	}
}

func (x Xcom) GetCanonicalURL(url string) string {
	m := urlRegexp.FindStringSubmatch(url)
	if m == nil {
		return ""
	}
	return m[1]
}

func (x Xcom) OpenPage(url string) (*rod.Page, string, []byte, error) {
	return core.OpenPage(x.browser, url, core.OpenPageOptions{
		Cookies: []*proto.NetworkCookieParam{
			{
				Name:   "auth_token",
				Value:  core.Config.Xcom.AuthToken,
				Domain: "x.com",
			},
		},
		SupportRawImage: false,
		WaitNavigation:  true,
		WaitRequestIdle: true,
	})
}

func (x Xcom) GetImageURLs(page *rod.Page, canonicalURL string) (*core.Result, error) {
	result := core.NewResult()

	//
	// Find tweet element by link
	//

	// timeURL == "/ImageLnk/status/1670350649484267520"
	m := timeURLRegexp.FindStringSubmatch(canonicalURL)
	if m == nil {
		return nil, fmt.Errorf("unsupported URL %s", canonicalURL)
	}
	timeURL := m[1]

	// Edited tweet have history element instead of time element.
	// historyURL == "/C2_STAFF/status/1679790222937296898/history"
	historyURL := timeURL + "/history"

	var tweet *rod.Element
	if elements, err := page.Elements(`*[data-testid="tweet"]`); err == nil {
		for _, e := range elements {
			timeElement := core.FindElement(e, fmt.Sprintf(`a[href="%s"]`, timeURL))
			if timeElement != nil {
				tweet = e
				break
			}

			historyElement := core.FindElement(e, fmt.Sprintf(`a[href="%s"]`, historyURL))
			if historyElement != nil {
				tweet = e
				break
			}
		}
	}

	if tweet == nil {
		// If tweet is not found, it's age-restricted tweet
		result.Title = "error: tweet element is not found"
		return nil, core.NewErrMandatoryElementNotFound(fmt.Sprintf("tweet timeURL:%s", timeURL))
	}

	//
	// Get Title
	//

	userName := core.FindElement(tweet, `*[data-testid="User-Name"] a:first-of-type`)
	if userName != nil {
		result.Title = userName.MustText() + ": "
	}

	if tweetTexts, err := tweet.Elements(`*[data-testid="tweetText"]`); err == nil {
		if len(tweetTexts) > 0 {
			tweetText := tweetTexts[0]

			if elements, err := tweetText.Elements(`& > *`); err == nil {
				for _, element := range elements {
					if tagName, err := element.Property("tagName"); err == nil {
						if tagName.String() == "A" {
							href := core.GetAttribute(element, "href")
							if href != "" {
								result.Title += " " + href + " "
								result.AppendExtraURL(href)
							}
						} else if tagName.String() == "IMG" {
							// emoji
							alt := core.GetAttribute(element, "alt")
							if alt != "" {
								result.Title += alt
							}
						} else {
							result.Title += element.MustText()
						}
					}
				}
			}
		}
	}

	//
	// Get ImageURLs
	//

	// The quote tweet might have multiple tweetPhoto blocks.
	// We use the first block.

	if tweetPhotos, err := tweet.Elements(`*[data-testid="tweetPhoto"]`); err == nil {
		if len(tweetPhotos) > 0 {
			tweetPhotoParent := tweetPhotos[0]. // <div>
								MustParent(). // <div>
								MustParent(). // <a>
								MustParent(). // <div>
								MustParent(). // <div>
								MustParent(). // <div>
								MustParent()  // <div>

			if tweetPhotoParent != nil {
				if tweetPhotos, err := tweetPhotoParent.Elements(`*[data-testid="tweetPhoto"] img`); err == nil {
					for _, tweetPhoto := range tweetPhotos {
						src := core.GetAttribute(tweetPhoto, "src")
						if src != "" {
							// src == "https://pbs.twimg.com/media/Ef1ej1WUYAATrzu?format=jpg&name=900x900"
							// Remove name=XXX query to avoid image shrinkage.
							imageURL := imageURLNameQueryRegexp.ReplaceAllString(src, "")

							result.AppendImageURL(imageURL)
						}
					}
				}
			}
		}
	}

	if videos, err := tweet.Elements(`video`); err == nil {
		for _, video := range videos {
			poster := core.GetAttribute(video, "poster")
			if poster != "" {
				result.AppendImageURL(poster)
			}
		}
	}

	if cardWrappers, err := tweet.Elements(`*[data-testid="card.wrapper"] a:first-of-type`); err == nil {
		for _, cardWrapper := range cardWrappers {
			href := core.GetAttribute(cardWrapper, "href")
			if href != "" {
				result.AppendExtraURL(href)
			}
		}
	}

	return result, nil
}
