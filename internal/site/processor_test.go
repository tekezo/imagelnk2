package site

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedirectedURL(t *testing.T) {
	t.Run("x.com", func(t *testing.T) {
		p := NewProcessor(nil)
		assert.Equal(t, "https://x.com/example?123", p.redirectedURL("https://twitter.com/example?123"))
		assert.Equal(t, "https://example.com?url=https://x.com/example", p.redirectedURL("https://example.com?url=https://x.com/example"))
	})
}
