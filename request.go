package ngebut

import "net/url"

type Request struct {
	method  string
	headers map[string]string
	body    []byte
	uri     *url.URL
}
