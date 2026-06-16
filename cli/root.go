// Package cli builds the tto command tree.
package cli

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/tamnd/tiengtrungonline-cli/tiengtrungonline"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

const (
	exitError  = 1
	exitUsage  = 2
	exitNoData = 3
)

// ExitError carries an exit code and an optional wrapped error.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit %d", e.Code)
}

func (e *ExitError) Unwrap() error { return e.Err }

func codeError(code int, err error) error { return &ExitError{Code: code, Err: err} }

// App holds state shared across all cobra commands.
type App struct {
	client   *tiengtrungonline.Client
	cfg      tiengtrungonline.Config
	output   string
	fields   []string
	noHeader bool
	template string
	limit    int
}

// Root assembles the root cobra command with all subcommands.
func Root() *cobra.Command {
	app := &App{cfg: tiengtrungonline.DefaultConfig()}

	root := &cobra.Command{
		Use:           "tto",
		Short:         "Browse Tieng Trung Online from the command line",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return app.setup()
		},
	}

	pf := root.PersistentFlags()
	pf.StringVarP(&app.output, "output", "o", "auto", "output: table|json|jsonl|csv|tsv|url|raw")
	pf.StringSliceVar(&app.fields, "fields", nil, "comma-separated columns to include")
	pf.BoolVar(&app.noHeader, "no-header", false, "omit header row in table/csv/tsv")
	pf.StringVar(&app.template, "template", "", "Go text/template per record")
	pf.IntVarP(&app.limit, "limit", "n", 0, "limit number of records (0 = all)")
	pf.DurationVar(&app.cfg.Rate, "delay", app.cfg.Rate, "minimum spacing between requests")
	pf.DurationVar(&app.cfg.Timeout, "timeout", app.cfg.Timeout, "per-request timeout")
	pf.IntVar(&app.cfg.Retries, "retries", app.cfg.Retries, "retry attempts on 429/5xx")

	root.AddCommand(
		app.postsCmd(),
		app.categoriesCmd(),
		app.listCmd(),
		app.searchCmd(),
		app.wordCmd(),
		app.lessonCmd(),
		app.parseCmd(),
		app.exportCmd(),
		app.infoCmd(),
		newVersionCmd(),
	)
	return root
}

func (a *App) setup() error {
	if a.output == "" || a.output == "auto" {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			a.output = string(FormatTable)
		} else {
			a.output = string(FormatJSONL)
		}
	}
	if !Format(a.output).Valid() {
		return codeError(exitUsage, fmt.Errorf("unknown output format %q", a.output))
	}
	a.client = tiengtrungonline.NewClient(a.cfg)
	return nil
}

func (a *App) render(records any) error {
	r := NewRenderer(os.Stdout, Format(a.output), a.fields, a.noHeader, a.template)
	return r.Render(records)
}

func (a *App) renderOrEmpty(records any, n int) error {
	if err := a.render(records); err != nil {
		return err
	}
	if n == 0 {
		return codeError(exitNoData, nil)
	}
	return nil
}

func mapFetchErr(err error) error {
	if err == nil {
		return nil
	}
	return codeError(exitError, err)
}
