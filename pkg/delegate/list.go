package delegate

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kameshsampath/harness-cli/pkg/common"
	"github.com/kameshsampath/harness-cli/pkg/types"
	"github.com/kameshsampath/harness-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ListOptions struct {
	// The project identifier used to identify the project
	ProjectID string
	// The scope of the resource "account", "org" or "project"
	Scope string
	// The tags that will be used to filter the delegate
	Tags []string
}

type List struct {
	// APIKey holds the API Key for the API calls
	APIKey string `json:"-"`
	// AccountID holds the AccountID that will be used for API calls
	AccountID string `json:"-"`
	// ProjectIdentifier the project identifier will attach the resource the project
	ProjectIdentifier string `json:"-"`
	// OrgID the organisation ID under which the project "ProjectIdentifier" resides
	OrgID string `json:"-"`
	// The scope of the resource "account", "org" or "project"
	Scope string `json:"-"`
	// The tags that will be used to filter the delegate
	Tags []string `json:"tags"`
}

// AddFlags implements types.Command
func (lo *ListOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&lo.Scope, "delegate-scope", "", "project", `The scope of the delegate. Valid value is one of "project", "org", "account"`)
	cmd.Flags().StringSliceVarP(&lo.Tags, "tags", "t", []string{}, "The tags that will be used to filter the delegate")
}

// Call implements common.RESTCall
func (l *List) Call() (map[string]interface{}, error) {
	req := utils.NewHTTPRequest(l.APIKey, l.AccountID)
	utils.AddScopedIDQueryParams(req, l.Scope, l.OrgID, l.ProjectIdentifier)

	log.Infof("Getting list of delegates for tags %v ", l.Tags)

	return utils.PostJSON(req, "https://app.harness.io/gateway/ng/api/delegate-group-tags/delegate-groups", l)
}

// Execute implements types.Command
func (lo *ListOptions) Execute(cmd *cobra.Command, args []string) error {
	l := &List{
		APIKey:    viper.GetString("api-key"),
		AccountID: viper.GetString("account-id"),
		Tags:      lo.Tags,
		Scope:     lo.Scope,
	}

	if lo.Scope == "project" {
		l.OrgID = viper.GetString("org-id")
		l.ProjectIdentifier = lo.ProjectID
	} else if lo.Scope == "org" {
		l.OrgID = viper.GetString("org-id")
	}

	l.Print(l.Call())

	return nil
}

// Print implements types.Command
func (l *List) Print(rm map[string]interface{}, err error) {
	if err != nil {
		log.Errorf("%s", err)
	}

	log.Tracef("%#v", rm)
	if v, ok := rm["status"]; ok && v == "ERROR" {
		fmt.Printf(rm["message"].(string))
	}

	var resMap []map[string]string
	if v, ok := rm["resource"]; ok {
		resources := v.([]interface{})
		for _, r := range resources {
			res := r.(map[string]interface{})
			id := res["identifier"].(string)
			name := res["name"].(string)
			resMap = append(resMap, map[string]string{
				"name": name,
				"id":   id,
			})
		}
		en := json.NewEncoder(os.Stdout)
		en.Encode(resMap)
	}
}

// Validate implements types.Command
func (lo *ListOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())
	return nil
}

// (TODO:kamesh) update the examples
var listCommandExample = fmt.Sprintf(`
# List existing delegates
%[1]s delegate list --account-id <your account id> --project-id <project id> --tag "foo" --tag "bar"
`, common.ExamplePrefix())

// newListCommand instantiates the new instance of the delegate list command
func newListCommand() *cobra.Command {
	lo := &ListOptions{}

	ldCmd := &cobra.Command{
		Use:     "list",
		Short:   "List existing delegates by tag",
		Example: listCommandExample,
		RunE:    lo.Execute,
		PreRunE: lo.Validate,
	}

	lo.AddFlags(ldCmd)

	return ldCmd
}

var _ types.Command = (*ListOptions)(nil)
var _ types.RESTCall = (*List)(nil)
