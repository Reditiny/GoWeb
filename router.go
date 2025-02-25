package GoWeb

import (
	"log"
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, SLASH)

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	log.Printf("Route %4s - %s", method, pattern)

	parts := parsePattern(pattern)

	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &node{}
	}

	r.roots[method].insert(pattern, parts, 0)

	key := method + SLASH + pattern
	r.handlers[key] = handler
}

func (r *router) getRouteAndParams(method string, path string) (*node, map[string]string) {
	if _, ok := r.roots[method]; !ok {
		return nil, nil
	}

	searchParts := parsePattern(path)
	n := r.roots[method].search(searchParts, 0)
	if n == nil {
		return nil, nil
	}

	params := make(map[string]string)
	parts := parsePattern(n.pattern)
	for index, part := range parts {
		if part[0] == ':' {
			params[part[1:]] = searchParts[index]
		}
		if part[0] == '*' && len(part) > 1 {
			params[part[1:]] = strings.Join(searchParts[index:], "/")
			break
		}
	}
	return n, params
}

func (r *router) handle(c *Context) {
	n, params := r.getRouteAndParams(c.Method, c.Path)
	if n == nil {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		return
	}

	c.Params = params
	key := c.Method + SLASH + n.pattern
	c.handlers = append(c.handlers, r.handlers[key])
	c.Next()
}
