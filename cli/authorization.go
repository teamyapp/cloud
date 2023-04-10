package main

import (
	"errors"
	"github.com/spf13/cobra"
	"github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/errs"
	"os"
)

var authorizeCmd = &cobra.Command{
	Use: "authorize",
}

var generateCodeAuthorizeCmd = &cobra.Command{
	Use: "gen",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, authorizationOption := range cliConfig.AuthorizationOptions {
			authorizationSrcFile := authorizationOption.ConfigFilePath
			if len(args) >= 1 {
				authorizationSrcFile = args[0]
			}

			config, err := authorization.ParseConfig(authorizationSrcFile)
			if err != nil {
				return err.ToError()
			}

			authorizationOutputDir := authorizationOption.OutputDir
			if len(args) >= 2 {
				authorizationOutputDir = args[1]
			}

			if _, err := os.Stat(authorizationOutputDir); errors.Is(err, os.ErrNotExist) {
				err := os.Mkdir(authorizationOutputDir, os.ModePerm)
				if err != nil {
					logger.Error(errs.NewError(errs.Unknown, err.Error()))
				}
			}

			internalErr := authorization.GenerateCode(config, logger, authorizationOutputDir)
			if internalErr != nil {
				logger.Error(internalErr)
			}
		}

		return nil
	},
}

func addAuthorizationCmd() {
	authorizeCmd.AddCommand(generateCodeAuthorizeCmd)
	rootCmd.AddCommand(authorizeCmd)
}
