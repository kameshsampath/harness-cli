package gcp

import (
	"fmt"

	"github.com/kameshsampath/harness-cli/pkg/common"
	"github.com/kameshsampath/harness-cli/pkg/types"
	"github.com/kameshsampath/harness-cli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// AddFlags implements types.Command
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

// Execute implements types.Command
func (co *CreateOptions) Execute(cmd *cobra.Command, args []string) error {
	c := &types.Connector{
		APIKey:     viper.GetString("api-key"),
		AccountID:  viper.GetString("account-id"),
		Name:       co.Name,
		Identifier: utils.IDFromName(co.Name),
		Type:       connectorType,
		Scope:      co.Scope,
	}

	if co.Scope == "project" {
		c.OrgID = viper.GetString("org-id")
		c.ProjectID = co.ProjectID
	} else if co.Scope == "org" {
		c.OrgID = viper.GetString("org-id")
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

	c.Spec = spec

	ci := &types.ConnectorInfo{
		ConnectorInfo: *c,
	}

	ci.Call()

	return nil
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

// Validate implements types.Command
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

// newConnectorCommand instantiates the new instance of the newConnectorCommand
func newConnectorCommand() *cobra.Command {
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

var _ types.Command = (*CreateOptions)(nil)
