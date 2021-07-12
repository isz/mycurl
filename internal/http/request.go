package http

import (
	"github.com/pkg/errors"
	urllib "net/url"
	"strconv"
)

type request struct {
	method  Method
	url     *urllib.URL
	headers map[string]string
	body    []byte
}

type Method string

const (
	MethodGET  = "GET"
	MethodPOST = "POST"
)

var methods = [...]Method{MethodGET, MethodPOST}

func (method Method) IsValid() bool {
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

func NewRequest(method Method, url string, contentType string, body []byte) (*request, error) {

	if !method.IsValid() {
		return nil, errors.New("method is not valid")
	}

	urlParsed, err := urllib.Parse(url)
	if err != nil {
		return nil, errors.Wrap(err, "parse URL error")
	}

	if urlParsed.Scheme != "http" {
		return nil, errors.New("unsupported schema")
	}

	headers := make(map[string]string)

	length := len(body)
	if length != 0 {
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		headers["Content-Type"] = contentType
		headers["Content-Length"] = strconv.Itoa(length)
	}

	return &request{
		method:  method,
		url:     urlParsed,
		headers: headers,
		body:    body,
	}, nil
}

func (r *request) SetHeader(name, value string) {
	r.headers[name] = value
}
