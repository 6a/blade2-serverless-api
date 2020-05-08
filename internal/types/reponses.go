// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package types defines types and contstants for this application.
package types

import (
	"encoding/json"
)

// HTTPCode is a uint16 typedef used for the enumeration of HTTP response codes.
type HTTPCode uint16

// LambdaResponse is a container used as the return value for the lambda function in its entirety.
type LambdaResponse struct {
	StatusCode HTTPCode `json:"statusCode"`
	Body       string   `json:"body"`
}

// HTTPResponse describes the JSON format body for all HTTP responses.
type HTTPResponse struct {
	Code    B2ResultCode `json:"code"`
	Payload interface{}  `json:"payload"`
}

// ToJSON returns the response body as a json format string. Returns a default value if it cannot
// marshal the payload.
func (r *HTTPResponse) ToJSON() (s string) {

	// Attempt to marshal the response - an error returns a specific default, with a marshal error
	// code and message indicating that there was an error.
	bytes, err := json.Marshal(*r)
	if err != nil {

		// On error, set some default values for this response object, and then re-marshal it.
		r.Code = ResponseMarshalError
		r.Payload = "Could not marshal response payload"
		bytes, _ = json.Marshal(*r)
	}

	return string(bytes)
}

// MakeLambdaResponse is a helper function for creating an entire lambda response
// Returns a generic error response if marshalling fails.
func MakeLambdaResponse(httpCode HTTPCode, b2code B2ResultCode, payload interface{}) LambdaResponse {

	// Create a new lambda response with the specified HTTP status code.
	lambdaResponse := LambdaResponse{
		StatusCode: httpCode,
		Body:       "",
	}

	// create a new HTTP response, with the code and payload specified.
	httpresponse := HTTPResponse{
		Code:    b2code,
		Payload: payload,
	}

	// Attempt to marshal the http message body. On failure, set the lambda status code to 500.
	httpresponseJSONString := httpresponse.ToJSON()
	if httpresponse.Code == ResponseMarshalError {
		lambdaResponse.StatusCode = 500
	}

	// Add the http response body string to the lambda response.
	lambdaResponse.Body = httpresponseJSONString

	return lambdaResponse
}
