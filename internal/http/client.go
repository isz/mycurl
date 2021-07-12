package http

import (
	"fmt"
	"github.com/pkg/errors"
	"mycurl/internal/tcp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHttpPort = "80"
)

type httpClient struct {
	timeout time.Duration
	resp    *Response
}

func NewHttpClient(timeout time.Duration) (*httpClient, error) {
	return &httpClient{
		timeout: timeout,
	}, nil
}

func (c *httpClient) Do(req *request) (*Response, error) {

	port := req.url.Port()
	if port == "" {
		port = defaultHttpPort
	}
	client, err := tcp.NewClient(req.url.Hostname(), port, c.timeout)

	if err != nil {
		return nil, errors.Wrap(err, "create TCP client error")
	}

	defer client.Close()

	return c.do(client, req)
}

func (c *httpClient) do(tcpClient *tcp.TcpClient, req *request) (*Response, error) {
	if err := c.doRequest(tcpClient, req); err != nil {
		return nil, errors.Wrap(err, "request error")
	}
	return c.getResponse(tcpClient)
}

func (c *httpClient) doRequest(tcpClient *tcp.TcpClient, req *request) error {
	path := req.url.Path
	if path == "" {
		path = "/"
	}
	r := []string{fmt.Sprintf("%s %s HTTP/1.1\r\nHost: %s", req.method, path, req.url.Hostname())}

	r = appendHeaders(r, req.headers)
	r = append(r, "\r\n")

	data := append([]byte(strings.Join(r, "\r\n")), req.body...)
	return tcpClient.Write(data)
}

func (c *httpClient) getResponse(tcpClient *tcp.TcpClient) (*Response, error) {
	var status statusReceiver
	if err := tcpClient.Read(&status); err != nil {
		return nil, errors.Wrap(err, "read start line error")
	}

	headers := headersReceiver(make(map[string]string))
	if err := tcpClient.Read(&headers); err != nil {
		return nil, errors.Wrap(err, "read headers error")
	}

	receiver, err := getReceiver(headers)
	if err != nil {
		return nil, err
	}

	if err := tcpClient.Read(receiver); err != nil {
		return nil, errors.Wrap(err, "read body error")
	}

	return NewResponse(status.Code(), headers, receiver.(body).getBody()), nil
}

func getReceiver(headers map[string]string) (tcp.Receiver, error) {
	encoding, exist := headers["Transfer-Encoding"]
	if !exist {
		return getBodyReceiver(headers)
	}

	return getChunkedBodyReceiver(encoding)
}

func getChunkedBodyReceiver(encoding string) (tcp.Receiver, error) {
	if encoding != "chunked" {
		return nil, errors.New("unsupported encoding")
	}
	return newChunkedBodyReceiver(), nil
}

func getBodyReceiver(headers map[string]string) (tcp.Receiver, error) {
	lengthStr, exist := headers["Content-Length"]
	if !exist {
		return nil, errors.New("content length header not found")
	}

	length, err := strconv.ParseInt(lengthStr, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "parse content length error")
	}

	return newBodyReceiver(length), nil
}

func appendHeaders(req []string, headers map[string]string) []string {
	for hName, hValue := range headers {
		req = append(req, fmt.Sprintf("%s: %s", hName, hValue))
	}
	return req
}
