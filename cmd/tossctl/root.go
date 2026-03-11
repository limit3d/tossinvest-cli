package main

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/junghoonkye/toss-investment-cli/internal/auth"
	tossclient "github.com/junghoonkye/toss-investment-cli/internal/client"
	"github.com/junghoonkye/toss-investment-cli/internal/config"
	"github.com/junghoonkye/toss-investment-cli/internal/output"
	"github.com/junghoonkye/toss-investment-cli/internal/session"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	outputFormat string
	configDir    string
	sessionFile  string
}

type appContext struct {
	format      output.Format
	paths       config.Paths
	authService *auth.Service
	client      *tossclient.Client
}

func newRootCmd() *cobra.Command {
	opts := &rootOptions{}

	cmd := &cobra.Command{
		Use:   "tossctl",
		Short: "Read-only CLI for Toss Securities web data",
		Long: "tossctl is a Go-first CLI scaffold for a future read-only Toss Securities " +
			"client with browser-assisted login.",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			_, err := output.ParseFormat(opts.outputFormat)
			return err
		},
	}

	cmd.PersistentFlags().StringVar(
		&opts.outputFormat,
		"output",
		string(output.FormatTable),
		"Output format: table, json, csv",
	)
	cmd.PersistentFlags().StringVar(
		&opts.configDir,
		"config-dir",
		"",
		"Override the config directory",
	)
	cmd.PersistentFlags().StringVar(
		&opts.sessionFile,
		"session-file",
		"",
		"Override the session file path",
	)

	cmd.AddCommand(
		newAuthCmd(opts),
		newAccountCmd(opts),
		newPortfolioCmd(opts),
		newOrdersCmd(opts),
		newWatchlistCmd(opts),
		newQuoteCmd(opts),
		newExportCmd(),
	)

	return cmd
}

func newAppContext(opts *rootOptions) (*appContext, error) {
	format, err := output.ParseFormat(opts.outputFormat)
	if err != nil {
		return nil, err
	}

	paths, err := config.DefaultPaths()
	if err != nil {
		return nil, err
	}

	if opts.configDir != "" {
		paths.ConfigDir = opts.configDir
		paths.SessionFile = filepath.Join(opts.configDir, "session.json")
	}

	if opts.sessionFile != "" {
		paths.SessionFile = opts.sessionFile
	}

	store := session.NewFileStore(paths.SessionFile)
	sess, err := store.Load(context.Background())
	if err != nil && !errors.Is(err, session.ErrNoSession) {
		return nil, err
	}

	loginConfig := auth.DefaultLoginConfig(paths.CacheDir)
	client := tossclient.New(tossclient.Config{Session: sess})

	return &appContext{
		format: format,
		paths:  paths,
		authService: auth.NewService(store, paths.SessionFile, auth.Options{
			LoginConfig: loginConfig,
			Validator:   client,
		}),
		client: client,
	}, nil
}
