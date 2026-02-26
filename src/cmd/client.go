package cmd

import (
	"fmt"
	"os"
	"strings"

	"deployer/src/client"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newClientCmd() *cobra.Command {
	var (
		serverURL string
		component string
		key       string
		argFlags  []string
	)

	cmd := &cobra.Command{
		Use:   "client",
		Short: "send a deployment request to a deployer server",
		RunE: func(cmd *cobra.Command, args []string) error {
			extraArgs := make(map[string]string)

			for _, arg := range argFlags {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid arg format %q: expected key=value", arg)
				}

				extraArgs[parts[0]] = parts[1]
			}

			opts := client.Options{
				URL:       serverURL,
				Component: component,
				Key:       key,
				Args:      extraArgs,
			}

			exitCode, err := client.Run(cmd.Context(), log.StandardLogger(), opts)
			if err != nil {
				return err
			}

			os.Exit(exitCode)

			return nil
		},
	}

	cmd.Flags().StringVarP(&serverURL, "url", "u", "", "deployer server URL")
	cmd.Flags().StringVarP(&component, "component", "c", "", "component name")
	cmd.Flags().StringVarP(&key, "key", "k", "", "security key")
	cmd.Flags().StringArrayVarP(&argFlags, "arg", "a", nil, "extra args as key=value (repeatable)")

	for _, flagName := range []string{"url", "component"} {
		err := cmd.MarkFlagRequired(flagName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	return cmd
}
