package github

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/kameshsampath/harness-cli/pkg/common"
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
func (dc *DeleteConnector) Call() (*resty.Response, error) {
	var resp *resty.Response

	client := resty.New()
	req := client.R().
		EnableTrace().
		SetHeader("x-api-key", dc.APIKey).
		SetQueryParam("accountIdentifier", dc.AccountID)

	if dc.Scope == "project" {
		req.SetQueryParam("orgIdentifier", dc.OrgID)
		req.SetQueryParam("projectIdentifier", dc.ProjectIdentifier)
	} else if dc.Scope == "org" {
		req.SetQueryParam("orgIdentifier", dc.OrgID)
	}

	req.
		SetPathParams(map[string]string{
			"id": dc.Identifier,
		})

	log.Tracef("%#v", req)

	resp, err := req.
		Delete("https://app.harness.io/gateway/ng/api/connectors/{id}")
	if err != nil {
		return nil, err
	}
	var rm map[string]interface{}
	err = json.Unmarshal(resp.Body(), &rm)
	if err != nil {
		return nil, err
	}

	log.Tracef("%#v", rm)

	return resp, err
}

// AddFlags implements Command
func (do *DeleteOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&do.Name, "name", "n", "", "The name of the secret to delete.")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&do.ProjectID, "project-id", "p", "", `The project where the secret will be deleted.`)
	cmd.Flags().StringVarP(&do.Scope, "connector-scope", "", "project", `The secret scope. Valid value is one of "project", "org", "account"`)
}

// Execute implements Command
func (do *DeleteOptions) Execute(cmd *cobra.Command, args []string) error {
	ds := &DeleteConnector{
		APIKey:     viper.GetString("api-key"),
		AccountID:  viper.GetString("account-id"),
		Identifier: utils.IDFromName(do.Name),
		Scope:      do.Scope,
	}

	if do.Scope == "project" {
		ds.OrgID = viper.GetString("org-id")
		ds.ProjectIdentifier = do.ProjectID
	} else if do.Scope == "org" {
		ds.OrgID = viper.GetString("org-id")
	}

	resp, err := ds.Call()
	if err != nil {
		return err
	}
	var rm map[string]interface{}
	err = json.Unmarshal(resp.Body(), &rm)
	if err != nil {
		return err
	}
	if v, ok := rm["status"]; ok && v == "SUCCESS" {
		log.Tracef("%#v", rm)
		if rm["data"].(bool) {
			fmt.Printf("GitHub Connector %s deleted successfully", do.Name)
		}
	} else {
		log.Errorf("%#v", rm)
	}
	return nil
}

// Validate implements Command
func (do *DeleteOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())
	return nil
}

// (TODO:kamesh) add more examples
var deleteGHConnectorCommandExample = fmt.Sprintf(`
  # Delete connector default options
  %[1]s github delete --name foo --account-id <your account id> --project-id <project id>
  # Delete connector specific project id
  %[1]s github delete --name foo --account-id <your account id> --project-id <project id>  --org-id=<orgid>
  # Delete connector at account level
  %[1]s github delete --name foo --account-id <your account id>  --secret-scope="account"
  # Create delete at org level
  %[1]s github delete --name foo --account-id <your account id>  --secret-scope="org"
`, common.ExamplePrefix())

// NewDeleteGitHubConnectorCommand instantiates the new instance of the NewDeleteGitHubConnectorCommand
func NewDeleteGitHubConnectorCommand() *cobra.Command {
	do := &DeleteOptions{}

	ghcCmd := &cobra.Command{
		Use:     "delete",
		Short:   "Deletes a Github Connector.",
		Example: deleteGHConnectorCommandExample,
		RunE:    do.Execute,
		PreRunE: do.Validate,
	}

	do.AddFlags(ghcCmd)

	return ghcCmd
}

var _ common.Command = (*DeleteOptions)(nil)
var _ common.RESTCall = (*DeleteConnector)(nil)
