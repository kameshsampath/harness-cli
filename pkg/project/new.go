package project

import (
	"fmt"
	"strings"

	"github.com/kameshsampath/harness-cli/pkg/common"
	"github.com/kameshsampath/harness-cli/pkg/types"
	"github.com/kameshsampath/harness-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CreateOptions struct {
	Name        string
	Description string
	Modules     []string
	Tags        []string
}

type ProjectInfo struct {
	ProjectInfo Project `json:"project"`
}
type Project struct {
	APIKey      string            `json:"-"`
	OrgID       string            `json:"orgIdentifier"`
	AccountID   string            `json:"accountIdentifier"`
	Identifier  string            `json:"identifier"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Modules     []string          `json:"modules"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// Run implements RESTCall
func (p *Project) Call() (map[string]interface{}, error) {
	req := utils.NewHTTPRequest(p.APIKey, p.AccountID)
	return utils.PostJSON(req, "https://app.harness.io/gateway/ng/api/projects", ProjectInfo{ProjectInfo: *p})
}

// Validate implements types.Command
func (p *Project) Print(rm map[string]interface{}, err error) {
	if v, ok := rm["status"]; ok && v == "SUCCESS" {
		data := rm["data"].(map[string]interface{})
		proj := data["project"].(map[string]interface{})
		fmt.Println(proj["identifier"].(string))
	} else {
		if v, ok := rm["code"]; ok && v == "DUPLICATE_FIELD" {
			fmt.Printf("Project with name '%s' already exists", p.Name)
			return
		}
		log.Errorf("%#v", rm)
	}
}

// AddFlags implements Command
func (co *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&co.Name, "name", "n", "", "The name of the project to create.")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&co.Description, "description", "d", "", "The description for the project.")
	cmd.Flags().StringSliceVarP(&co.Modules, "modules", "m", []string{"CI"}, `The modules to attach to the project. Valid values are "CD" "CI" "CV" "CF" "CE" "STO" "CORE" "PMS" "TEMPLATESERVICE" "GOVERNANCE" "CHAOS".`)
	cmd.Flags().StringArrayVarP(&co.Tags, "tags", "t", []string{""}, "The tags to attach to the project, in the format of key:value e.g. foo:bar.")
}

// Execute implements Command
func (co *CreateOptions) Execute(cmd *cobra.Command, args []string) error {
	p := &Project{
		APIKey:      viper.GetString("api-key"),
		AccountID:   viper.GetString("account-id"),
		OrgID:       viper.GetString("org-id"),
		Name:        co.Name,
		Identifier:  utils.IDFromName(co.Name),
		Description: co.Description,
		Modules:     co.Modules,
	}

	p.Tags = utils.TagMapFromStringArray(co.Tags)

	p.Print(p.Call())

	return nil
}

// Validate implements Command
func (co *CreateOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())

	if co.Modules = viper.GetStringSlice("modules"); len(co.Modules) > 0 {
		for _, m := range co.Modules {
			if m != "CI" && m != "CD" {
				return fmt.Errorf("module string should be one of CI or CD")
			}
		}
	}

	if len(co.Tags) > 0 {
		for _, t := range co.Tags {
			if !strings.Contains(t, ":") {
				return fmt.Errorf("tags should be of format 'key:value'")
			}
		}
	}

	return nil
}

var projectCommandExample = fmt.Sprintf(`
  # Create project with default options
  %[1]s project new --name foo --account-id <your account id> 
  # Create project with specific organization id
  %[1]s project new --name foo --account-id <your account id> --org-id=<orgid>
`, common.ExamplePrefix())

// newProjectCommand instantiates the new instance of the NewProjectCommand
func newProjectCommand() *cobra.Command {
	po := &CreateOptions{}

	projCmd := &cobra.Command{
		Use:     "new",
		Short:   "Creates a new project.",
		Example: projectCommandExample,
		RunE:    po.Execute,
		PreRunE: po.Validate,
	}

	po.AddFlags(projCmd)

	return projCmd
}

var _ types.Command = (*CreateOptions)(nil)
var _ types.RESTCall = (*Project)(nil)
