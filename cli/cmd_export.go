package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *App) exportCmd() *cobra.Command {
	var (
		categoryAlias string
		all           bool
	)
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export posts as JSONL",
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

			perPage := 100
			var collected int
			for page := 1; ; page++ {
				r, err := a.client.PostsWithTotal(ctx, perPage, page, catID)
				if err != nil {
					return mapFetchErr(err)
				}
				if len(r.Posts) == 0 {
					break
				}
				posts := r.Posts
				if a.limit > 0 {
					remaining := a.limit - collected
					if len(posts) > remaining {
						posts = posts[:remaining]
					}
				}
				if err := a.render(posts); err != nil {
					return err
				}
				collected += len(posts)
				if a.limit > 0 && collected >= a.limit {
					break
				}
				if !all || len(r.Posts) < perPage {
					break
				}
			}
			if collected == 0 {
				return codeError(exitNoData, nil)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&categoryAlias, "category", "", "category: grammar|vocab|hsk|lesson|<slug>")
	cmd.Flags().BoolVar(&all, "all", false, "paginate through all posts (may be slow)")
	return cmd
}
