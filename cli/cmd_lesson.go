package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/tiengtrungonline-cli/tiengtrungonline"
)

func (a *App) lessonCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lesson <slug>",
		Short: "Fetch a single lesson by slug",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			lesson, err := a.client.Lesson(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty([]*tiengtrungonline.Lesson{lesson}, 1)
		},
	}
}
