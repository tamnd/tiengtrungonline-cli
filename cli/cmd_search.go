package cli

import "github.com/spf13/cobra"

func (a *App) searchCmd() *cobra.Command {
	var perPage int
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search posts by title",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := a.client.Search(cmd.Context(), args[0], perPage)
			if err != nil {
				return mapFetchErr(err)
			}
			if a.limit > 0 && len(results) > a.limit {
				results = results[:a.limit]
			}
			return a.renderOrEmpty(results, len(results))
		},
	}
	cmd.Flags().IntVar(&perPage, "per-page", 20, "number of results to return")
	return cmd
}
