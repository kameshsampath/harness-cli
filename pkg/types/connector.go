package types

import (
	"fmt"

	"github.com/kameshsampath/harness-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
)

// ConnectorInfo is the wrapper to hold the connector details
type ConnectorInfo struct {
	ConnectorInfo Connector `json:"connector"`
}

// Connector holds the resource data of a Connector resource
// each connector varies with Spec, which is defined by respective connectors
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

// Call implements common.RESTCall
func (ci *ConnectorInfo) Call() (map[string]interface{}, error) {
	// b, _ := json.Marshal(c)
	// log.Infof("Payload:%s", string(b))
	req := utils.NewHTTPRequest(ci.ConnectorInfo.APIKey, ci.ConnectorInfo.AccountID)
	log.Infof(`Creating Connector %s of type "%s" `, ci.ConnectorInfo.Name, ci.ConnectorInfo.Type)
	ci.Print(utils.PostJSON(req, "https://app.harness.io/gateway/ng/api/connectors", ci))
	return nil, nil
}

// Call implements Command
func (ci *ConnectorInfo) Print(rm map[string]interface{}, err error) {
	if v, ok := rm["status"]; ok && v == "SUCCESS" {
		log.Tracef("%#v", rm)
		data := rm["data"].(map[string]interface{})
		conn := data["connector"].(map[string]interface{})
		fmt.Printf("%s Connector with name '%s' created with ID '%s' ", ci.ConnectorInfo.Type, ci.ConnectorInfo.Name, conn["identifier"].(string))
	} else {
		if v, ok := rm["code"]; ok && v == "DUPLICATE_FIELD" {
			fmt.Printf("GitHub Connector with name '%s' already exists", ci.ConnectorInfo.Name)
			return
		}
		log.Errorf("%#v", rm)
	}
}

var _ RESTCall = (*ConnectorInfo)(nil)
