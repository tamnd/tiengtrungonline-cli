package cli

import "github.com/spf13/cobra"

func (a *App) postsCmd() *cobra.Command {
	var (
		perPage    int
		page       int
		categoryID int
	)
	cmd := &cobra.Command{
		Use:   "posts",
		Short: "List posts from Tieng Trung Online",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			posts, err := a.client.Posts(cmd.Context(), perPage, page, categoryID)
			if err != nil {
				return mapFetchErr(err)
			}
			if a.limit > 0 && len(posts) > a.limit {
				posts = posts[:a.limit]
			}
			return a.renderOrEmpty(posts, len(posts))
		},
	}
	cmd.Flags().IntVar(&perPage, "per-page", 20, "number of posts per page")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().IntVar(&categoryID, "category", 0, "filter by category ID (0 = all)")
	return cmd
}
