package ngebut

import (
	"github.com/panjf2000/gnet/v2"
	"io"
	"net/http"
)

type engine struct {
	gnet.BuiltinEventEngine

	addr      string
	multicore bool
	eng       gnet.Engine
	handler   Handler
}

func (hs *engine) OnBoot(eng gnet.Engine) gnet.Action {
	hs.eng = eng
	return gnet.None
}

func (hs *engine) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	return nil, gnet.None
}

func (hs *engine) OnTraffic(c gnet.Conn) gnet.Action {
	buf, _ := c.Next(-1)
	req, err := parseRequest(c, buf)
	if err != nil {
		return gnet.Close
	}

	resp := &responseWriter{
		conn: c,
		Response: &Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: Header{
				"Content-Type": []string{"text/plain; charset=utf-8"},
				"Date":         []string{http.TimeFormat},
				"Server":       []string{"ngebut"},
			},
			Body: io.NopCloser(nil),
		},
	}

	hs.handler.ServeHTTP(resp, req)
	return gnet.None
}

func (hs *engine) OnShutdown(eng gnet.Engine) {
	hs.eng = eng
}

func (hs *engine) OnClose(c gnet.Conn, err error) gnet.Action {
	return gnet.None
}
