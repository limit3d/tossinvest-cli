package main

import "github.com/spf13/cobra"

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export read-only data to machine-friendly formats",
	}

	cmd.AddCommand(
		newStubCommand("positions", "Export positions data"),
		newStubCommand("orders", "Export orders data"),
	)

	return cmd
}
