// Command tiengtrungonline is the tto single-binary command line for tiengtrungonline.com.
package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/fang"
	"github.com/tamnd/tiengtrungonline-cli/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := cli.Root()
	if err := fang.Execute(ctx, root,
		fang.WithVersion(cli.Version),
		fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM),
	); err != nil {
		var xe *cli.ExitError
		if errors.As(err, &xe) {
			os.Exit(xe.Code)
		}
		os.Exit(1)
	}
}
