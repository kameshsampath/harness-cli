package project

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
	Name string
}

type DeleteProject struct {
	APIKey     string
	AccountID  string
	Identifier string
	OrgID      string
}

// Run implements RESTCall
func (dp *DeleteProject) Call() (*resty.Response, error) {
	var resp *resty.Response

	client := resty.New()
	req := client.R().
		EnableTrace().
		SetHeader("x-api-key", dp.APIKey).
		SetQueryParam("accountIdentifier", dp.AccountID).
		SetQueryParam("orgIdentifier", dp.OrgID)

	req.
		SetPathParams(map[string]string{
			"id": dp.Identifier,
		})

	log.Tracef("%#v", req)

	resp, err := req.
		Delete("https://app.harness.io/gateway/ng/api/projects/{id}")
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
}

// Execute implements Command
func (do *DeleteOptions) Execute(cmd *cobra.Command, args []string) error {
	ds := &DeleteProject{
		APIKey:     viper.GetString("api-key"),
		AccountID:  viper.GetString("account-id"),
		OrgID:      viper.GetString("org-id"),
		Identifier: utils.IDFromName(do.Name),
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
			fmt.Printf("Project %s deleted successfully", do.Name)
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
var deleteProjectCommandExample = fmt.Sprintf(`
  # Delete project 
  %[1]s project  delete --name foo --account-id <your account id>
  # Delete project 
  %[1]s project delete --name foo --account-id <your account id> --org-id=<org id>
`, common.ExamplePrefix())

// NewDeleteProjectCommand instantiates the new instance of the NewDeleteProjectCommand
func NewDeleteProjectCommand() *cobra.Command {
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

var _ common.Command = (*DeleteOptions)(nil)
var _ common.RESTCall = (*DeleteProject)(nil)
