package ngebut

import (
	"github.com/evanphx/wildcat"
	"github.com/panjf2000/gnet/v2"
)

type httpServer struct {
	gnet.BuiltinEventEngine

	addr         string
	multicore    bool
	eng          gnet.Engine
	router       *Router
	errorHandler ErrorHandlerFunc
}

func (hs *httpServer) OnBoot(eng gnet.Engine) gnet.Action {
	hs.eng = eng
	return gnet.None
}

func (hs *httpServer) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	c.SetContext(&httpParser{parser: wildcat.NewHTTPParser()})
	return nil, gnet.None
}

func (hs *httpServer) OnTraffic(c gnet.Conn) gnet.Action {
	codec := c.Context().(*httpParser)
	buf, _ := c.Next(-1)

	req, err := codec.parseRequest(buf)
	if err != nil {
		return gnet.Close
	}

	ctx := &Context{
		conn:    c,
		router:  hs.router,
		Request: req,
		Response: &Response{
			headers: make(map[string]string),
			status:  200,
			body:    []byte{},
		},
	}

	handlers, _, found := hs.router.Find(req.method, req.uri.Path)
	if !found {
		ctx.Response.status = 404
		ctx.Response.headers["Content-Type"] = "text/plain"
		ctx.Response.body = []byte("404 Not Found")
	} else {
		for _, handler := range handlers {
			if err := handler(ctx); err != nil {
				if err := hs.errorHandler(ctx, err); err != nil {
					ctx.Response.status = 500
					ctx.Response.headers["Content-Type"] = "text/plain"
					ctx.Response.body = []byte("500 Internal Server Error")
				}
				break
			}
		}
	}

	responseBytes := ctx.Response.serialize()
	_, _ = c.Write(responseBytes)
	return gnet.None
}
