package github

import (
	"fmt"

	"github.com/kameshsampath/harness-cli/pkg/common"
	"github.com/kameshsampath/harness-cli/pkg/types"
	"github.com/kameshsampath/harness-cli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CreateOptions struct {
	// Http or Ssh
	AuthenticationType string
	// Token or GithubApp or OAuth
	APIAccessType       string
	DelegateSelectors   []string
	Name                string
	ExecuteOnDelegate   bool
	EnableAPIAccess     bool
	ProjectID           string
	UserName            string
	PersonalAccessToken string
	Scope               string
	// Account or Repo
	URLType        string
	URL            string
	ValidationRepo string
}

type APIAccessTokenSpec struct {
	TokenRef string `json:"tokenRef"`
}

type APIAccess struct {
	// "GithubApp" "Token" "OAuth"
	Type string      `json:"type"`
	Spec interface{} `json:"spec"`
}
type PATCredentialsSpec struct {
	UserName            string `json:"username"`
	PersonalAccessToken string `json:"tokenRef"`
}
type HTTPCredentialsSpec struct {
	Type string             `json:"type"`
	Spec PATCredentialsSpec `json:"spec"`
}

type SSHCredentialsSpec struct {
}

type Authentication struct {
	Type string      `json:"type"`
	Spec interface{} `json:"spec"`
}

type Spec struct {
	Authentication Authentication `json:"authentication"`
	APIAccess      APIAccess      `json:"apiAccess,omitempty"`
	//Always GitHubConnector
	ConnectorType     string   `json:"connectorType"`
	URL               string   `json:"url"`
	ValidationRepo    string   `json:"validationRepo"`
	ExecuteOnDelegate bool     `json:"executeOnDelegate"`
	Type              string   `json:"type"`
	DelegateSelectors []string `json:"delegateSelectors,omitempty"`
}

// AddFlags implements types.Command
func (co *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&co.Name, "name", "n", "", "The name of the connector.")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&co.UserName, "username", "u", "", "The GitHub user name.")
	cmd.MarkFlagRequired("username")
	cmd.Flags().StringVarP(&co.PersonalAccessToken, "pat", "", "", "The GitHub user Personal access Token(PAT) secret ID. Use scope for finer access e.g. account.mypat, org.mypat etc.,")
	cmd.MarkFlagRequired("pat")
	cmd.Flags().StringVarP(&co.AuthenticationType, "auth-type", "", "Http", "The github authentication type. Valid values are Http or Ssh")
	cmd.Flags().StringVarP(&co.URL, "url", "", "", "The GitHub account URL e.g. https://github.com/org-name")
	cmd.MarkFlagRequired("url")
	cmd.Flags().StringVarP(&co.URLType, "url-type", "", "Account", "The GitHub account URL type. Valid values are Account, Repo")
	cmd.Flags().StringVarP(&co.ValidationRepo, "validation-repo", "", "", "The GitHub to validate the credentials. Typically the repo under your account")
	cmd.Flags().StringVarP(&co.ProjectID, "project-id", "p", "", `The project where the connector will be created.`)
	cmd.Flags().StringVarP(&co.Scope, "connector-scope", "", "project", `The connector scope. Valid value is one of "project", "org", "account"`)
	cmd.Flags().BoolVarP(&co.ExecuteOnDelegate, "execute-on-delegate", "", true, "Allow the connector to execute on available delegate.")
	cmd.Flags().BoolVarP(&co.EnableAPIAccess, "enable-api-access", "", true, "Enable GitHub API Access. Only token type supported.")
	cmd.Flags().StringVarP(&co.APIAccessType, "api-access-type", "", "Token", `GitHub API Access type. One of "Token" or "GithubApp" or "OAuth"`)
	cmd.Flags().StringSliceVarP(&co.DelegateSelectors, "delegate-tags", "", []string{}, `The delegate tags that will be used to select the available delegate that will be used by the connector.`)
}

// Execute implements types.Command
func (co *CreateOptions) Execute(cmd *cobra.Command, args []string) error {
	c := &types.Connector{
		APIKey:     viper.GetString("api-key"),
		AccountID:  viper.GetString("account-id"),
		Name:       co.Name,
		Identifier: utils.IDFromName(co.Name),
		Type:       "Github",
		Scope:      co.Scope,
	}

	if co.Scope == "project" {
		c.OrgID = viper.GetString("org-id")
		c.ProjectID = co.ProjectID
	} else if co.Scope == "org" {
		c.OrgID = viper.GetString("org-id")
	}

	// (TODO: kamesh) Enable API Access, Delegates
	spec := &Spec{
		ConnectorType:     fmt.Sprintf("%sConnector", c.Type),
		URL:               co.URL,
		ValidationRepo:    co.ValidationRepo,
		Type:              co.URLType,
		ExecuteOnDelegate: co.ExecuteOnDelegate,
		Authentication: Authentication{
			Type: co.AuthenticationType,
		},
	}

	if co.URLType == "Repo" {
		spec.ValidationRepo = co.URL
	}

	if co.AuthenticationType == "Http" {
		spec.Authentication.Spec = HTTPCredentialsSpec{
			Type: "UsernameToken",
			Spec: PATCredentialsSpec{
				UserName:            co.UserName,
				PersonalAccessToken: scopedName(co.Scope, co.PersonalAccessToken),
			},
		}
	} else if co.AuthenticationType == "Ssh" {
		spec.Authentication.Spec = SSHCredentialsSpec{}
	}

	if co.EnableAPIAccess {
		spec.APIAccess = APIAccess{
			Type: co.APIAccessType,
			Spec: APIAccessTokenSpec{
				TokenRef: scopedName(co.Scope, co.PersonalAccessToken),
			},
		}
	}

	if len(co.DelegateSelectors) > 0 {
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
	return nil
}

var ghcCommandExample = fmt.Sprintf(`
  # Create gh-connector with default options
  %[1]s github new --name github --account-id <your account id> 
  # Create project with specific organization id
  %[1]s github new --name github --account-id <your account id> --org-id=<orgid>
`, common.ExamplePrefix())

// newGitHubConnectorCommand instantiates the new instance of the newGitHubConnectorCommand
func newGitHubConnectorCommand() *cobra.Command {
	co := &CreateOptions{}

	ghcCmd := &cobra.Command{
		Use:     "new",
		Short:   "Creates a new GitHub connector if not exists.",
		Example: ghcCommandExample,
		RunE:    co.Execute,
		PreRunE: co.Validate,
	}

	co.AddFlags(ghcCmd)

	return ghcCmd
}

var _ types.Command = (*CreateOptions)(nil)
