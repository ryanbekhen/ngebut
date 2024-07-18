package ngebut

func (c *Context) Set(key, value string) *Context {
	c.Response.headers[key] = value
	return c
}

func (c *Context) SendString(body string) error {
	c.Response.body = []byte(body)
	return nil
}

func (c *Context) Status(status int) *Context {
	c.Response.status = status
	return c
}

func (c *Context) SendStatus(status int) error {
	c.Response.status = status
	return nil
}

func (c *Context) Send(body []byte) error {
	c.Response.body = body
	return nil
}
