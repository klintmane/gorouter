package trails

import (
	"net/http"
	"strings"
)

type Router struct {
	routes    *route // Tree of route nodes
	wildcards map[string]http.HandlerFunc
	NotFound  http.HandlerFunc
}

func New() *Router {
	rootRoute := route{match: "/", isParam: false, methods: make(map[string]http.HandlerFunc)}
	return &Router{routes: &rootRoute, wildcards: make(map[string]http.HandlerFunc)}
}

func (r *Router) Handle(method, path string, handler http.HandlerFunc) {
	if path == "*" {
		r.wildcards[method] = handler
		return
	}
	if path[0] != '/' {
		panic("Path has to start with a /.")
	}
	r.routes.addNode(method, path, handler)
}

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Split the URL into parts
	parts := strings.Split(req.URL.Path, "/")
	length := len(parts)

	// Remove first empty string from the split and optionaly the last one
	if length > 2 && parts[length-1] == "" {
		parts = parts[1 : length-1]
	} else {
		parts = parts[1:]
	}

	route, _, ctx := router.routes.traverse(parts, req.Context())
	req = req.WithContext(ctx)

	if handler := route.methods[req.Method]; handler != nil {
		handler(w, req)
	} else if handler := router.wildcards[req.Method]; handler != nil {
		handler(w, req)
	} else if router.NotFound != nil {
		router.NotFound(w, req)
	}
}

func Param(r *http.Request, param string) string {
	p, ok := r.Context().Value(param).(string)
	if !ok {
		return ""
	}
	return p
}
