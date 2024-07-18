package ngebut

import (
	"fmt"
	"net/http"
)

type Response struct {
	headers map[string]string
	body    []byte
	status  int
}

func (r *Response) serialize() []byte {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.status, http.StatusText(r.status))
	bodyLen := len(r.body)
	r.headers["Content-Length"] = fmt.Sprintf("%d", bodyLen)
	for key, value := range r.headers {
		response += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	response += "\r\n"
	response += string(r.body)
	return []byte(response)
}
