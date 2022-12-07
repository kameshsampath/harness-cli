/*
 * Copyright (c) 2022.
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *             http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and limitations under the License.
 */

package commands

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

type Command interface {

	//AddFlags adds the flags to the commands
	AddFlags(cmd *cobra.Command)

	// Validate checks the commands line arguments and configuration
	// prior to executing the commands.
	Validate(cmd *cobra.Command, args []string) error

	// Execute executes the commands
	Execute(cmd *cobra.Command, args []string) error
}

type RESTCall interface {
	Call() (*resty.Response, error)
}
