package common

type Connector struct {
	Name        string            `json:"name"`
	APIKey      string            `json:"-"`
	AccountID   string            `json:"accountIdentifier"`
	Identifier  string            `json:"identifier"`
	Description string            `json:"description,omitempty"`
	OrgID       string            `json:"orgIdentifier"`
	ProjectID   string            `json:"projectIdentifier,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Type        string            `json:"type"`
	Scope       string            `json:"-"`
	Spec        interface{}       `json:"spec"`
}
