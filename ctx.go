package ngebut

import (
	"github.com/panjf2000/gnet/v2"
)

type Context struct {
	conn     gnet.Conn
	router   *Router
	Request  *Request
	Response *Response
}

func (c *Context) Conn() gnet.Conn {
	return c.conn
}
