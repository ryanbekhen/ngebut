package ngebut

import "net/url"

func (c *Context) Get(key string) string {
	return c.Request.headers[key]
}

func (c *Context) Param(key string) string {
	params, ok := c.router.FindParams(c.Request.method, c.Request.uri.Path)
	if !ok {
		return ""
	}

	return params[key]
}

func (c *Context) Params() map[string]string {
	params, ok := c.router.FindParams(c.Request.method, c.Request.uri.Path)
	if !ok {
		return nil
	}

	return params
}

func (c *Context) Query(key string) string {
	return c.Request.uri.Query().Get(key)
}

func (c *Context) Method() string {
	return c.Request.method
}

func (c *Context) Headers() map[string]string {
	return c.Request.headers
}

func (c *Context) Body() []byte {
	return c.Request.body
}

func (c *Context) URI() *url.URL {
	return c.Request.uri
}

func (c *Context) Path() string {
	return c.Request.uri.Path
}
