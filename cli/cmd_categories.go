package cli

import "github.com/spf13/cobra"

func (a *App) categoriesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "categories",
		Short: "List all categories on Tieng Trung Online",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cats, err := a.client.Categories(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			if a.limit > 0 && len(cats) > a.limit {
				cats = cats[:a.limit]
			}
			return a.renderOrEmpty(cats, len(cats))
		},
	}
}
