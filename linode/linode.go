package linode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// APIBaseURL is the URL of the linode API endpoint.
const APIBaseURL = "https://api.linode.com/"

// LinodeError represents an error returned from the Linode API.
type LinodeError struct {
	Code    int    `json:"ERRORCODE"`
	Message string `json:"ERRORMESSAGE"`
}

type linodeResponse struct {
	Errors []LinodeError `json:"ERRORARRAY"`
	Action string        `json:"ACTION"`
	Data   interface{}   `json:"DATA"`
}

// LinodeErrors represents a collection of errors from the Linode API.
type LinodeErrors []LinodeError

// Satisfy the error interface.
func (self *LinodeError) Error() string {
	return fmt.Sprintf("linode error %d: %s", self.Code, self.Message)
}

// Satisfy the error interface.
func (self LinodeErrors) Error() string {
	slice := []LinodeError(self)
	strs := make([]string, len(slice))

	for i, err := range slice {
		strs[i] = err.Error()
	}

	return strings.Join(strs, ", ")
}

// APIRequest holds the information required for a single linode API request.
type APIRequest struct {
	params url.Values
}

// NewAPIRequest creates a linode API request object
func NewAPIRequest(action, apiKey string, args map[string]interface{}) APIRequest {
	val := url.Values{}

	for k, v := range args {
		val[k] = []string{fmt.Sprint(v)}
	}
	val["api_action"] = []string{action}
	val["api_key"] = []string{apiKey}

	return APIRequest{val}
}

func getWholeBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// PerformRequest runs an API request and returns the response data.
func PerformRequest(request APIRequest) (interface{}, error) {
	resp, err := http.PostForm(APIBaseURL, request.params)
	if err != nil {
		return nil, err
	}

	body, err := getWholeBody(resp)
	if err != nil {
		return nil, err
	}

	var result linodeResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	switch len(result.Errors) {
	case 0:
		return result.Data, nil
	case 1:
		return nil, &result.Errors[0]
	default:
		return nil, LinodeErrors(result.Errors)
	}
}

// SingleResponse holds the information for one response of a batch.
type SingleResponse struct {
	Data   interface{}
	Errors []LinodeError
}

// Batch sends a group of API requests in a single call.
func Batch(requests []APIRequest) ([]SingleResponse, error) {
	var apiKey string
	params := url.Values{}

	if len(requests) == 0 {
		return []SingleResponse{}, nil
	}
	requestArray := make([]map[string]string, len(requests))
	apiKey = requests[0].params["api_key"][0]
	params["api_key"] = []string{apiKey}
	params["api_action"] = []string{"batch"}

	for i, req := range requests {
		m := map[string]string{}
		for k, v := range req.params {
			if k == "api_key" {
				if v[0] != apiKey {
					return nil, errors.New(
						"all requests in a batch must have the same api_key")
				}
			}
			m[k] = v[0]
		}
		requestArray[i] = m
	}

	if bytes, err := json.Marshal(requestArray); err != nil {
		return nil, err
	} else {
		params["api_requestArray"] = []string{string(bytes)}
	}

	resp, err := http.PostForm(APIBaseURL, params)
	if err != nil {
		return nil, err
	}

	body, err := getWholeBody(resp)
	if err != nil {
		return nil, err
	}

	results := make([]linodeResponse, len(requests))
	if err = json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	final := make([]SingleResponse, len(requests))
	for i, lr := range results {
		final[i] = SingleResponse{lr.Data, lr.Errors}
	}
	return final, nil
}
