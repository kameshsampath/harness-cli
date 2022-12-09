package delegate

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/kameshsampath/harness-cli/pkg/common"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ListOptions struct {
	ProjectID string
	Scope     string
	Tags      []string
}

type List struct {
	APIKey            string   `json:"-"`
	AccountID         string   `json:"-"`
	ProjectIdentifier string   `json:"-"`
	OrgID             string   `json:"-"`
	Scope             string   `json:"-"`
	Tags              []string `json:"tags"`
}

// AddFlags implements common.Command
func (lo *ListOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&lo.Scope, "delegate-scope", "", "project", `The scope of the delegate. Valid value is one of "project", "org", "account"`)
	cmd.Flags().StringSliceVarP(&lo.Tags, "tags", "t", []string{}, "The tags that will be used to filter the delegate")
}

// Call implements common.RESTCall
func (l *List) Call() (*resty.Response, error) {
	var resp *resty.Response
	var err error

	client := resty.New()
	req := client.R().
		EnableTrace().
		SetHeader("x-api-key", l.APIKey).
		SetQueryParam("accountIdentifier", l.AccountID)

	if l.Scope == "project" {
		req.SetQueryParam("orgIdentifier", l.OrgID)
		req.SetQueryParam("projectIdentifier", l.ProjectIdentifier)
	} else if l.Scope == "org" {
		req.SetQueryParam("orgIdentifier", l.OrgID)
	}

	log.Infof("Getting list of delegates for tags %v ", l.Tags)

	resp, err = req.
		SetHeader("Content-Type", "application/json").
		SetBody(l).
		Post("https://app.harness.io/gateway/ng/api/delegate-group-tags/delegate-groups")

	log.Tracef("URL %s", resp.Request.URL)
	log.Tracef("BODY %s", resp.Request.Body)

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

// Execute implements common.Command
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

	resp, err := l.Call()
	if err != nil {
		return err
	}
	var rm map[string]interface{}
	err = json.Unmarshal(resp.Body(), &rm)
	if err != nil {
		return err
	}
	log.Tracef("%#v", rm)
	if v, ok := rm["status"]; ok && v == "ERROR" {
		fmt.Printf(rm["message"].(string))
		return nil
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
	return err
}

// Validate implements common.Command
func (lo *ListOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())
	return nil
}

// (TODO:kamesh) update the examples
var listCommandExample = fmt.Sprintf(`
# List existing delegates
%[1]s delegate list --account-id <your account id> --project-id <project id> --tag "foo" --tag "bar"
`, common.ExamplePrefix())

// NewDelegatesListCommand instantiates the new instance of the NewDelegatesListCommand
func NewDelegatesListCommand() *cobra.Command {
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

var _ common.Command = (*ListOptions)(nil)
var _ common.RESTCall = (*List)(nil)
