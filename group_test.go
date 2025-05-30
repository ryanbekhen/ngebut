package ngebut

import (
	"testing"
)

// TestRouterGroup tests the Group method of Router
func TestRouterGroup(t *testing.T) {
	router := NewRouter()
	group := router.Group("/api")

	if group == nil {
		t.Fatal("Router.Group() returned nil")
	}

	if group.prefix != "/api" {
		t.Errorf("group.prefix = %q, want %q", group.prefix, "/api")
	}

	if group.router != router {
		t.Error("group.router is not the same as the router")
	}

	if len(group.middlewareFuncs) != 0 {
		t.Errorf("len(group.middlewareFuncs) = %d, want 0", len(group.middlewareFuncs))
	}
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
	if result != group {
		t.Error("Group.Use() did not return the group")
	}

	if len(group.middlewareFuncs) != 1 {
		t.Errorf("len(group.middlewareFuncs) = %d, want 1", len(group.middlewareFuncs))
	}

	// Test with multiple middleware functions
	middleware2 := func(c *Ctx) {
		c.Next()
	}

	group.Use(middleware2)
	if len(group.middlewareFuncs) != 2 {
		t.Errorf("len(group.middlewareFuncs) = %d, want 2", len(group.middlewareFuncs))
	}

	// Test with invalid middleware (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Group.Use() with invalid middleware did not panic")
		}
	}()
	group.Use("not a middleware")
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
	if result != group {
		t.Error("Group.Handle() did not return the group")
	}

	// Check that the route was added to the router
	if len(router.Routes) != 1 {
		t.Errorf("len(router.Routes) = %d, want 1", len(router.Routes))
	}

	route := router.Routes[0]
	if route.Pattern != "/api" {
		t.Errorf("route.Pattern = %q, want %q", route.Pattern, "/api")
	}
	if route.Method != "GET" {
		t.Errorf("route.Method = %q, want %q", route.Method, "GET")
	}
	if len(route.Handlers) != 1 {
		t.Errorf("len(route.Handlers) = %d, want 1", len(route.Handlers))
	}

	// Test with pattern that starts with /
	group.Handle("/users", "POST", handler)
	if len(router.Routes) != 2 {
		t.Errorf("len(router.Routes) = %d, want 2", len(router.Routes))
	}

	route = router.Routes[1]
	if route.Pattern != "/api/users" {
		t.Errorf("route.Pattern = %q, want %q", route.Pattern, "/api/users")
	}
	if route.Method != "POST" {
		t.Errorf("route.Method = %q, want %q", route.Method, "POST")
	}

	// Test with pattern that doesn't start with /
	group.Handle("items", "PUT", handler)
	if len(router.Routes) != 3 {
		t.Errorf("len(router.Routes) = %d, want 3", len(router.Routes))
	}

	route = router.Routes[2]
	if route.Pattern != "/api/items" {
		t.Errorf("route.Pattern = %q, want %q", route.Pattern, "/api/items")
	}
	if route.Method != "PUT" {
		t.Errorf("route.Method = %q, want %q", route.Method, "PUT")
	}

	// Test with multiple handlers
	handler2 := func(c *Ctx) {
		// Another handler
	}
	group.Handle("/multi", "DELETE", handler, handler2)
	if len(router.Routes) != 4 {
		t.Errorf("len(router.Routes) = %d, want 4", len(router.Routes))
	}

	route = router.Routes[3]
	if route.Pattern != "/api/multi" {
		t.Errorf("route.Pattern = %q, want %q", route.Pattern, "/api/multi")
	}
	if route.Method != "DELETE" {
		t.Errorf("route.Method = %q, want %q", route.Method, "DELETE")
	}
	if len(route.Handlers) != 2 {
		t.Errorf("len(route.Handlers) = %d, want 2", len(route.Handlers))
	}
}

// TestGroupHTTPMethods tests the HTTP method registration methods of Group
func TestGroupHTTPMethods(t *testing.T) {
	router := NewRouter()
	group := router.Group("/api")
	handler := func(c *Ctx) {}

	// Test GET
	result := group.GET("/users", handler)
	if result != group {
		t.Error("Group.GET() did not return the group")
	}
	if len(router.Routes) != 1 {
		t.Errorf("len(router.Routes) = %d, want 1", len(router.Routes))
	}
	if router.Routes[0].Method != "GET" {
		t.Errorf("router.Routes[0].Method = %q, want %q", router.Routes[0].Method, "GET")
	}

	// Test HEAD
	group.HEAD("/users", handler)
	if router.Routes[1].Method != "HEAD" {
		t.Errorf("router.Routes[1].Method = %q, want %q", router.Routes[1].Method, "HEAD")
	}

	// Test POST
	group.POST("/users", handler)
	if router.Routes[2].Method != "POST" {
		t.Errorf("router.Routes[2].Method = %q, want %q", router.Routes[2].Method, "POST")
	}

	// Test PUT
	group.PUT("/users", handler)
	if router.Routes[3].Method != "PUT" {
		t.Errorf("router.Routes[3].Method = %q, want %q", router.Routes[3].Method, "PUT")
	}

	// Test DELETE
	group.DELETE("/users", handler)
	if router.Routes[4].Method != "DELETE" {
		t.Errorf("router.Routes[4].Method = %q, want %q", router.Routes[4].Method, "DELETE")
	}

	// Test CONNECT
	group.CONNECT("/users", handler)
	if router.Routes[5].Method != "CONNECT" {
		t.Errorf("router.Routes[5].Method = %q, want %q", router.Routes[5].Method, "CONNECT")
	}

	// Test OPTIONS
	group.OPTIONS("/users", handler)
	if router.Routes[6].Method != "OPTIONS" {
		t.Errorf("router.Routes[6].Method = %q, want %q", router.Routes[6].Method, "OPTIONS")
	}

	// Test TRACE
	group.TRACE("/users", handler)
	if router.Routes[7].Method != "TRACE" {
		t.Errorf("router.Routes[7].Method = %q, want %q", router.Routes[7].Method, "TRACE")
	}

	// Test PATCH
	group.PATCH("/users", handler)
	if router.Routes[8].Method != "PATCH" {
		t.Errorf("router.Routes[8].Method = %q, want %q", router.Routes[8].Method, "PATCH")
	}
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
	if subGroup == nil {
		t.Fatal("Group.Group() returned nil")
	}

	if subGroup.prefix != "/api/v1" {
		t.Errorf("subGroup.prefix = %q, want %q", subGroup.prefix, "/api/v1")
	}

	if subGroup.router != router {
		t.Error("subGroup.router is not the same as the router")
	}

	if len(subGroup.middlewareFuncs) != 1 {
		t.Errorf("len(subGroup.middlewareFuncs) = %d, want 1", len(subGroup.middlewareFuncs))
	}

	// Test with empty prefix
	subGroup2 := group.Group("")
	if subGroup2.prefix != "/api" {
		t.Errorf("subGroup2.prefix = %q, want %q", subGroup2.prefix, "/api")
	}

	// Test with prefix that starts with /
	subGroup3 := group.Group("/v2")
	if subGroup3.prefix != "/api/v2" {
		t.Errorf("subGroup3.prefix = %q, want %q", subGroup3.prefix, "/api/v2")
	}

	// Test with prefix that doesn't start with /
	subGroup4 := group.Group("v3")
	if subGroup4.prefix != "/api/v3" {
		t.Errorf("subGroup4.prefix = %q, want %q", subGroup4.prefix, "/api/v3")
	}

	// Test nested sub-groups
	nestedGroup := subGroup.Group("/users")
	if nestedGroup.prefix != "/api/v1/users" {
		t.Errorf("nestedGroup.prefix = %q, want %q", nestedGroup.prefix, "/api/v1/users")
	}
}
