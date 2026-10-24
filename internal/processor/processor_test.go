package processor

import (
	"context"
	"imagelnk2/internal/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedirectedURL(t *testing.T) {
	t.Run("x.com", func(t *testing.T) {
		p := New(nil)
		assert.Equal(t, "https://x.com/example?123", p.redirectedURL("https://twitter.com/example?123"))
		assert.Equal(t, "https://example.com?url=https://x.com/example", p.redirectedURL("https://example.com?url=https://x.com/example"))
	})
}

func TestPrivateURLIsRejected(t *testing.T) {
	p := New(nil)

	t.Run("GetImageURLs", func(t *testing.T) {
		_, err := p.GetImageURLs(context.Background(), "http://127.0.0.1/private")

		var privateURLError *core.ErrPrivateURL
		assert.ErrorAs(t, err, &privateURLError)
	})

	t.Run("SaveHTML", func(t *testing.T) {
		err := p.SaveHTML("http://192.168.0.1/private", "private.html")

		var privateURLError *core.ErrPrivateURL
		assert.ErrorAs(t, err, &privateURLError)
	})

	t.Run("Debug", func(t *testing.T) {
		_, err := p.Debug("http://[::1]/private", "private.html")

		var privateURLError *core.ErrPrivateURL
		assert.ErrorAs(t, err, &privateURLError)
	})
}
