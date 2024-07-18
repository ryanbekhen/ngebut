package ngebut

import (
	"net/url"
	"regexp"
	"strings"
)

type ErrorHandlerFunc func(*Context, error) error
type HandlerFunc func(*Context) error

type routeKey struct {
	method   string
	pattern  *regexp.Regexp
	template string
}

type Router struct {
	routes   []routeKey
	handlers map[routeKey][]HandlerFunc
}

func (r *Router) Add(method, path string, handlers ...HandlerFunc) {
	pattern, template := pathToPattern(path)
	key := routeKey{method: method, pattern: pattern, template: template}
	r.routes = append(r.routes, key)
	r.handlers[key] = handlers
}

func (r *Router) Find(method, path string) ([]HandlerFunc, map[string]string, bool) {
	for _, key := range r.routes {
		if key.method != method {
			continue
		}
		if key.pattern.MatchString(path) {
			params := extractParams(key.template, path)
			return r.handlers[key], params, true
		}
	}
	return nil, nil, false
}

func (r *Router) FindParams(method, path string) (map[string]string, bool) {
	for _, key := range r.routes {
		if key.method != method {
			continue
		}
		if key.pattern.MatchString(path) {
			params := extractParams(key.template, path)
			return params, true
		}
	}
	return nil, false
}

// pathToPattern converts a path with dynamic segments into a regex pattern
func pathToPattern(path string) (*regexp.Regexp, string) {
	// Replace dynamic segment placeholders like :id with a regex group
	pattern := strings.ReplaceAll(path, ":id", "([^/]+)")
	// Correctly handle an optional trailing slash
	pattern = "^" + pattern + "/?$"
	compiledPattern := regexp.MustCompile(pattern)
	return compiledPattern, path
}

// extractParams extracts parameters from the path based on the template
func extractParams(template, path string) map[string]string {
	params := make(map[string]string)
	tokens := strings.Split(template, "/")
	pathTokens := strings.Split(strings.TrimRight(path, "/"), "/")

	for i, token := range tokens {
		if strings.HasPrefix(token, ":") {
			paramName := token[1:]
			if i < len(pathTokens) {
				params[paramName], _ = url.QueryUnescape(pathTokens[i])
			}
		}
	}

	return params
}
