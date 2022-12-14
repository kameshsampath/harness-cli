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

package secret

import (
	"github.com/spf13/cobra"
)

func NewSecretCommands() *cobra.Command {
	secretCmd := &cobra.Command{
		Use:              "secret",
		Short:            "Group of commands to manipulate the connectors.",
		TraverseChildren: true,
	}

	//Commands

	secretCmd.AddCommand(newSecretCommand())
	secretCmd.AddCommand(newDeleteSecretCommand())

	return secretCmd
}
