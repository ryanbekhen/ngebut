package ngebut

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRouterGroup tests the Group method of Router
func TestRouterGroup(t *testing.T) {
	router := NewRouter()
	group := router.Group("/api")

	assert.NotNil(t, group, "Router.Group() returned nil")
	assert.Equal(t, "/api", group.prefix, "group.prefix doesn't match expected value")
	assert.Same(t, router, group.router, "group.router is not the same as the router")
	assert.Empty(t, group.middlewareFuncs, "group.middlewareFuncs should be empty")
}

// TestGroupUse tests the Use method of Group
func TestGroupUse(t *testing.T) {
	router := NewRouter()
	group := router.Group("/api")

	// Test with middleware function
	middleware1 := func(c *Ctx) {
		c.Next()
	}

	result := group.Use(middleware1)
	assert.Same(t, group, result, "Group.Use() did not return the group")
	assert.Len(t, group.middlewareFuncs, 1, "group.middlewareFuncs should have length 1")

	// Test with multiple middleware functions
	middleware2 := func(c *Ctx) {
		c.Next()
	}

	group.Use(middleware2)
	assert.Len(t, group.middlewareFuncs, 2, "group.middlewareFuncs should have length 2")

	// Test with invalid middleware (should panic)
	assert.Panics(t, func() {
		group.Use("not a middleware")
	}, "Group.Use() with invalid middleware should panic")
}

// TestGroupHandle tests the Handle method of Group
func TestGroupHandle(t *testing.T) {
	router := NewRouter()
	group := router.Group("/api")

	// Test with empty pattern
	handler := func(c *Ctx) {
		// Handler function
	}

	result := group.Handle("", "GET", handler)
	assert.Same(t, group, result, "Group.Handle() did not return the group")

	// Check that the route was added to the router
	assert.Len(t, router.Routes, 1, "router.Routes should have length 1")

	route := router.Routes[0]
	assert.Equal(t, "/api", route.Pattern, "route.Pattern doesn't match expected value")
	assert.Equal(t, "GET", route.Method, "route.Method doesn't match expected value")
	assert.Len(t, route.Handlers, 1, "route.Handlers should have length 1")

	// Test with pattern that starts with /
	group.Handle("/users", "POST", handler)
	assert.Len(t, router.Routes, 2, "router.Routes should have length 2")

	route = router.Routes[1]
	assert.Equal(t, "/api/users", route.Pattern, "route.Pattern doesn't match expected value")
	assert.Equal(t, "POST", route.Method, "route.Method doesn't match expected value")

	// Test with pattern that doesn't start with /
	group.Handle("items", "PUT", handler)
	assert.Len(t, router.Routes, 3, "router.Routes should have length 3")

	route = router.Routes[2]
	assert.Equal(t, "/api/items", route.Pattern, "route.Pattern doesn't match expected value")
	assert.Equal(t, "PUT", route.Method, "route.Method doesn't match expected value")

	// Test with multiple handlers
	handler2 := func(c *Ctx) {
		// Another handler
	}
	group.Handle("/multi", "DELETE", handler, handler2)
	assert.Len(t, router.Routes, 4, "router.Routes should have length 4")

	route = router.Routes[3]
	assert.Equal(t, "/api/multi", route.Pattern, "route.Pattern doesn't match expected value")
	assert.Equal(t, "DELETE", route.Method, "route.Method doesn't match expected value")
	assert.Len(t, route.Handlers, 2, "route.Handlers should have length 2")
}

// TestGroupHTTPMethods tests the HTTP method registration methods of Group
func TestGroupHTTPMethods(t *testing.T) {
	router := NewRouter()
	group := router.Group("/api")
	handler := func(c *Ctx) {}

	// Test GET
	result := group.GET("/users", handler)
	assert.Same(t, group, result, "Group.GET() did not return the group")
	assert.Len(t, router.Routes, 1, "router.Routes should have length 1")
	assert.Equal(t, "GET", router.Routes[0].Method, "router.Routes[0].Method doesn't match expected value")

	// Test HEAD
	group.HEAD("/users", handler)
	assert.Equal(t, "HEAD", router.Routes[1].Method, "router.Routes[1].Method doesn't match expected value")

	// Test POST
	group.POST("/users", handler)
	assert.Equal(t, "POST", router.Routes[2].Method, "router.Routes[2].Method doesn't match expected value")

	// Test PUT
	group.PUT("/users", handler)
	assert.Equal(t, "PUT", router.Routes[3].Method, "router.Routes[3].Method doesn't match expected value")

	// Test DELETE
	group.DELETE("/users", handler)
	assert.Equal(t, "DELETE", router.Routes[4].Method, "router.Routes[4].Method doesn't match expected value")

	// Test CONNECT
	group.CONNECT("/users", handler)
	assert.Equal(t, "CONNECT", router.Routes[5].Method, "router.Routes[5].Method doesn't match expected value")

	// Test OPTIONS
	group.OPTIONS("/users", handler)
	assert.Equal(t, "OPTIONS", router.Routes[6].Method, "router.Routes[6].Method doesn't match expected value")

	// Test TRACE
	group.TRACE("/users", handler)
	assert.Equal(t, "TRACE", router.Routes[7].Method, "router.Routes[7].Method doesn't match expected value")

	// Test PATCH
	group.PATCH("/users", handler)
	assert.Equal(t, "PATCH", router.Routes[8].Method, "router.Routes[8].Method doesn't match expected value")
}

// TestGroupSubGroup tests the Group method of Group
func TestGroupSubGroup(t *testing.T) {
	router := NewRouter()
	group := router.Group("/api")

	// Add middleware to the parent group
	middleware := func(c *Ctx) {
		c.Next()
	}
	group.Use(middleware)

	// Create a sub-group
	subGroup := group.Group("/v1")
	assert.NotNil(t, subGroup, "Group.Group() returned nil")
	assert.Equal(t, "/api/v1", subGroup.prefix, "subGroup.prefix doesn't match expected value")
	assert.Same(t, router, subGroup.router, "subGroup.router is not the same as the router")
	assert.Len(t, subGroup.middlewareFuncs, 1, "subGroup.middlewareFuncs should have length 1")

	// Test with empty prefix
	subGroup2 := group.Group("")
	assert.Equal(t, "/api", subGroup2.prefix, "subGroup2.prefix doesn't match expected value")

	// Test with prefix that starts with /
	subGroup3 := group.Group("/v2")
	assert.Equal(t, "/api/v2", subGroup3.prefix, "subGroup3.prefix doesn't match expected value")

	// Test with prefix that doesn't start with /
	subGroup4 := group.Group("v3")
	assert.Equal(t, "/api/v3", subGroup4.prefix, "subGroup4.prefix doesn't match expected value")

	// Test nested sub-groups
	nestedGroup := subGroup.Group("/users")
	assert.Equal(t, "/api/v1/users", nestedGroup.prefix, "nestedGroup.prefix doesn't match expected value")
}
