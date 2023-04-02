package main

import (
	"github.com/spf13/cobra"
	"github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/errs"
)

var authorizeCmd = &cobra.Command{
	Use: "authorize",
}

var generateCodeAuthorizeCmd = &cobra.Command{
	Use: "gen",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errs.NewError(errs.InvalidArgument, "Must provide authorization config file path and output directory").
				ToError()
		}

		config, err := authorization.ParseConfig(args[0], dataCollector)
		if err != nil {
			return err
		}

		internalErr := authorization.GenerateCode(config, args[1])
		if internalErr != nil {
			return internalErr.ToError()
		}

		return nil
	},
}

func addAuthorizationCmd() {
	authorizeCmd.AddCommand(generateCodeAuthorizeCmd)
	rootCmd.AddCommand(authorizeCmd)
}
