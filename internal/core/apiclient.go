package core

import "context"

type APIClient interface {
	GetCanonicalURL(url string) string
	GetImageURLs(ctx context.Context, canonicalURL string) (*Result, error)
}
