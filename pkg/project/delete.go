package project

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
	Name string
}

type DeleteProject struct {
	APIKey     string
	AccountID  string
	Identifier string
	OrgID      string
}

// Run implements types.RESTCall
func (dp *DeleteProject) Call() (map[string]interface{}, error) {
	req := utils.NewHTTPRequest(dp.APIKey, dp.AccountID)
	utils.AddScopedIDQueryParams(req, "", dp.OrgID, "")
	return utils.DeleteResourceByID(req, "https://app.harness.io/gateway/ng/api/projects/{id}", dp.Identifier)
}

// Validate implements types.Command
func (dp *DeleteProject) Print(rm map[string]interface{}, err error) {
	if v, ok := rm["status"]; ok && v == "SUCCESS" {
		log.Tracef("%#v", rm)
		if rm["data"].(bool) {
			fmt.Printf(`Project "%s" deleted successfully`, dp.Identifier)
		}
	} else {
		log.Errorf("%#v", rm)
	}
}

// Validate implements types.Command
func (do *DeleteOptions) Print(rm map[string]interface{}, err error) {
	if v, ok := rm["status"]; ok && v == "SUCCESS" {
		log.Tracef("%#v", rm)
		if rm["data"].(bool) {
			fmt.Printf(`Project "%s" deleted successfully`, do.Name)
		}
	} else {
		log.Errorf("%#v", rm)
	}
}

// AddFlags implements types.Command
func (do *DeleteOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&do.Name, "name", "n", "", "The name of the secret to delete.")
	cmd.MarkFlagRequired("name")
}

// Execute implements types.Command
func (do *DeleteOptions) Execute(cmd *cobra.Command, args []string) error {
	ds := &DeleteProject{
		APIKey:     viper.GetString("api-key"),
		AccountID:  viper.GetString("account-id"),
		OrgID:      viper.GetString("org-id"),
		Identifier: utils.IDFromName(do.Name),
	}

	ds.Print(ds.Call())

	return nil
}

// Validate implements types.Command
func (do *DeleteOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())
	return nil
}

// (TODO:kamesh) add more examples
var deleteProjectCommandExample = fmt.Sprintf(`
  # Delete project 
  %[1]s project  delete --name foo --account-id <your account id>
  # Delete project 
  %[1]s project delete --name foo --account-id <your account id> --org-id=<org id>
`, common.ExamplePrefix())

// newDeleteProjectCommand instantiates the new instance of the newDeleteProjectCommand
func newDeleteProjectCommand() *cobra.Command {
	do := &DeleteOptions{}

	sfCmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a project.",
		Example: deleteProjectCommandExample,
		RunE:    do.Execute,
		PreRunE: do.Validate,
	}

	do.AddFlags(sfCmd)

	return sfCmd
}

var _ types.Command = (*DeleteOptions)(nil)
var _ types.RESTCall = (*DeleteProject)(nil)
