package secret

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/kameshsampath/harness-cli/pkg/common"
	"github.com/kameshsampath/harness-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Options struct {
	Name        string
	Description string
	ProjectID   string
	File        string
	Tags        []string
}

type Spec struct {
	SecretManagerID string `json:"secretManagerIdentifier"`
}
type Secret struct {
	APIKey            string            `json:"-"`
	File              string            `json:"-"`
	OrgID             string            `json:"orgIdentifier"`
	AccountID         string            `json:"accountIdentifier"`
	Identifier        string            `json:"identifier"`
	ProjectIdentifier string            `json:"projectIdentifier"`
	PrivateSecret     bool              `json:"privateSecret"`
	Name              string            `json:"name"`
	Description       string            `json:"description,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	Type              string            `json:"type"`
	Spec              Spec              `json:"spec"`
}

// Run implements RESTCall
func (s *Secret) Call() (*resty.Response, error) {
	client := resty.New()

	ms := map[string]Secret{
		"secret": *s,
	}

	b, err := json.Marshal(ms)
	if err != nil {
		return nil, err
	}

	log.Infof("Payload:%s", string(b))

	resp, err := client.R().
		EnableTrace().
		SetHeader("x-api-key", s.APIKey).
		SetFiles(map[string]string{
			"file": s.File,
		}).
		SetFormData(map[string]string{
			"spec": string(b),
		}).
		SetQueryParam("accountIdentifier", s.AccountID).
		SetQueryParam("orgIdentifier", s.OrgID).
		SetQueryParam("projectIdentifier", s.ProjectIdentifier).
		SetQueryParam("privateSecret", strconv.FormatBool(s.PrivateSecret)).
		Post("https://app.harness.io/gateway/ng/api/v2/secrets/files")

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
func (so *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&so.Name, "name", "n", "", "The name of the project to create.")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&so.Description, "description", "d", "", "The description for the project.")
	cmd.Flags().StringVarP(&so.ProjectID, "project-id", "p", "", `The project where the secret will be created.`)
	cmd.Flags().StringVarP(&so.File, "file", "m", "", `The file content corresponding to secret.`)
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringArrayVarP(&so.Tags, "tags", "t", []string{}, "The tags to attach to the project, in the format of key:value e.g. foo:bar.")
}

// Execute implements Command
func (so *Options) Execute(cmd *cobra.Command, args []string) error {
	p := &Secret{
		APIKey:            viper.GetString("api-key"),
		AccountID:         viper.GetString("account-id"),
		OrgID:             viper.GetString("org-id"),
		Name:              so.Name,
		PrivateSecret:     false,
		Identifier:        utils.IDFromName(so.Name),
		ProjectIdentifier: so.ProjectID,
		File:              so.File,
		Description:       so.Description,
		Type:              "SecretFile",
		Spec: Spec{
			SecretManagerID: "harnessSecretManager",
		},
	}

	p.Tags = utils.TagMapFromStringArray(so.Tags)

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
		proj := data["secret"].(map[string]interface{})
		log.Infoln(proj["identifier"].(string))
	} else {
		if v, ok := rm["code"]; ok && v == "DUPLICATE_FIELD" {
			fmt.Printf("Secret with name '%s' already exists", so.Name)
			return nil
		}
		log.Errorf("%#v", rm)
	}
	return nil
}

// Validate implements Command
func (so *Options) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())

	if so.File = viper.GetString("file"); so.File != "" {
		_, err := os.Stat(so.File)
		if err != nil {
			return err
		}
	}

	if len(so.Tags) > 0 {
		for _, t := range so.Tags {
			if !strings.Contains(t, ":") {
				return fmt.Errorf("tags should be of format 'key:value'")
			}
		}
	}

	return nil
}

var fileSecretCommandExample = fmt.Sprintf(`
  # Create new secret from file with default options
  %[1]s new-file-secret --name foo --account-id <your account id> --project-id <project id> --file foo.txt
  # Create new secret from file with specific organization id
  %[1]s new-file-secret --name foo --account-id <your account id> --project-id <project id> --file foo.txt --org-id=<orgid>
`, common.ExamplePrefix())

// NewStartCommand instantiates the new instance of the StartCommand
func NewFileSecretCommand() *cobra.Command {
	so := &Options{}

	sfCmd := &cobra.Command{
		Use:     "new-file-secret",
		Short:   "Create a new file secret.",
		Example: fileSecretCommandExample,
		RunE:    so.Execute,
		PreRunE: so.Validate,
	}

	so.AddFlags(sfCmd)

	return sfCmd
}

var _ common.Command = (*Options)(nil)
var _ common.RESTCall = (*Secret)(nil)
