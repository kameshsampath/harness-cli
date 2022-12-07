package secret

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

type DeleteSecret struct {
	APIKey            string
	AccountID         string
	Identifier        string
	OrgID             string
	ProjectIdentifier string
	Scope             string
}

// Run implements RESTCall
func (ds *DeleteSecret) Call() (*resty.Response, error) {
	var resp *resty.Response

	client := resty.New()
	req := client.R().
		EnableTrace().
		SetHeader("x-api-key", ds.APIKey).
		SetQueryParam("accountIdentifier", ds.AccountID)

	if ds.Scope == "project" {
		req.SetQueryParam("orgIdentifier", ds.OrgID)
		req.SetQueryParam("projectIdentifier", ds.ProjectIdentifier)
	} else if ds.Scope == "org" {
		req.SetQueryParam("orgIdentifier", ds.OrgID)
	}

	req.
		SetPathParams(map[string]string{
			"id": ds.Identifier,
		})

	log.Tracef("%#v", req)

	resp, err := req.
		Delete("https://app.harness.io/gateway/ng/api/v2/secrets/{id}")
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
	cmd.Flags().StringVarP(&do.Scope, "secret-scope", "", "project", `The secret scope. Valid value is one of "project", "org", "account"`)
}

// Execute implements Command
func (do *DeleteOptions) Execute(cmd *cobra.Command, args []string) error {
	ds := &DeleteSecret{
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
			fmt.Printf("Secret %s deleted successfully", do.Name)
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
var deleteSecretCommandExample = fmt.Sprintf(`
  # Delete secret from file with default options
  %[1]s delete-secret --name foo --account-id <your account id> --project-id <project id>
  # Delete secret from file with  specific organization id
  %[1]s delete-secret --name foo --account-id <your account id> --project-id <project id>  --org-id=<orgid>
  # Create new secret from text 
  %[1]s new-secret --name foo --account-id <your account id> --project-id <project id> --text foo --type=SecretText
  # Create new secret from text at account scope, default is project
  %[1]s delete-secret --name foo --account-id <your account id>  --secret-scope="account"
  # Create delete secret at org scope, default is project
  %[1]s delete-secret --name foo --account-id <your account id>  --secret-scope="org"
`, common.ExamplePrefix())

// NewDeleteSecretCommand instantiates the new instance of the DeleteSecretCommand
func NewDeleteSecretCommand() *cobra.Command {
	do := &DeleteOptions{}

	sfCmd := &cobra.Command{
		Use:     "delete-secret",
		Short:   "Delete a secret.",
		Example: deleteSecretCommandExample,
		RunE:    do.Execute,
		PreRunE: do.Validate,
	}

	do.AddFlags(sfCmd)

	return sfCmd
}

var _ common.Command = (*DeleteOptions)(nil)
var _ common.RESTCall = (*DeleteSecret)(nil)
