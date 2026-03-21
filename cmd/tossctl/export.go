package main

import (
	"github.com/junghoonkye/tossinvest-cli/internal/output"
	"github.com/spf13/cobra"
)

func newExportCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export read-only data to machine-friendly formats (CSV/JSON)",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "positions",
			Short: "Export current portfolio positions",
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := newAppContext(opts)
				if err != nil {
					return err
				}
				positions, err := app.client.ListPositions(cmd.Context())
				if err != nil {
					return userFacingCommandError(err)
				}
				return output.WritePositions(cmd.OutOrStdout(), output.FormatCSV, positions)
			},
		},
		&cobra.Command{
			Use:   "orders",
			Short: "Export completed order history",
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := newAppContext(opts)
				if err != nil {
					return err
				}
				orders, err := app.client.ListCompletedOrders(cmd.Context(), "all")
				if err != nil {
					return userFacingCommandError(err)
				}
				return output.WriteCompletedOrders(cmd.OutOrStdout(), output.FormatCSV, orders)
			},
		},
	)

	return cmd
}
