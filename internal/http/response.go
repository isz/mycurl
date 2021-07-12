package http

type Response struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

func NewResponse(status int, headers map[string]string, body []byte) *Response {
	return &Response{
		Status:  status,
		Headers: headers,
		Body:    body,
	}
}
