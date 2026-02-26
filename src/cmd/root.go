package cmd

import (
	"fmt"
	"os"

	"deployer/src/config"
	"deployer/src/server"

	"github.com/spf13/cobra"
)

func Execute() {
	var configFile string

	rootCmd := &cobra.Command{
		Use:     "deployer",
		Short:   "secure CI/CD deployment server",
		Version: config.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Run(configFile)
		},
	}

	rootCmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "configuration filename")
	rootCmd.SilenceErrors = true

	rootCmd.AddCommand(newClientCmd())

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
