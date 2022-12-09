package connector

import (
	"fmt"

	"github.com/kameshsampath/harness-cli/pkg/common"
	"github.com/kameshsampath/harness-cli/pkg/types"
	"github.com/kameshsampath/harness-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type DeleteOptions struct {
	Name      string
	ProjectID string
	// account, org, project
	Scope string
}

type DeleteConnector struct {
	APIKey            string
	AccountID         string
	Identifier        string
	OrgID             string
	ProjectIdentifier string
	Scope             string
}

// Run implements RESTCall
func (dc *DeleteConnector) Call() (map[string]interface{}, error) {
	req := utils.NewHTTPRequest(dc.APIKey, dc.AccountID)
	utils.AddScopedIDQueryParams(req, dc.Scope, dc.OrgID, dc.ProjectIdentifier)
	return utils.DeleteResourceByID(req, "https://app.harness.io/gateway/ng/api/connectors/{id}", dc.Identifier)
}

// Print implements common.Command
func (dc *DeleteConnector) Print(rm map[string]interface{}, err error) {
	if v, ok := rm["status"]; ok && v == "SUCCESS" {
		log.Tracef("%#v", rm)
		if rm["data"].(bool) {
			fmt.Printf(`Connector "%s" deleted successfully`, dc.Identifier)
		}
	} else {
		log.Errorf("%#v", rm)
	}
}

// AddFlags implements Command
func (do *DeleteOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&do.Name, "name", "n", "", "The name of the docker registry connector to delete.")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&do.ProjectID, "project-id", "p", "", `The project where the secret will be deleted.`)
	cmd.Flags().StringVarP(&do.Scope, "connector-scope", "", "project", `The secret scope. Valid value is one of "project", "org", "account"`)
}

// Execute implements Command
func (do *DeleteOptions) Execute(cmd *cobra.Command, args []string) error {
	dc := &DeleteConnector{
		APIKey:     viper.GetString("api-key"),
		AccountID:  viper.GetString("account-id"),
		Identifier: utils.IDFromName(do.Name),
		Scope:      do.Scope,
	}

	if do.Scope == "project" {
		dc.OrgID = viper.GetString("org-id")
		dc.ProjectIdentifier = do.ProjectID
	} else if do.Scope == "org" {
		dc.OrgID = viper.GetString("org-id")
	}

	dc.Print(dc.Call())

	return nil
}

// Validate implements Command
func (do *DeleteOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())
	return nil
}

// (TODO:kamesh) add more examples
var delCommandExample = fmt.Sprintf(`
  # Delete connector default options
  %[1]s delete --name foo --account-id <your account id> --project-id <project id>
  # Delete connector specific project id
  %[1]s delete --name foo --account-id <your account id> --project-id <project id>  --org-id=<orgid>
  # Delete connector at account level
  %[1]s delete --name foo --account-id <your account id>  --secret-scope="account"
  # Create delete at org level
  %[1]s delete --name foo --account-id <your account id>  --secret-scope="org"
`, common.ExamplePrefix())

// NewDeleteDockerConnectorCommand instantiates the new instance of the NewDeleteDockerConnectorCommand
func NewDeleteConnectorCommand() *cobra.Command {
	do := &DeleteOptions{}

	drCmd := &cobra.Command{
		Use:     "delete",
		Short:   "Deletes a Connector.",
		Example: delCommandExample,
		RunE:    do.Execute,
		PreRunE: do.Validate,
	}

	do.AddFlags(drCmd)

	return drCmd
}

var _ types.Command = (*DeleteOptions)(nil)
var _ types.RESTCall = (*DeleteConnector)(nil)
