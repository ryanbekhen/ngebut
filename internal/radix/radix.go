package radix

import (
	"github.com/ryanbekhen/ngebut/internal/unsafe"
	"strings"
	"sync"
)

// PathMatchContext is a reusable context for path matching operations
// It pre-allocates memory for common operations to reduce allocations
type PathMatchContext struct {
	// Segments for path matching
	segments []string

	// Temporary byte slice for path operations
	pathBytes []byte

	// Reusable parameter map
	params map[string]string
}

// Reset resets the context for reuse
func (c *PathMatchContext) Reset() {
	// Clear segments without deallocating
	c.segments = c.segments[:0]

	// Clear pathBytes without deallocating
	c.pathBytes = c.pathBytes[:0]

	// Clear params without deallocating
	for k := range c.params {
		delete(c.params, k)
	}
}

// pathMatchContextPool is a pool of PathMatchContext objects
var pathMatchContextPool = sync.Pool{
	New: func() interface{} {
		return &PathMatchContext{
			segments:  make([]string, 0, 16),      // Pre-allocate for common path depth
			pathBytes: make([]byte, 0, 128),       // Pre-allocate for common path length
			params:    make(map[string]string, 8), // Pre-allocate for common number of params
		}
	},
}

// getPathMatchContext gets a PathMatchContext from the pool
func getPathMatchContext() *PathMatchContext {
	return pathMatchContextPool.Get().(*PathMatchContext)
}

// releasePathMatchContext returns a PathMatchContext to the pool
func releasePathMatchContext(ctx *PathMatchContext) {
	ctx.Reset()
	pathMatchContextPool.Put(ctx)
}

// segmentsPool is a pool of string slices for reuse when splitting paths
var segmentsPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 16) // Pre-allocate with capacity for 16 segments
	},
}

// getSegments gets a segments slice from the pool
func getSegments() []string {
	return segmentsPool.Get().([]string)
}

// releaseSegments returns a segments slice to the pool
func releaseSegments(s []string) {
	// Clear the slice without deallocating
	s = s[:0]
	segmentsPool.Put(s)
}

// Kind represents the type of node in the radix tree
type Kind uint8

const (
	// Static represents a static path segment
	Static Kind = iota
	// Param represents a parameter path segment (e.g., :id)
	Param
	// Wildcard represents a wildcard path segment (e.g., *)
	Wildcard
)

// Node represents a node in the radix tree
type Node struct {
	// Path is the path segment this node represents
	Path string
	// Kind is the type of node (static, param, wildcard)
	Kind Kind
	// Children are the child nodes
	Children []*Node
	// Handlers are the handlers for this node, indexed by HTTP method
	Handlers map[string]interface{}
	// ParamName is the name of the parameter (for Param nodes)
	ParamName string
	// IsEnd indicates if this node is the end of a route
	IsEnd bool
}

// NewNode creates a new radix tree node
func NewNode() *Node {
	return &Node{
		Children: make([]*Node, 0),
		Handlers: make(map[string]interface{}),
	}
}

// Tree represents a radix tree for routing
type Tree struct {
	Root *Node
}

// NewTree creates a new radix tree
func NewTree() *Tree {
	return &Tree{
		Root: NewNode(),
	}
}

// Insert adds a route to the radix tree
func (t *Tree) Insert(path string, method string, handler interface{}) {
	if path == "" {
		return
	}

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Split the path into segments
	segments := splitPath(path)
	defer releaseSegments(segments) // Release the segments slice back to the pool when done

	// Start at the root node
	current := t.Root

	// Traverse the tree and insert nodes as needed
	for i, segment := range segments {
		if segment == "" {
			continue
		}

		// Determine the kind of segment
		var kind Kind
		var paramName string

		if segment[0] == ':' {
			kind = Param
			paramName = segment[1:] // Remove the : prefix
		} else if segment == "*" {
			kind = Wildcard
		} else {
			kind = Static
		}

		// Look for an existing child node that matches
		var matchingChild *Node
		for _, child := range current.Children {
			if child.Kind == kind && (kind != Static || child.Path == segment) {
				if kind == Param && child.ParamName != paramName {
					continue
				}
				matchingChild = child
				break
			}
		}

		// If no matching child was found, create a new one
		if matchingChild == nil {
			matchingChild = &Node{
				Path:      segment,
				Kind:      kind,
				Children:  make([]*Node, 0),
				Handlers:  make(map[string]interface{}),
				ParamName: paramName,
			}
			current.Children = append(current.Children, matchingChild)
		}

		// Move to the matching child
		current = matchingChild

		// If this is the last segment, mark it as the end of a route
		if i == len(segments)-1 {
			current.IsEnd = true
			current.Handlers[method] = handler
		}
	}
}

// Find searches for a route in the radix tree
func (t *Tree) Find(path string, params map[string]string) (map[string]interface{}, bool) {
	if path == "" {
		return nil, false
	}

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Split the path into segments
	segments := splitPath(path)

	// Start at the root node
	current := t.Root

	// Traverse the tree to find the matching node
	result, found := findNode(current, segments, 0, params)

	// Release the segments slice back to the pool
	releaseSegments(segments)

	return result, found
}

// FindBytesWithContext searches for a route in the radix tree using a byte slice path and a pre-allocated context
// This is the most optimized version that avoids all allocations
func (t *Tree) FindBytesWithContext(path []byte, ctx *PathMatchContext) (map[string]interface{}, bool) {
	if len(path) == 0 {
		return nil, false
	}

	// Use the pre-allocated segments slice
	segments := ctx.segments

	// Handle path directly as bytes to avoid string conversion
	if path[0] != '/' {
		// For paths without leading slash, add an empty segment at the beginning
		segments = append(segments, "")

		// Split the path manually to avoid allocations
		start := 0
		for i := 0; i < len(path); i++ {
			if path[i] == '/' {
				// Add segment to the slice using unsafe for zero-allocation conversion
				if i > start {
					segments = append(segments, unsafe.B2S(path[start:i]))
				} else {
					segments = append(segments, "")
				}
				start = i + 1
			}
		}

		// Add the last segment
		if start < len(path) {
			segments = append(segments, unsafe.B2S(path[start:]))
		} else if start == len(path) {
			segments = append(segments, "")
		}
	} else {
		// Path already has leading slash, use normal splitPath logic but inline it
		// to avoid an extra function call and string conversion

		// Remove trailing slash if present
		pathLen := len(path)
		if pathLen > 1 && path[pathLen-1] == '/' {
			pathLen--
			path = path[:pathLen]
		}

		// Split the path manually to avoid allocations
		start := 1 // Start after the leading slash
		for i := 1; i < pathLen; i++ {
			if path[i] == '/' {
				// Add segment to the slice using unsafe for zero-allocation conversion
				if i > start {
					segments = append(segments, unsafe.B2S(path[start:i]))
				} else {
					segments = append(segments, "")
				}
				start = i + 1
			}
		}

		// Add the last segment
		if start < pathLen {
			segments = append(segments, unsafe.B2S(path[start:pathLen]))
		} else if start == pathLen {
			segments = append(segments, "")
		}
	}

	// Update the context's segments
	ctx.segments = segments

	// Start at the root node
	current := t.Root

	// Traverse the tree to find the matching node
	result, found := findNode(current, segments, 0, ctx.params)

	return result, found
}

// FindBytes searches for a route in the radix tree using a byte slice path
// This avoids string conversion when processing HTTP requests
func (t *Tree) FindBytes(path []byte, params map[string]string) (map[string]interface{}, bool) {
	if len(path) == 0 {
		return nil, false
	}

	// Get a path match context from the pool
	ctx := getPathMatchContext()
	defer releasePathMatchContext(ctx)

	// Use the pre-allocated segments slice
	segments := ctx.segments

	// Handle path directly as bytes to avoid string conversion
	if path[0] != '/' {
		// For paths without leading slash, add an empty segment at the beginning
		segments = append(segments, "")

		// Split the path manually to avoid allocations
		start := 0
		for i := 0; i < len(path); i++ {
			if path[i] == '/' {
				// Add segment to the slice using unsafe for zero-allocation conversion
				if i > start {
					segments = append(segments, unsafe.B2S(path[start:i]))
				} else {
					segments = append(segments, "")
				}
				start = i + 1
			}
		}

		// Add the last segment
		if start < len(path) {
			segments = append(segments, unsafe.B2S(path[start:]))
		} else if start == len(path) {
			segments = append(segments, "")
		}
	} else {
		// Path already has leading slash, use normal splitPath logic but inline it
		// to avoid an extra function call and string conversion

		// Remove trailing slash if present
		pathLen := len(path)
		if pathLen > 1 && path[pathLen-1] == '/' {
			pathLen--
			path = path[:pathLen]
		}

		// Split the path manually to avoid allocations
		start := 1 // Start after the leading slash
		for i := 1; i < pathLen; i++ {
			if path[i] == '/' {
				// Add segment to the slice using unsafe for zero-allocation conversion
				if i > start {
					segments = append(segments, unsafe.B2S(path[start:i]))
				} else {
					segments = append(segments, "")
				}
				start = i + 1
			}
		}

		// Add the last segment
		if start < pathLen {
			segments = append(segments, unsafe.B2S(path[start:pathLen]))
		} else if start == pathLen {
			segments = append(segments, "")
		}
	}

	// Update the context's segments
	ctx.segments = segments

	// Start at the root node
	current := t.Root

	// Traverse the tree to find the matching node
	result, found := findNode(current, segments, 0, params)

	return result, found
}

// FindStatic searches for a static route in the radix tree without parameter extraction
// This is an optimization for routes without parameters
func (t *Tree) FindStatic(path string) (map[string]interface{}, bool) {
	if path == "" {
		return nil, false
	}

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Split the path into segments
	segments := splitPath(path)

	// Start at the root node
	current := t.Root

	// Traverse the tree to find the matching node
	result, found := findStaticNode(current, segments, 0)

	// Release the segments slice back to the pool
	releaseSegments(segments)

	return result, found
}

// FindStaticBytes searches for a static route in the radix tree using a byte slice path
// This avoids string conversion when processing HTTP requests
func (t *Tree) FindStaticBytes(path []byte) (map[string]interface{}, bool) {
	if len(path) == 0 {
		return nil, false
	}

	// Get a path match context from the pool
	ctx := getPathMatchContext()
	defer releasePathMatchContext(ctx)

	// Use the pre-allocated segments slice
	segments := ctx.segments

	// Handle path directly as bytes to avoid string conversion
	if path[0] != '/' {
		// For paths without leading slash, add an empty segment at the beginning
		segments = append(segments, "")

		// Split the path manually to avoid allocations
		start := 0
		for i := 0; i < len(path); i++ {
			if path[i] == '/' {
				// Add segment to the slice using unsafe for zero-allocation conversion
				if i > start {
					segments = append(segments, unsafe.B2S(path[start:i]))
				} else {
					segments = append(segments, "")
				}
				start = i + 1
			}
		}

		// Add the last segment
		if start < len(path) {
			segments = append(segments, unsafe.B2S(path[start:]))
		} else if start == len(path) {
			segments = append(segments, "")
		}
	} else {
		// Path already has leading slash, use normal splitPath logic but inline it
		// to avoid an extra function call and string conversion

		// Remove trailing slash if present
		pathLen := len(path)
		if pathLen > 1 && path[pathLen-1] == '/' {
			pathLen--
			path = path[:pathLen]
		}

		// Split the path manually to avoid allocations
		start := 1 // Start after the leading slash
		for i := 1; i < pathLen; i++ {
			if path[i] == '/' {
				// Add segment to the slice using unsafe for zero-allocation conversion
				if i > start {
					segments = append(segments, unsafe.B2S(path[start:i]))
				} else {
					segments = append(segments, "")
				}
				start = i + 1
			}
		}

		// Add the last segment
		if start < pathLen {
			segments = append(segments, unsafe.B2S(path[start:pathLen]))
		} else if start == pathLen {
			segments = append(segments, "")
		}
	}

	// Update the context's segments
	ctx.segments = segments

	// Start at the root node
	current := t.Root

	// Traverse the tree to find the matching node
	result, found := findStaticNode(current, segments, 0)

	return result, found
}

// findStaticNode recursively searches for a matching static node
// This is an optimization that avoids parameter extraction
func findStaticNode(node *Node, segments []string, index int) (map[string]interface{}, bool) {
	// If we've processed all segments, check if this is a valid endpoint
	if index >= len(segments) {
		if node.IsEnd {
			return node.Handlers, true
		}
		return nil, false
	}

	segment := segments[index]
	if segment == "" {
		// Skip empty segments
		return findStaticNode(node, segments, index+1)
	}

	// Only check static nodes for better performance
	for _, child := range node.Children {
		if child.Kind == Static && child.Path == segment {
			return findStaticNode(child, segments, index+1)
		}
	}

	// No static match found
	return nil, false
}

// findNode recursively searches for a matching node
func findNode(node *Node, segments []string, index int, params map[string]string) (map[string]interface{}, bool) {
	// If we've processed all segments, check if this is a valid endpoint
	if index >= len(segments) {
		if node.IsEnd {
			return node.Handlers, true
		}
		return nil, false
	}

	segment := segments[index]
	if segment == "" {
		// Skip empty segments
		return findNode(node, segments, index+1, params)
	}

	// Single pass through children with early returns for better performance
	for _, child := range node.Children {
		switch child.Kind {
		case Static:
			// Static nodes must match the segment exactly
			if child.Path == segment {
				return findNode(child, segments, index+1, params)
			}
		case Param:
			// Parameter nodes match any segment
			// Store the parameter value using unsafe conversion if possible
			if params != nil {
				// Use direct string assignment since segment is already a string
				// The compiler should optimize this
				params[child.ParamName] = segment
			}
			return findNode(child, segments, index+1, params)
		case Wildcard:
			// Wildcard matches all remaining segments
			if params != nil && child.ParamName != "" {
				// Join remaining segments if this is a named wildcard
				remainingPath := strings.Join(segments[index:], "/")
				params[child.ParamName] = remainingPath
			}
			return child.Handlers, child.IsEnd
		}
	}

	// No match found
	return nil, false
}

// splitPath splits a path into segments
func splitPath(path string) []string {
	// Remove trailing slash if present
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	// Get a segments slice from the pool
	segments := getSegments()

	// Convert path to byte slice without allocation
	pathBytes := unsafe.S2B(path)

	// Split the path manually to avoid allocations
	start := 0
	for i := 0; i < len(pathBytes); i++ {
		if pathBytes[i] == '/' {
			// Add segment to the slice
			if i > start {
				// Use unsafe to avoid allocation when slicing
				segments = append(segments, unsafe.B2S(pathBytes[start:i]))
			} else {
				segments = append(segments, "")
			}
			start = i + 1
		}
	}

	// Add the last segment
	if start < len(pathBytes) {
		// Use unsafe to avoid allocation when slicing
		segments = append(segments, unsafe.B2S(pathBytes[start:]))
	} else if start == len(pathBytes) {
		segments = append(segments, "")
	}

	return segments
}
