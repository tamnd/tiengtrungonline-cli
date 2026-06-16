package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tamnd/tiengtrungonline-cli/tiengtrungonline"
)

func (a *App) parseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "parse <url>",
		Short: "Parse any tiengtrungonline.com URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			rawURL := args[0]
			u, err := url.Parse(rawURL)
			if err != nil {
				return codeError(exitUsage, fmt.Errorf("invalid URL: %v", err))
			}
			host := u.Host
			if host != "tiengtrungonline.com" && host != "www.tiengtrungonline.com" {
				return codeError(exitUsage, fmt.Errorf("not a tiengtrungonline.com URL: %s", rawURL))
			}
			seg := strings.Trim(u.Path, "/")
			// Looks like a post URL (ends in .html or is a single path segment)
			if strings.HasSuffix(seg, ".html") || (!strings.Contains(seg, "/") && seg != "") {
				slug := strings.TrimSuffix(seg, ".html")
				lesson, err := a.client.Lesson(ctx, slug)
				if err != nil {
					return mapFetchErr(err)
				}
				return a.renderOrEmpty([]*tiengtrungonline.Lesson{lesson}, 1)
			}
			// Category-like URL
			if seg != "" {
				id, err := a.client.CategoryID(ctx, seg)
				if err != nil {
					// Try as post slug
					lesson, err2 := a.client.Lesson(ctx, seg)
					if err2 != nil {
						return mapFetchErr(err)
					}
					return a.renderOrEmpty([]*tiengtrungonline.Lesson{lesson}, 1)
				}
				posts, err := a.client.Posts(ctx, 20, 1, id)
				if err != nil {
					return mapFetchErr(err)
				}
				if a.limit > 0 && len(posts) > a.limit {
					posts = posts[:a.limit]
				}
				return a.renderOrEmpty(posts, len(posts))
			}
			return codeError(exitUsage, fmt.Errorf("cannot parse URL: %s", rawURL))
		},
	}
}
