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
	Name            string
	Description     string
	File            string
	ProjectID       string
	SecretManagerID string
	Tags            []string
	// account, org, project
	Scope string
	// allowed "SecretFile" "SecretText" "SSHKey" "WinRmCredentials"
	Type string
	Text string
	// allowed values are "Inline", "Reference",CustomSecretManagerValues
	ValueType string
}

type Spec struct {
	ErrorMessageForInvalidYAML string `json:"errorMessageForInvalidYaml,omitempty"`
	SecretManagerID            string `json:"secretManagerIdentifier"`
	SecretValue                string `json:"value,omitempty"`
	SecretValueType            string `json:"valueType,omitempty"`
	Type                       string `json:"type,omitempty"`
}

type Secret struct {
	APIKey            string            `json:"-"`
	AccountID         string            `json:"accountIdentifier"`
	Description       string            `json:"description,omitempty"`
	File              string            `json:"-"`
	Identifier        string            `json:"identifier"`
	Name              string            `json:"name"`
	OrgID             string            `json:"orgIdentifier,omitempty"`
	ProjectIdentifier string            `json:"projectIdentifier,omitempty"`
	PrivateSecret     bool              `json:"privateSecret"`
	Spec              Spec              `json:"spec"`
	Scope             string            `json:"-"`
	Tags              map[string]string `json:"tags,omitempty"`
	Type              string            `json:"type"`
	Text              string            `json:"text,omitempty"`
}

// Run implements RESTCall
func (s *Secret) Call() (*resty.Response, error) {
	var resp *resty.Response
	var err error

	client := resty.New()
	req := client.R().
		EnableTrace().
		SetHeader("x-api-key", s.APIKey).
		SetQueryParam("accountIdentifier", s.AccountID).
		SetQueryParam("privateSecret", strconv.FormatBool(s.PrivateSecret))

	if s.Scope == "project" {
		req.SetQueryParam("orgIdentifier", s.OrgID)
		req.SetQueryParam("projectIdentifier", s.ProjectIdentifier)
	} else if s.Scope == "org" {
		req.SetQueryParam("orgIdentifier", s.OrgID)
	}

	log.Infof("Creating Secret %s of type %s", s.Name, s.Type)

	if s.Type == "SecretText" {
		s.Spec.SecretValue = s.Text
		ms := map[string]Secret{
			"secret": *s,
		}
		log.Debugf("Payload:%#v", ms)
		resp, err = req.
			SetHeader("Content-Type", "application/json").
			SetBody(ms).
			Post("https://app.harness.io/gateway/ng/api/v2/secrets")
		if err != nil {
			return nil, err
		}
	} else {
		ms := map[string]Secret{
			"secret": *s,
		}

		b, err := json.Marshal(ms)
		if err != nil {
			return nil, err
		}
		resp, err = req.
			SetFiles(map[string]string{
				"file": s.File,
			}).
			SetFormData(map[string]string{
				"spec": string(b),
			}).
			Post("https://app.harness.io/gateway/ng/api/v2/secrets/files")
		if err != nil {
			return nil, err
		}
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
	cmd.Flags().StringVarP(&so.Name, "name", "n", "", "The name of the secret to create.")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&so.Description, "description", "d", "", "The description for the project.")
	cmd.Flags().StringVarP(&so.ProjectID, "project-id", "p", "", `The project where the secret will be created.`)
	cmd.Flags().StringVarP(&so.File, "file", "m", "", `The file content corresponding to secret. Used only when secret type is "SecretFile" or  "SSHKey" or "WinRmCredentials"`)
	cmd.Flags().StringArrayVarP(&so.Tags, "tags", "t", []string{}, "The tags to attach to the project, in the format of key:value e.g. foo:bar.")
	cmd.Flags().StringVarP(&so.SecretManagerID, "secret-manager-id", "s", "harnessSecretManager", `The secret manager id to use.`)
	cmd.Flags().StringVarP(&so.Text, "text", "", "", `The Secret text if the secret type is SecretText.`)
	cmd.Flags().StringVarP(&so.Type, "secret-type", "", "SecretFile", `The secret type. Valid value is one of "SecretFile" "SecretText" "SSHKey" "WinRmCredentials"`)
	cmd.Flags().StringVarP(&so.Scope, "secret-scope", "", "project", `The secret scope. Valid value is one of "project", "org", "account"`)
	cmd.Flags().StringVarP(&so.ValueType, "secret-value-type", "", "Inline", `The secret value type, if the secret type is "SecretText". Valid value is one of "Inline", "Reference", "CustomSecretManagerValues"`)
}

// Execute implements Command
func (so *Options) Execute(cmd *cobra.Command, args []string) error {
	s := &Secret{
		APIKey:        viper.GetString("api-key"),
		AccountID:     viper.GetString("account-id"),
		Name:          so.Name,
		PrivateSecret: false,
		Identifier:    utils.IDFromName(so.Name),
		File:          so.File,
		Description:   so.Description,
		Type:          so.Type,
		Text:          so.Text,
		Scope:         so.Scope,
		Spec: Spec{
			SecretManagerID: so.SecretManagerID,
			Type:            fmt.Sprintf("%sSpec", so.Type),
			SecretValueType: so.ValueType,
		},
	}

	if so.Scope == "project" {
		s.OrgID = viper.GetString("org-id")
		s.ProjectIdentifier = so.ProjectID
	} else if so.Scope == "org" {
		s.OrgID = viper.GetString("org-id")
	}

	s.Tags = utils.TagMapFromStringArray(so.Tags)

	resp, err := s.Call()
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
		secret := data["secret"].(map[string]interface{})
		fmt.Println(secret["identifier"].(string))
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

	st := viper.GetString("secret-type")

	if st != "SecretFile" && st != "SecretText" && st != "SSHKey" && st != "WinRmCredentials" {
		return fmt.Errorf(`"secret-type" should be one of "SecretFile" "SecretText" "SSHKey" "WinRmCredentials"`)
	}

	switch st {
	case "SecretText":
		if so.Text = viper.GetString("text"); so.Text == "" {
			return fmt.Errorf(`"text" is required when the secret type is "SecretText"`)
		}
	case "SecretFile":
		if so.File = viper.GetString("file"); so.File != "" {
			_, err := os.Stat(so.File)
			if err != nil {
				return err
			}
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

// (TODO:kamesh) add more examples
var fileSecretCommandExample = fmt.Sprintf(`
  # Create new secret from file with default options
  %[1]s new-secret --name foo --account-id <your account id> --project-id <project id> --file foo.txt
  # Create new secret from file with specific organization id
  %[1]s new-secret --name foo --account-id <your account id> --project-id <project id> --file foo.txt --org-id=<orgid>
  # Create new secret from text 
  %[1]s new-secret --name foo --account-id <your account id> --project-id <project id> --text foo --type=SecretText
  # Create new secret from text at account scope, default is project
  %[1]s new-secret --name foo --account-id <your account id>  --text foo --type=SecretText --secret-scope="account"
  # Create new secret from text at org scope, default is project
  %[1]s new-secret --name foo --account-id <your account id> --text foo --type=SecretText --secret-scope="org"
`, common.ExamplePrefix())

// NewSecretCommand instantiates the new instance of the SecretCommand
func NewSecretCommand() *cobra.Command {
	so := &Options{}

	sfCmd := &cobra.Command{
		Use:     "new-secret",
		Short:   "Create a new secret.",
		Example: fileSecretCommandExample,
		RunE:    so.Execute,
		PreRunE: so.Validate,
	}

	so.AddFlags(sfCmd)

	return sfCmd
}

var _ common.Command = (*Options)(nil)
var _ common.RESTCall = (*Secret)(nil)
