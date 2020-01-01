package types

import (
	"encoding/json"
)

// HTTPCode is a uint16 typedef used for the enumeration of HTTP response codes
type HTTPCode uint16

// LambdaResponse is a container used as the return value for the lambda function in its entirety
type LambdaResponse struct {
	StatusCode HTTPCode `json:"statusCode"`
	Body       string   `json:"body"`
}

// HTTPResponse describes the JSON format body for all HTTP responses
type HTTPResponse struct {
	Code    B2ResultCode `json:"code"`
	Payload interface{}  `json:"payload"`
}

// ToJSON returns the reponse as a json format string
// Returns a default if it cannot marshal the payload
func (r *HTTPResponse) ToJSON() (s string) {
	bytes, err := json.Marshal(*r)
	if err != nil {
		r.Code = ResponseMarshalError
		r.Payload = "Could not marshal response payload"
		bytes, _ = json.Marshal(*r)
	}

	return string(bytes)
}

// Make is a helper function for creating an HTTPResponse inline
func Make(code B2ResultCode, payload interface{}) (r HTTPResponse) {
	return
}

// MakeLambdaResponse is a helper function for creating an entire lambda response
// Returns a generic error response if marshalling fails
func MakeLambdaResponse(httpCode HTTPCode, b2code B2ResultCode, payload interface{}) LambdaResponse {
	lambdaResponse := LambdaResponse{
		StatusCode: httpCode,
		Body:       "",
	}

	httpresponse := HTTPResponse{
		Code:    b2code,
		Payload: payload,
	}

	httpresponseJSONString := httpresponse.ToJSON()
	if httpresponse.Code == ResponseMarshalError {
		lambdaResponse.StatusCode = 500
	}

	lambdaResponse.Body = httpresponseJSONString

	return lambdaResponse
}
