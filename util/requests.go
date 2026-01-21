package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type HTTPError struct {
	StatusCode int
}

func (re HTTPError) Error() string {
	return fmt.Sprintf("server responded with %d: %v", re.StatusCode, http.StatusText(re.StatusCode))
}

type ServiceError struct {
	MessageField string // Only set don't read
	ErrorField   error  // Only set don't read
}

func (se ServiceError) Message() string {
	return se.MessageField
}

func (se ServiceError) Error() string {
	if se.ErrorField == nil {
		return ""
	}
	return se.ErrorField.Error()
}

// Helper function to generate a URL to the server (e.g. /some -> https://server.com/some)
func DefaultPath(path string) string {
	return fmt.Sprintf("http://%s%s", os.Getenv("LISTEN"), path)
}

// Helper function for creating headers with just the credential
func CredentialHeaders() Headers {
	return Headers{
		"Credential": GetCredential(),
	}
}

type Headers = map[string]string

// Send a post request to any URL with headers attached
func Post[T any](url string, body any, headers Headers) (T, error) {

	// Declared here so it can be returned as nil before it's actually used
	var data T

	// Encode body to JSON
	byteBody, err := json.Marshal(body)
	if err != nil {
		return data, err
	}

	// Set headers
	reqHeaders := http.Header{}
	reqHeaders.Set("Content-Type", "application/json")
	for key, value := range headers {
		reqHeaders.Set(key, value)
	}

	// Send the request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(byteBody))
	if err != nil {
		return data, err
	}
	req.Header = reqHeaders

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return data, err
	}

	// Use the extracted function to read and parse the response body
	return readResponseBody[T](res)
}

// Send a form-encoded post request to any URL with headers attached
func PostForm[T any](urlStr string, formData url.Values, headers Headers) (T, error) {

	// Declared here so it can be returned as nil before it's actually used
	var data T

	// Set headers
	reqHeaders := http.Header{}
	reqHeaders.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range headers {
		reqHeaders.Set(key, value)
	}

	// Send the request
	req, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(formData.Encode()))
	if err != nil {
		return data, err
	}
	req.Header = reqHeaders

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return data, err
	}

	// Use the extracted function to read and parse the response body
	return readResponseBody[T](res)
}

// Send a form-encoded post request that allows non-200 responses (for OAuth error handling)
func PostFormAllowErrors[T any](urlStr string, formData url.Values, headers Headers) (T, error) {

	// Declared here so it can be returned as nil before it's actually used
	var data T

	// Set headers
	reqHeaders := http.Header{}
	reqHeaders.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range headers {
		reqHeaders.Set(key, value)
	}

	// Send the request
	req, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(formData.Encode()))
	if err != nil {
		return data, err
	}
	req.Header = reqHeaders

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return data, err
	}

	// Use the extracted function to read and parse the response body without status check
	return readResponseBodyAllowErrors[T](res)
}

// Send a put request to any URL with headers attached
func Put[T any](url string, headers Headers) (T, error) {

	// Declared here so it can be returned as nil before it's actually used
	var data T

	// Set headers
	reqHeaders := http.Header{}
	for key, value := range headers {
		reqHeaders.Set(key, value)
	}

	// Send the request
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return data, err
	}
	req.Header = reqHeaders

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return data, err
	}

	// Use the extracted function to read and parse the response body
	return readResponseBody[T](res)
}

// Send a get request to any URL with headers attached
func Get[T any](url string, headers Headers) (T, error) {

	// Declared here so it can be returned as nil before it's actually used
	var data T

	// Set headers
	reqHeaders := http.Header{}
	for key, value := range headers {
		reqHeaders.Set(key, value)
	}

	// Send the request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return data, err
	}
	req.Header = reqHeaders

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return data, err
	}

	// Use the extracted function to read and parse the response body
	return readResponseBody[T](res)
}

// readResponseBody reads the HTTP response body and unmarshals it into the provided type
func readResponseBody[T any](res *http.Response) (T, error) {
	var data T
	defer res.Body.Close()

	// Make sure to properly handle a case where the status code != 200
	if res.StatusCode != http.StatusOK {
		return data, HTTPError{
			StatusCode: res.StatusCode,
		}
	}

	// Grab all bytes from the buffer
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, res.Body)
	if err != nil {
		return data, err
	}

	// Parse body into JSON
	err = json.Unmarshal(buf.Bytes(), &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

// readResponseBodyAllowErrors reads the HTTP response body and unmarshals it without checking status code
// This is useful for OAuth flows where error responses are valid JSON
func readResponseBodyAllowErrors[T any](res *http.Response) (T, error) {
	var data T
	defer res.Body.Close()

	// Grab all bytes from the buffer
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, res.Body)
	if err != nil {
		return data, err
	}

	// Parse body into JSON
	err = json.Unmarshal(buf.Bytes(), &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

// Get a url for a path on a server (with api version, etc.)
func ServerPath(server string, path string) string {

	// Make sure there is a protocol specified on the server
	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "https://" + server
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return server + path
}
