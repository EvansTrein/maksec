package rest

import (
	"fmt"
	"net/http"
)

type HandlerFunc func(c *Context)

type Route struct {
	Method      string
	Path        string
	Handler     HandlerFunc
	Middlewares []func(http.Handler) http.Handler
}

// Asset — статический ресурс (UI, файлы), раздаётся по префиксу пути.
type Asset struct {
	Prefix  string
	Handler http.Handler
}

type Router struct {
	version string
	service string
	routes  []*Route
	assets  []*Asset
}

func NewRouter(version, service string) *Router {
	r := &Router{
		version: version,
		service: service,
		routes:  make([]*Route, 0),
	}

	return r
}

func (r *Router) GetVersion() string {
	return r.version
}

func (r *Router) GetService() string {
	return r.service
}

func (r *Router) AddRoute(route *Route) {
	r.routes = append(r.routes, route)
}

func (r *Router) AddAsset(prefix string, handler http.Handler) {
	r.assets = append(r.assets, &Asset{Prefix: prefix, Handler: handler})
}

func (r *Router) GetRoutes() []*Route {
	return r.routes
}

func (r *Router) GetAssets() []*Asset {
	return r.assets
}

func (r *Router) pattern(route *Route) string {
	p := fmt.Sprintf("/%s/%s", r.service, r.version)
	if route.Path != "" {
		p += "/" + route.Path
	}
	return p
}
