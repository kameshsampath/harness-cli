package secret

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

type DeleteSecret struct {
	APIKey            string
	AccountID         string
	Identifier        string
	OrgID             string
	ProjectIdentifier string
	Scope             string
}

// Run implements RESTCall
func (ds *DeleteSecret) Call() (map[string]interface{}, error) {
	req := utils.NewHTTPRequest(ds.APIKey, ds.AccountID)
	utils.AddScopedIDQueryParams(req, ds.Scope, ds.OrgID, ds.ProjectIdentifier)
	return utils.DeleteResourceByID(req, "https://app.harness.io/gateway/ng/api/v2/secrets/{id}", ds.Identifier)
}

// Print implements Command
func (ds *DeleteSecret) Print(rm map[string]interface{}, err error) {
	if v, ok := rm["status"]; ok && v == "SUCCESS" {
		log.Tracef("%#v", rm)
		if rm["data"].(bool) {
			fmt.Printf(`Secret "%s" deleted successfully`, ds.Identifier)
		}
	} else {
		log.Errorf("%#v", rm)
	}
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

	ds.Print(ds.Call())

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
  %[1]s secret delete --name foo --account-id <your account id> --project-id <project id>
  # Delete secret from file with  specific organization id
  %[1]s secret delete --name foo --account-id <your account id> --project-id <project id>  --org-id=<orgid>
  # Create new secret from text 
  %[1]s delete --name foo --account-id <your account id> --project-id <project id> --text foo --type=SecretText
  # Create new secret from text at account scope, default is project
  %[1]s secret delete --name foo --account-id <your account id>  --secret-scope="account"
  # Create delete secret at org scope, default is project
  %[1]s secret delete --name foo --account-id <your account id>  --secret-scope="org"
`, common.ExamplePrefix())

// newDeleteSecretCommand instantiates the new instance of the newDeleteSecretCommand
func newDeleteSecretCommand() *cobra.Command {
	do := &DeleteOptions{}

	sfCmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a secret.",
		Example: deleteSecretCommandExample,
		RunE:    do.Execute,
		PreRunE: do.Validate,
	}

	do.AddFlags(sfCmd)

	return sfCmd
}

var _ types.Command = (*DeleteOptions)(nil)
var _ types.RESTCall = (*DeleteSecret)(nil)
