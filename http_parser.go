package ngebut

import (
	"errors"
	"github.com/evanphx/wildcat"
	"net/url"
	"strings"
)

var ErrMalformedRequest = errors.New("malformed request")

type httpParser struct {
	parser *wildcat.HTTPParser
}

func (c *httpParser) parseRequest(buf []byte) (*Request, error) {
	headerEndOffset, err := c.parser.Parse(buf)
	if err != nil {
		return nil, err
	}

	headerLines := strings.Split(string(buf[:headerEndOffset]), "\r\n")
	if len(headerLines) < 1 {
		return nil, ErrMalformedRequest
	}

	firstLineParts := strings.Split(headerLines[0], " ")
	if len(firstLineParts) < 3 {
		return nil, ErrMalformedRequest
	}
	method, path := firstLineParts[0], firstLineParts[1]

	headers := make(map[string]string)
	for _, line := range headerLines[1:] {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}

	parsedURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	body := buf[headerEndOffset:]

	return &Request{
		method:  method,
		headers: headers,
		body:    body,
		uri:     parsedURL,
	}, nil
}
