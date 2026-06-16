package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// categoryAliases maps friendly names to category slugs.
var categoryAliases = map[string]string{
	"grammar": "ngu-phap-tieng-trung",
	"vocab":   "tu-vung-tieng-trung",
	"hsk":     "tieng-trung-hsk-hskk",
	"lesson":  "duong-dai",
}

func resolveCategory(alias string) string {
	if slug, ok := categoryAliases[alias]; ok {
		return slug
	}
	return alias
}

func (a *App) listCmd() *cobra.Command {
	var (
		categoryAlias string
		perPage       int
		page          int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List posts, optionally filtered by category",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			catID := 0
			if categoryAlias != "" {
				slug := resolveCategory(categoryAlias)
				id, err := a.client.CategoryID(ctx, slug)
				if err != nil {
					return codeError(exitUsage, fmt.Errorf("unknown category: %s", categoryAlias))
				}
				catID = id
			}
			posts, err := a.client.Posts(ctx, perPage, page, catID)
			if err != nil {
				return mapFetchErr(err)
			}
			if a.limit > 0 && len(posts) > a.limit {
				posts = posts[:a.limit]
			}
			return a.renderOrEmpty(posts, len(posts))
		},
	}
	cmd.Flags().StringVar(&categoryAlias, "category", "", "category: grammar|vocab|hsk|lesson|<slug>")
	cmd.Flags().IntVar(&perPage, "per-page", 20, "number of posts per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}
