/*
 * Copyright © 2022  Kamesh Sampath <kamesh.sampath@hotmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package commands

import (
	"io"
	"os"
	"strings"

	"github.com/kameshsampath/harness-cli/pkg/connector"
	"github.com/kameshsampath/harness-cli/pkg/delegate"
	"github.com/kameshsampath/harness-cli/pkg/project"
	"github.com/kameshsampath/harness-cli/pkg/secret"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var v, apiKey, accountID, orgID string

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "harness-cli",
		Short: "A simple tool to interact with Harness API https://apidocs.harness.io.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := logSetup(os.Stdout, v); err != nil {
				return err
			}
			return nil
		},
		TraverseChildren: true,
	}

	// what should the env prefix while loading the environment variables
	viper.SetEnvPrefix("harness")
	// since environment variable by convention are _, replace "-" with "_"
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	// load flags from environment variables
	viper.AutomaticEnv()
	pf := rootCmd.PersistentFlags()

	pf.StringVarP(&v, "verbose", "v", log.InfoLevel.String(), "The logging level to set")
	pf.StringVarP(&apiKey, "api-key", "k", "", "The Harness API Key.")
	rootCmd.MarkFlagRequired("api-key")
	pf.StringVarP(&accountID, "account-id", "a", "", "The harness account id.")
	rootCmd.MarkFlagRequired("account-id")
	pf.StringVarP(&orgID, "org-id", "o", "default", "The organization id to use.")

	//Commands
	rootCmd.AddCommand(NewVersionCommand())
	rootCmd.AddCommand(project.NewProjectCommands())
	rootCmd.AddCommand(secret.NewSecretCommands())
	rootCmd.AddCommand(connector.NewConnectorsCommands())
	rootCmd.AddCommand(delegate.NewDelegateCommands())

	return rootCmd
}

func logSetup(out io.Writer, level string) error {
	log.SetOutput(out)
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}
