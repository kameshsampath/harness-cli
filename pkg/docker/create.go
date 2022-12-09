package docker

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
	connectorType     = "DockerRegistry"
	passwordAuthType  = "UsernamePassword"
	anonymousAuthType = "Anonymous"
)

type CreateOptions struct {
	// UsernamePassword or Anonymous
	AuthenticationType string
	Name               string
	ExecuteOnDelegate  bool
	ProjectID          string
	UserName           string
	Password           string
	Scope              string
	URL                string
	// "DockerHub" "Harbor" "Quay" "Other"
	ProviderType string
}

type UserNamePasswordAuth struct {
	UserName string `json:"username"`
	Password string `json:"passwordRef"`
}

type Authentication struct {
	Type string      `json:"type"`
	Spec interface{} `json:"spec,omitempty"`
}

type Spec struct {
	Authentication    Authentication `json:"auth"`
	URL               string         `json:"dockerRegistryUrl"`
	ExecuteOnDelegate bool           `json:"executeOnDelegate"`
	// "DockerHub" "Harbor" "Quay" "Other"
	ProviderType string `json:"providerType"`
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
	cmd.Flags().StringVarP(&co.UserName, "username", "u", "", "The docker registry user name.")
	cmd.MarkFlagRequired("username")
	cmd.Flags().StringVarP(&co.Password, "password", "", "", "The ID docker registry password secret. Use scope for finer access e.g. account.dockerhubpassword, org.dockerhubpassword etc.,")
	cmd.MarkFlagRequired("password")
	cmd.Flags().StringVarP(&co.AuthenticationType, "auth-type", "", "password", "The authentication type valid values are 'password' or 'anonymous' ")
	cmd.Flags().StringVarP(&co.URL, "registry-url", "", "https://registry.hub.docker.com/v2/", "The Docker Registry v2 URL e.g. https://registry.hub.docker.com/v2/")
	cmd.MarkFlagRequired("url")
	cmd.Flags().StringVarP(&co.ProviderType, "provider-type", "", "DockerHub", `The Docker Registry provider type. Valid values are "DockerHub" "Harbor" "Quay" "Other"`)
	cmd.Flags().StringVarP(&co.ProjectID, "project-id", "p", "", `The project where the connector will be created.`)
	cmd.Flags().StringVarP(&co.Scope, "connector-scope", "", "project", `The connector scope. Valid value is one of "project", "org", "account"`)
	cmd.Flags().BoolVarP(&co.ExecuteOnDelegate, "execute-on-delegate", "", true, "Allow the connector to execute on delegate.")
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

	// (TODO: kamesh) Enable API Access, Delegates
	spec := &Spec{
		URL:               co.URL,
		ProviderType:      co.ProviderType,
		ExecuteOnDelegate: co.ExecuteOnDelegate,
		Authentication: Authentication{
			Type: co.AuthenticationType,
		},
	}

	if co.AuthenticationType == "password" {
		spec.Authentication = Authentication{
			Type: passwordAuthType,
			Spec: UserNamePasswordAuth{
				UserName: co.UserName,
				Password: scopedName(co.Scope, co.Password),
			},
		}
	} else {
		spec.Authentication = Authentication{
			Type: anonymousAuthType,
			Spec: nil,
		}
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
	return nil
}

// (TODO:kamesh) update the examples
var newCommandExample = fmt.Sprintf(`
# Create new docker registry connector with username and password with default options
%[1]s docker-registry new --name foo --account-id <your account id> --project-id <project id> --username foo --password foo-password
# Create new docker registry connector with username and password with specific organization id
%[1]s docker-registry new --name foo --account-id <your account id> --project-id <project id> --username foo --password foo-password --org-id=<orgid>
# Create new docker registry connector with username and password at account scope
%[1]s docker-registry new --name foo --account-id <your account id> --username foo --password foo-password --connector-scope="account"
# Create new docker registry connector with username and password at org scope, default is project
%[1]s docker-registry new --name foo --account-id <your account id> --username foo --password foo-password  --org-id=<orgid> --connector-scope="org"
`, common.ExamplePrefix())

// NewDockerConnectorCommand instantiates the new instance of the NewDockerConnectorCommand
func NewDockerConnectorCommand() *cobra.Command {
	co := &CreateOptions{}

	dCmd := &cobra.Command{
		Use:     "new",
		Short:   "Creates a new Docker registry connector if not exists.",
		Example: newCommandExample,
		RunE:    co.Execute,
		PreRunE: co.Validate,
	}

	co.AddFlags(dCmd)

	return dCmd
}

var _ common.Command = (*CreateOptions)(nil)
var _ common.RESTCall = (*Connector)(nil)
