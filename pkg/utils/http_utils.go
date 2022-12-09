package utils

import (
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

const (
	headerAPIKey        = "x-api-key"
	queryParamAccountID = "accountIdentifier"
	queryParamOrgID     = "orgIdentifier"
	queryParamProjectID = "projectIdentifier"
)

// NewHTTPRequest builds and returns the HTTP Request using resty
// NewHTTPRequest also sets the mandatory headers required to make request
func NewHTTPRequest(apiKey, accountID string) *resty.Request {
	client := resty.New()
	resMap := new(map[string]interface{})
	return client.R().
		EnableTrace().
		SetHeader(headerAPIKey, apiKey).
		SetQueryParam(queryParamAccountID, accountID).
		SetResult(resMap).
		SetError(resMap)
}

// AddScopedIDQueryParams adds ID parameters like orgID, ProjectID as
// HTTP request parameter
func AddScopedIDQueryParams(req *resty.Request, scope, orgID, projectIdentifier string) {
	switch scope {
	case "project":
		req.SetQueryParam(queryParamOrgID, orgID)
		req.SetQueryParam(queryParamProjectID, projectIdentifier)
	case "org":
		req.SetQueryParam(queryParamOrgID, orgID)
	default:
		if orgID != "" {
			req.SetQueryParam(queryParamOrgID, orgID)
		}
		if projectIdentifier != "" {
			req.SetQueryParam(queryParamProjectID, projectIdentifier)
		}
	}
}

// PostJSON executes the POST HTTP method to post the JSON(body)
func PostJSON(req *resty.Request, url string, body interface{}) (map[string]interface{}, error) {
	resp, err := req.
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(url)

	if err != nil {
		return nil, err
	}

	log.Tracef("URL %s", resp.Request.URL)
	log.Tracef("BODY %s", resp.Request.Body)

	resMap := resp.Result().(*map[string]interface{})

	return *resMap, nil
}

// DeleteResourceByID deletes the resource by ID
// The request url should have a path parameter named "{id}"
func DeleteResourceByID(req *resty.Request, url, id string) (map[string]interface{}, error) {
	req.
		SetPathParams(map[string]string{
			"id": id,
		})

	resp, err := req.Delete(url)
	if err != nil {
		return nil, err
	}

	log.Tracef("URL %s", resp.Request.URL)
	log.Tracef("BODY %s", resp.Request.Body)

	resMap := resp.Result().(*map[string]interface{})
	log.Tracef("Response %#v", *resMap)

	return *resMap, nil
}
