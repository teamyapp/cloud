package main

import (
	"github.com/spf13/cobra"
	"github.com/teamyapp/cloud/libs/authorization"
)

var authorizeCmd = &cobra.Command{
	Use: "authorize",
}

var generateCodeAuthorizeCmd = &cobra.Command{
	Use: "gen",
	RunE: func(cmd *cobra.Command, args []string) error {
		authorizationSrcFile := cliConfig.AuthorizationCoreSrcFile
		if len(args) >= 1 {
			authorizationSrcFile = args[0]
		}

		config, err := authorization.ParseConfig(authorizationSrcFile, dataCollector)
		if err != nil {
			return err
		}

		authorizationOutputFile := cliConfig.AuthorizationCoreOutputDir
		if len(args) >= 2 {
			authorizationOutputFile = args[1]
		}

		internalErr := authorization.GenerateCode(config, authorizationOutputFile)
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
