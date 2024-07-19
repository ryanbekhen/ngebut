package ngebut

import (
	"net/textproto"
)

type Header map[string][]string

func (h Header) Add(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	h[key] = append(h[key], value)
}

func (h Header) Set(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	h[key] = []string{value}
}

func (h Header) Get(key string) string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	if values, ok := h[key]; ok {
		return values[0]
	}
	return ""
}

func (h Header) Del(key string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	delete(h, key)
}
