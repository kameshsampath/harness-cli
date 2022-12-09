package secret

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kameshsampath/harness-cli/pkg/common"
	"github.com/kameshsampath/harness-cli/pkg/types"
	"github.com/kameshsampath/harness-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CreateOptions struct {
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

type Info struct {
	Secret Secret `json:"secret"`
}

// Run implements RESTCall
func (s *Secret) Call() (map[string]interface{}, error) {
	req := utils.NewHTTPRequest(s.APIKey, s.AccountID)
	utils.AddScopedIDQueryParams(req, s.Scope, s.OrgID, s.ProjectIdentifier)
	req.
		SetQueryParam("privateSecret", strconv.FormatBool(s.PrivateSecret))

	if s.Type == "SecretText" {
		s.Spec.SecretValue = s.Text
		return utils.PostJSON(req, "https://app.harness.io/gateway/ng/api/v2/secrets", Info{Secret: *s})
	}

	ms := Info{Secret: *s}
	b, err := json.Marshal(ms)
	if err != nil {
		return nil, err
	}

	resp, err := req.
		SetFiles(map[string]string{
			"file": s.File,
		}).
		SetFormData(map[string]string{
			"spec": string(b),
		}).
		Post("https://app.harness.io/gateway/ng/api/v2/secrets/files")

	log.Tracef("URL %s", resp.Request.URL)
	log.Tracef("BODY %s", resp.Request.Body)

	if err != nil {
		return nil, err
	}

	resMap := resp.Result().(*map[string]interface{})
	return *resMap, err
}

// Print implements Command
func (s *Secret) Print(rm map[string]interface{}, err error) {
	log.Tracef("Resp:%#v", rm)
	if err != nil {
		log.Errorf("%s", err)
		return
	}

	if v, ok := rm["status"]; ok && v == "SUCCESS" {
		log.Tracef("%#v", rm)
		data := rm["data"].(map[string]interface{})
		secret := data["secret"].(map[string]interface{})
		fmt.Printf("Secret with name '%s' created with ID '%s' ", s.Name, secret["identifier"].(string))
	} else {
		if v, ok := rm["code"]; ok && v == "DUPLICATE_FIELD" {
			fmt.Printf("Secret with name '%s' already exists", s.Name)
			return
		}
		log.Tracef("%#v", err)
		fmt.Printf("%s", rm["message"].(string))
	}
}

// AddFlags implements Command
func (co *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&co.Name, "name", "n", "", "The name of the secret to create.")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&co.Description, "description", "d", "", "The description for the project.")
	cmd.Flags().StringVarP(&co.ProjectID, "project-id", "p", "", `The project where the secret will be created.`)
	cmd.Flags().StringVarP(&co.File, "file", "m", "", `The file content corresponding to secret. Used only when secret type is "SecretFile" or  "SSHKey" or "WinRmCredentials"`)
	cmd.Flags().StringArrayVarP(&co.Tags, "tags", "t", []string{}, "The tags to attach to the project, in the format of key:value e.g. foo:bar.")
	cmd.Flags().StringVarP(&co.SecretManagerID, "secret-manager-id", "s", "harnessSecretManager", `The secret manager id to use.`)
	cmd.Flags().StringVarP(&co.Text, "text", "", "", `The Secret text if the secret type is SecretText.`)
	cmd.Flags().StringVarP(&co.Type, "secret-type", "", "SecretFile", `The secret type. Valid value is one of "SecretFile" "SecretText" "SSHKey" "WinRmCredentials"`)
	cmd.Flags().StringVarP(&co.Scope, "secret-scope", "", "project", `The secret scope. Valid value is one of "project", "org", "account"`)
	cmd.Flags().StringVarP(&co.ValueType, "secret-value-type", "", "Inline", `The secret value type, if the secret type is "SecretText". Valid value is one of "Inline", "Reference", "CustomSecretManagerValues"`)
}

// Execute implements Command
func (co *CreateOptions) Execute(cmd *cobra.Command, args []string) error {
	s := &Secret{
		APIKey:        viper.GetString("api-key"),
		AccountID:     viper.GetString("account-id"),
		Name:          co.Name,
		PrivateSecret: false,
		Identifier:    utils.IDFromName(co.Name),
		File:          co.File,
		Description:   co.Description,
		Type:          co.Type,
		Text:          co.Text,
		Scope:         co.Scope,
		Spec: Spec{
			SecretManagerID: co.SecretManagerID,
			Type:            fmt.Sprintf("%sSpec", co.Type),
			SecretValueType: co.ValueType,
		},
	}

	if co.Scope == "project" {
		s.OrgID = viper.GetString("org-id")
		s.ProjectIdentifier = co.ProjectID
	} else if co.Scope == "org" {
		s.OrgID = viper.GetString("org-id")
	}

	s.Tags = utils.TagMapFromStringArray(co.Tags)

	s.Print(s.Call())

	return nil
}

// Validate implements Command
func (co *CreateOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())

	st := viper.GetString("secret-type")

	if st != "SecretFile" && st != "SecretText" && st != "SSHKey" && st != "WinRmCredentials" {
		return fmt.Errorf(`"secret-type" should be one of "SecretFile" "SecretText" "SSHKey" "WinRmCredentials"`)
	}

	switch st {
	case "SecretText":
		if co.Text = viper.GetString("text"); co.Text == "" {
			return fmt.Errorf(`"text" is required when the secret type is "SecretText"`)
		}
	case "SecretFile":
		if co.File = viper.GetString("file"); co.File != "" {
			_, err := os.Stat(co.File)
			if err != nil {
				return err
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

// (TODO:kamesh) add more examples
var fileSecretCommandExample = fmt.Sprintf(`
  # Create new secret from file with default options
  %[1]s secret new --name foo --account-id <your account id> --project-id <project id> --file foo.txt
  # Create new secret from file with specific organization id
  %[1]s secret new --name foo --account-id <your account id> --project-id <project id> --file foo.txt --org-id=<orgid>
  # Create new secret from text 
  %[1]s secret new --name foo --account-id <your account id> --project-id <project id> --text foo --type=SecretText
  # Create new secret from text at account scope, default is project
  %[1]s secret new --name foo --account-id <your account id>  --text foo --type=SecretText --secret-scope="account"
  # Create new secret from text at org scope, default is project
  %[1]s secret new --name foo --account-id <your account id> --text foo --type=SecretText --secret-scope="org"
`, common.ExamplePrefix())

// newSecretCommand instantiates the new instance of the SecretCommand
func newSecretCommand() *cobra.Command {
	so := &CreateOptions{}

	sfCmd := &cobra.Command{
		Use:     "new",
		Short:   "Creates a new secret.",
		Example: fileSecretCommandExample,
		RunE:    so.Execute,
		PreRunE: so.Validate,
	}

	so.AddFlags(sfCmd)

	return sfCmd
}

var _ types.Command = (*CreateOptions)(nil)
var _ types.RESTCall = (*Secret)(nil)
