package router

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type Router struct {
	globalChain []func(http.Handler) http.Handler
	isSubRouter bool
	path        string
	routeChain  []func(http.Handler) http.Handler
	*http.ServeMux
}

func New() *Router {
	return &Router{ServeMux: http.NewServeMux()}
}

func (r *Router) Use(mw ...func(http.Handler) http.Handler) {
	if r.isSubRouter {
		r.routeChain = append(r.routeChain, mw...)
	} else {
		r.globalChain = append(r.globalChain, mw...)
	}
}

func (r *Router) Group(prefix string, fn func(r *Router)) {
	subRouter := &Router{
		routeChain:  slices.Clone(r.routeChain),
		isSubRouter: true,
		path:        r.path + prefix,
		ServeMux:    r.ServeMux,
	}
	fn(subRouter)
}

func (r *Router) HandleFunc(pattern string, h http.HandlerFunc) {
	r.Handle(pattern, h)
}

func (r *Router) Handle(pattern string, h http.Handler) {
	methods := []string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	}

	p := pattern
	for _, m := range methods {
		if strings.HasPrefix(p, string(m)) {
			fullPath := r.path + strings.TrimSpace(strings.TrimPrefix(p, m))
			p = fmt.Sprintf("%s %s", m, fullPath)
		}
	}

	for _, mw := range slices.Backward(r.routeChain) {
		h = mw(h)
	}
	r.ServeMux.Handle(p, h)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	var h http.Handler = r.ServeMux

	for _, mw := range slices.Backward(r.globalChain) {
		h = mw(h)
	}
	h.ServeHTTP(w, rq)
}
