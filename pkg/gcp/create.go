package gcp

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/kameshsampath/harness-cli/pkg/common"
	"github.com/kameshsampath/harness-cli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

const (
	connectorType    = "Gcp"
	manualAuthType   = "ManualConfig"
	delegateAuthType = "InheritFromDelegate"
)

type CreateOptions struct {
	// manual or delegate
	AuthenticationType string
	Name               string
	ExecuteOnDelegate  bool
	ProjectID          string
	Scope              string
	SecretKey          string
	DelegateSelectors  []string
}

type ManualAuth struct {
	SecretKeyRef string `json:"secretKeyRef"`
}

type Authentication struct {
	Type string      `json:"type"`
	Spec interface{} `json:"spec,omitempty"`
}

type Spec struct {
	Authentication    Authentication `json:"credential"`
	DelegateSelectors []string       `json:"delegateSelectors,omitempty"`
	ExecuteOnDelegate bool           `json:"executeOnDelegate"`
}

type Connector struct {
	ConnectorInfo common.Connector `json:"connector"`
}

// Call implements common.RESTCall
func (c *Connector) Call() (*resty.Response, error) {
	b, _ := json.Marshal(c)
	log.Infof("Payload:%s", string(b))

	var resp *resty.Response
	var err error

	client := resty.New()
	req := client.R().
		EnableTrace().
		SetHeader("x-api-key", c.ConnectorInfo.APIKey).
		SetQueryParam("accountIdentifier", c.ConnectorInfo.AccountID)

	log.Infof("Creating Docker Registry Connector %s ", c.ConnectorInfo.Name)

	resp, err = req.
		SetHeader("Content-Type", "application/json").
		SetBody(c).
		Post("https://app.harness.io/gateway/ng/api/connectors")
	if err != nil {
		return nil, err
	}

	return resp, err
}

// AddFlags implements common.Command
func (co *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&co.Name, "name", "n", "", "The name of the connector.")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&co.AuthenticationType, "auth-type", "", "manual", "The authentication type valid values are 'manual' or 'delegate' ")
	cmd.Flags().StringVarP(&co.SecretKey, "secret-key", "", "", `The GCP Service Account Key secret name. This is required only when "auth-type" is "manual"`)
	cmd.Flags().StringVarP(&co.ProjectID, "project-id", "p", "", `The project where the connector will be created.`)
	cmd.Flags().StringVarP(&co.Scope, "connector-scope", "", "project", `The connector scope. Valid value is one of "project", "org", "account"`)
	cmd.Flags().BoolVarP(&co.ExecuteOnDelegate, "execute-on-delegate", "", true, "Allow the connector to execute on delegate.")
	cmd.Flags().StringSliceVarP(&co.DelegateSelectors, "delegate-tags", "", []string{}, `The delegate tags that will be used to select the available delegates when the "auth-type" is "delegate"`)
}

// Execute implements common.Command
func (co *CreateOptions) Execute(cmd *cobra.Command, args []string) error {
	ci := &common.Connector{
		APIKey:     viper.GetString("api-key"),
		AccountID:  viper.GetString("account-id"),
		Name:       co.Name,
		Identifier: utils.IDFromName(co.Name),
		Type:       connectorType,
		Scope:      co.Scope,
	}

	if co.Scope == "project" {
		ci.OrgID = viper.GetString("org-id")
		ci.ProjectID = co.ProjectID
	} else if co.Scope == "org" {
		ci.OrgID = viper.GetString("org-id")
	}

	spec := &Spec{
		ExecuteOnDelegate: co.ExecuteOnDelegate,
	}

	if co.AuthenticationType == "manual" {
		spec.Authentication = Authentication{
			Type: manualAuthType,
			Spec: ManualAuth{
				SecretKeyRef: scopedName(co.Scope, co.SecretKey),
			},
		}
	} else {
		spec.Authentication = Authentication{
			Type: delegateAuthType,
			Spec: nil,
		}
		spec.DelegateSelectors = co.DelegateSelectors
	}

	ci.Spec = spec

	c := &Connector{
		ConnectorInfo: *ci,
	}

	resp, err := c.Call()

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
		conn := data["connector"].(map[string]interface{})
		fmt.Println(conn["identifier"].(string))
	} else {
		if v, ok := rm["code"]; ok && v == "DUPLICATE_FIELD" {
			fmt.Printf("GitHub Connector with name '%s' already exists", co.Name)
			return nil
		}
		log.Errorf("%#v", rm)
	}

	return err
}

func scopedName(scope, name string) string {
	if scope == "account" {
		return fmt.Sprintf("account.%s", name)
	}

	if scope == "org" {
		return fmt.Sprintf("org.%s", name)
	}

	return name
}

// Validate implements common.Command
func (co *CreateOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())

	authType := co.AuthenticationType

	switch authType {
	case "manual":
		if co.SecretKey == "" {
			return fmt.Errorf("secret-key id is required")
		}
	case "delegate":
		if len(co.DelegateSelectors) == 0 {
			return fmt.Errorf(`at least one delegate ID "delegate" need to be specified`)
		}
	}

	return nil
}

// (TODO:kamesh) update the examples
var newCommandExample = fmt.Sprintf(`
# Create new GCP connector default options
%[1]s gcp new --name my-gcp --account-id <your account id> --project-id <project id> --secret-key my-gcp-key
`, common.ExamplePrefix())

// NewGCPConnectorCommand instantiates the new instance of the NewGCPConnectorCommand
func NewGCPConnectorCommand() *cobra.Command {
	co := &CreateOptions{}

	gcpCmd := &cobra.Command{
		Use:     "new",
		Short:   "Creates a new GCP connector if not exists.",
		Example: newCommandExample,
		RunE:    co.Execute,
		PreRunE: co.Validate,
	}

	co.AddFlags(gcpCmd)

	return gcpCmd
}

var _ common.Command = (*CreateOptions)(nil)
var _ common.RESTCall = (*Connector)(nil)
