package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/tiengtrungonline-cli/tiengtrungonline"
)

func (a *App) infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print site statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := a.client.SiteInfo(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty([]*tiengtrungonline.SiteInfo{info}, 1)
		},
	}
}
