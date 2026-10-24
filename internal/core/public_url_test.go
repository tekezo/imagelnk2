package core

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubIPResolver struct {
	addresses []netip.Addr
	err       error
}

func (r stubIPResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return r.addresses, r.err
}

func TestValidatePublicURL(t *testing.T) {
	t.Run("public IP address", func(t *testing.T) {
		err := validatePublicURL(context.Background(), "https://1.1.1.1/image.png", stubIPResolver{})
		require.NoError(t, err)
	})

	privateURLs := []string{
		"http://localhost/",
		"http://service.localhost/",
		"http://printer.local/",
		"http://127.0.0.1/",
		"http://10.0.0.1/",
		"http://172.16.0.1/",
		"http://192.168.0.1/",
		"http://169.254.169.254/latest/meta-data/",
		"http://[::1]/",
		"http://[fc00::1]/",
		"http://[fe80::1]/",
	}
	for _, rawURL := range privateURLs {
		t.Run(rawURL, func(t *testing.T) {
			err := validatePublicURL(context.Background(), rawURL, stubIPResolver{})
			var privateURLError *ErrPrivateURL
			require.ErrorAs(t, err, &privateURLError)
		})
	}

	t.Run("hostname resolving to a public address", func(t *testing.T) {
		resolver := stubIPResolver{addresses: []netip.Addr{netip.MustParseAddr("203.0.113.1")}}
		err := validatePublicURL(context.Background(), "https://example.com/", resolver)
		require.NoError(t, err)
	})

	t.Run("hostname resolving to a private address", func(t *testing.T) {
		resolver := stubIPResolver{addresses: []netip.Addr{netip.MustParseAddr("192.168.1.20")}}
		err := validatePublicURL(context.Background(), "https://internal.example.com/", resolver)

		var privateURLError *ErrPrivateURL
		require.ErrorAs(t, err, &privateURLError)
		assert.Equal(t, "internal.example.com", privateURLError.Host)
	})

	t.Run("mixed public and private addresses", func(t *testing.T) {
		resolver := stubIPResolver{addresses: []netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("10.0.0.1")}}
		err := validatePublicURL(context.Background(), "https://example.com/", resolver)

		var privateURLError *ErrPrivateURL
		require.ErrorAs(t, err, &privateURLError)
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		err := validatePublicURL(context.Background(), "file:///etc/passwd", stubIPResolver{})
		require.Error(t, err)
		assert.False(t, errors.As(err, new(*ErrPrivateURL)))
	})

	t.Run("missing host", func(t *testing.T) {
		err := validatePublicURL(context.Background(), "https:///path", stubIPResolver{})
		require.Error(t, err)
	})

	t.Run("resolver error", func(t *testing.T) {
		resolver := stubIPResolver{err: errors.New("lookup failed")}
		err := validatePublicURL(context.Background(), "https://example.com/", resolver)
		require.ErrorContains(t, err, "failed to resolve")
	})
}

func TestSaveImageCacheRejectsPrivateURL(t *testing.T) {
	result := NewResult()
	result.AppendImageURL("http://169.254.169.254/latest/meta-data/")

	err := result.SaveImageCache(nil, "https://example.com/")

	var privateURLError *ErrPrivateURL
	require.ErrorAs(t, err, &privateURLError)
}
