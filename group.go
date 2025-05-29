package ngebut

// Group represents a group of routes with a common prefix and middleware.
type Group struct {
	prefix          string
	router          *Router
	middlewareFuncs []MiddlewareFunc
}

// Group creates a new route group with the given prefix.
func (r *Router) Group(prefix string) *Group {
	return &Group{
		prefix:          prefix,
		router:          r,
		middlewareFuncs: []MiddlewareFunc{},
	}
}

// Use adds middleware to the group.
// It accepts middleware functions that take a context parameter.
func (g *Group) Use(middleware ...interface{}) *Group {
	for _, m := range middleware {
		switch m := m.(type) {
		case Middleware:
			g.middlewareFuncs = append(g.middlewareFuncs, m)
		case func(*Ctx):
			g.middlewareFuncs = append(g.middlewareFuncs, m)
		default:
			panic("middleware must be a function that takes a *Ctx parameter")
		}
	}
	return g
}

// GET registers a new route with the GET method.
func (g *Group) GET(pattern string, handlers ...Handler) *Group {
	g.Handle(pattern, MethodGet, handlers...)
	return g
}

// HEAD registers a new route with the HEAD method.
func (g *Group) HEAD(pattern string, handlers ...Handler) *Group {
	g.Handle(pattern, MethodHead, handlers...)
	return g
}

// POST registers a new route with the POST method.
func (g *Group) POST(pattern string, handlers ...Handler) *Group {
	g.Handle(pattern, MethodPost, handlers...)
	return g
}

// PUT registers a new route with the PUT method.
func (g *Group) PUT(pattern string, handlers ...Handler) *Group {
	g.Handle(pattern, MethodPut, handlers...)
	return g
}

// DELETE registers a new route with the DELETE method.
func (g *Group) DELETE(pattern string, handlers ...Handler) *Group {
	g.Handle(pattern, MethodDelete, handlers...)
	return g
}

// CONNECT registers a new route with the CONNECT method.
func (g *Group) CONNECT(pattern string, handlers ...Handler) *Group {
	g.Handle(pattern, MethodConnect, handlers...)
	return g
}

// OPTIONS registers a new route with the OPTIONS method.
func (g *Group) OPTIONS(pattern string, handlers ...Handler) *Group {
	g.Handle(pattern, MethodOptions, handlers...)
	return g
}

// TRACE registers a new route with the TRACE method.
func (g *Group) TRACE(pattern string, handlers ...Handler) *Group {
	g.Handle(pattern, MethodTrace, handlers...)
	return g
}

// PATCH registers a new route with the PATCH method.
func (g *Group) PATCH(pattern string, handlers ...Handler) *Group {
	g.Handle(pattern, MethodPatch, handlers...)
	return g
}

// Handle registers a new route with the given pattern and method.
func (g *Group) Handle(pattern, method string, handlers ...Handler) *Group {
	// Prepend the group prefix to the pattern
	fullPattern := g.prefix
	if pattern != "" {
		if pattern[0] != '/' {
			fullPattern += "/"
		}
		fullPattern += pattern
	}

	// Register the route with the router, passing all handlers
	g.router.Handle(fullPattern, method, handlers...)
	return g
}

// Group creates a sub-group with the given prefix.
func (g *Group) Group(prefix string) *Group {
	// Prepend the parent group's prefix to the new group's prefix
	fullPrefix := g.prefix
	if prefix != "" {
		if prefix[0] != '/' {
			fullPrefix += "/"
		}
		fullPrefix += prefix
	}

	// Create a new group with the combined prefix and parent's router
	subGroup := &Group{
		prefix:          fullPrefix,
		router:          g.router,
		middlewareFuncs: make([]MiddlewareFunc, len(g.middlewareFuncs)),
	}

	// Copy the parent group's middleware to the new group
	copy(subGroup.middlewareFuncs, g.middlewareFuncs)

	return subGroup
}
