package bluesky

import (
	"context"
	"fmt"
	"imagelnk2/internal/core"
	"regexp"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
)

var (
	urlRegexp = regexp.MustCompile(`^https://bsky.app/profile/([^/]+)/post/([^/\?]+)`)
)

type Bluesky struct {
}

func New() Bluesky {
	return Bluesky{}
}

func (b Bluesky) GetCanonicalURL(url string) string {
	m := urlRegexp.FindStringSubmatch(url)
	if m == nil {
		return ""
	}
	return fmt.Sprintf("https://bsky.app/profile/%s/post/%s", m[1], m[2])
}

func (b Bluesky) GetImageURLs(ctx context.Context, canonicalURL string) (*core.Result, error) {
	result := core.NewResult() // timeURL == "/ImageLnk/status/1670350649484267520"

	m := urlRegexp.FindStringSubmatch(canonicalURL)
	if m == nil {
		return nil, fmt.Errorf("unsupported URL %s", canonicalURL)
	}
	actor := m[1]
	rkey := m[2]

	client := &xrpc.Client{
		Host: "https://bsky.social",
	}

	sessionInput := &atproto.ServerCreateSession_Input{
		Identifier: core.Config.Bluesky.Identifier,
		Password:   core.Config.Bluesky.Password,
	}

	session, err := atproto.ServerCreateSession(ctx, client, sessionInput)
	if err != nil {
		return nil, fmt.Errorf("failed to log in: %v", err)
	}

	client.Auth = &xrpc.AuthInfo{
		AccessJwt:  session.AccessJwt,
		RefreshJwt: session.RefreshJwt,
		Did:        session.Did,
		Handle:     session.Handle,
	}

	profile, err := bsky.ActorGetProfile(ctx, client, actor)
	if err != nil {
		return nil, fmt.Errorf("failed to get actor profile: %v", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	uris := []string{
		fmt.Sprintf("at://%s/%s/%s", profile.Did, "app.bsky.feed.post", rkey),
	}
	output, err := bsky.FeedGetPosts(ctx, client, uris)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %v", err)
	}
	if output == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	for _, post := range output.Posts {
		var author string
		var text string

		if post.Author != nil {
			if post.Author.DisplayName != nil {
				author = *post.Author.DisplayName
			}

			if author == "" {
				author = post.Author.Handle
			}
		}

		if post.Record != nil {
			text = post.Record.Val.(*bsky.FeedPost).Text
		}

		if post.Embed != nil {
			if post.Embed.EmbedImages_View != nil {
				for _, image := range post.Embed.EmbedImages_View.Images {
					result.AppendImageURL(image.Fullsize)
				}
			}

			if post.Embed.EmbedExternal_View != nil {
				text = fmt.Sprintf("%s: %s %s", text,
					post.Embed.EmbedExternal_View.External.Title,
					post.Embed.EmbedExternal_View.External.Uri)

				if post.Embed.EmbedExternal_View.External.Thumb != nil {
					result.AppendImageURL(*post.Embed.EmbedExternal_View.External.Thumb)
				}
			}
		}

		result.Title = core.FormatTitle(fmt.Sprintf("%s: %s", author, text))
	}

	return result, nil
}
