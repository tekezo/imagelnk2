package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const publicURLLookupTimeout = 5 * time.Second

type ipResolver interface {
	LookupNetIP(ctx context.Context, network string, host string) ([]netip.Addr, error)
}

// ErrPrivateURL indicates that a URL points to a local or otherwise non-public
// network address.
type ErrPrivateURL struct {
	Host string
}

func (e *ErrPrivateURL) Error() string {
	return fmt.Sprintf("private URL is not allowed: %s", e.Host)
}

// ValidatePublicURL validates that rawURL is an HTTP(S) URL whose host resolves
// only to public IP addresses.
func ValidatePublicURL(ctx context.Context, rawURL string) error {
	return validatePublicURL(ctx, rawURL, net.DefaultResolver)
}

func newPublicHTTPClient() *http.Client {
	client := *http.DefaultClient
	client.CheckRedirect = func(request *http.Request, _ []*http.Request) error {
		return ValidatePublicURL(request.Context(), request.URL.String())
	}
	return &client
}

func validatePublicURL(ctx context.Context, rawURL string, resolver ipResolver) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid fetch URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported fetch URL scheme: %q", parsedURL.Scheme)
	}

	host := strings.TrimSuffix(strings.ToLower(parsedURL.Hostname()), ".")
	if host == "" {
		return fmt.Errorf("fetch URL has no host")
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
		return &ErrPrivateURL{Host: host}
	}

	if address, err := netip.ParseAddr(host); err == nil {
		if isPrivateAddress(address) {
			return &ErrPrivateURL{Host: host}
		}
		return nil
	}

	lookupContext, cancel := context.WithTimeout(ctx, publicURLLookupTimeout)
	defer cancel()

	addresses, err := resolver.LookupNetIP(lookupContext, "ip", host)
	if err != nil {
		return fmt.Errorf("failed to resolve fetch URL host %q: %w", host, err)
	}
	if len(addresses) == 0 {
		return fmt.Errorf("fetch URL host %q resolved to no addresses", host)
	}

	for _, address := range addresses {
		if isPrivateAddress(address) {
			return &ErrPrivateURL{Host: host}
		}
	}

	return nil
}

func isPrivateAddress(address netip.Addr) bool {
	address = address.Unmap()

	return !address.IsValid() ||
		address.IsPrivate() ||
		address.IsLoopback() ||
		address.IsLinkLocalUnicast() ||
		address.IsLinkLocalMulticast() ||
		address.IsMulticast() ||
		address.IsUnspecified()
}
