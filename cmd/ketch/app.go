package main

import (
	"io"

	"github.com/spf13/cobra"
)

func newAppCmd(cfg config, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage applications",
		Long:  `Manage applications`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	cmd.AddCommand(newAppCreateCmd(cfg, out))
	cmd.AddCommand(newAppDeployCmd(cfg, out))
	cmd.AddCommand(newAppUpdateCmd(cfg, out))
	cmd.AddCommand(newAppListCmd(cfg, out))
	cmd.AddCommand(newAppRemoveCmd(cfg, out, appRemove))
	cmd.AddCommand(newAppInfoCmd(cfg, out))
	cmd.AddCommand(newAppStartCmd(cfg, out, appStart))
	cmd.AddCommand(newAppStopCmd(cfg, out, appStop))
	cmd.AddCommand(newAppExportCmd(cfg, out, appExport))
	return cmd
}
