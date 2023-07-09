package main

import (
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use: "client",
}

var applyAuthorizationConfigAuthZCloudClientCmd = &cobra.Command{
	Use: "cloud:authZ:applyAuthorizationConfig",
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}

func addClientCmd() {
	clientCmd.AddCommand(applyAuthorizationConfigAuthZCloudClientCmd)
	rootCmd.AddCommand(clientCmd)
}
