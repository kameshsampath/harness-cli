package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ProjectOptions struct {
	Name        string
	Description string
	Modules     []string
	Tags        []string
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
func (p *Project) Call() (*resty.Response, error) {
	client := resty.New()
	return client.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetHeader("x-api-key", p.APIKey).
		SetBody(map[string]Project{
			"project": *p,
		}).
		Post("https://app.harness.io/gateway/ng/api/projects")
}

// AddFlags implements Command
func (po *ProjectOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&po.Name, "name", "n", "", "The name of the project to create.")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&po.Name, "description", "d", "", "The description for the project.")
	cmd.Flags().StringSliceVarP(&po.Modules, "modules", "m", []string{"CI"}, `The modules to attach to the project. Valid values are "CD" "CI" "CV" "CF" "CE" "STO" "CORE" "PMS" "TEMPLATESERVICE" "GOVERNANCE" "CHAOS".`)
	cmd.Flags().StringArrayVarP(&po.Tags, "tags", "t", []string{""}, "The tags to attach to the project, in the format of key:value e.g. foo:bar.")
}

// Execute implements Command
func (po *ProjectOptions) Execute(cmd *cobra.Command, args []string) error {
	p := &Project{
		APIKey:      viper.GetString("api-key"),
		AccountID:   viper.GetString("account-id"),
		OrgID:       viper.GetString("org-id"),
		Name:        po.Name,
		Identifier:  idFromName(po.Name),
		Description: po.Description,
		Modules:     po.Modules,
	}

	if len(po.Tags) > 0 {
		tm := make(map[string]string, len(po.Tags))
		for _, t := range po.Tags {
			if t != "" {
				s := strings.Split(t, ":")
				tm[s[0]] = s[1]
			}
		}
		p.Tags = tm
	}

	resp, err := p.Call()
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
		data := rm["data"].(map[string]interface{})
		proj := data["project"].(map[string]interface{})
		log.Infoln(proj["identifier"].(string))
	} else {
		if v, ok := rm["code"]; ok && v == "DUPLICATE_FIELD" {
			fmt.Printf("Project with name '%s' already exists", po.Name)
			return nil
		}
		log.Errorf("%#v", rm)
	}
	return nil
}

// Validate implements Command
func (po *ProjectOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())

	if po.Modules = viper.GetStringSlice("modules"); len(po.Modules) > 0 {
		for _, m := range po.Modules {
			if m != "CI" && m != "CD" {
				return fmt.Errorf("module string should be one of CI or CD")
			}
		}
	}

	if len(po.Tags) > 0 {
		for _, t := range po.Tags {
			if !strings.Contains(t, ":") {
				return fmt.Errorf("tags should be of format 'key:value'")
			}
		}
	}

	return nil
}

var projectCommandExample = fmt.Sprintf(`
  # Create project with default options
  %[1]s project --name foo --account-id <your account id> 
  # Create project with specific organization id
  %[1]s project --name foo --account-id <your account id> --org-id=<orgid>
`, ExamplePrefix())

// NewStartCommand instantiates the new instance of the StartCommand
func NewProjectCommand() *cobra.Command {
	po := &ProjectOptions{}

	projCmd := &cobra.Command{
		Use:     "project",
		Short:   "Create a project if not exists.",
		Example: projectCommandExample,
		RunE:    po.Execute,
		PreRunE: po.Validate,
	}

	po.AddFlags(projCmd)

	return projCmd
}

var _ Command = (*ProjectOptions)(nil)
var _ RESTCall = (*Project)(nil)
