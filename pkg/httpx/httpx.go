package httpx

import (
	"net/http"
	"strings"
)

type RouteGroup struct {
	prefix string
	mux    *http.ServeMux
}

func AddNewRouteGroup(prefix string) *RouteGroup {
	prefix = strings.TrimRight(prefix, "/")
	return &RouteGroup{
		prefix: strings.TrimRight(prefix, "/"),
		mux:    http.NewServeMux(),
	}
}

func (rg *RouteGroup) Handle(path string, handler http.HandlerFunc) {
	rg.mux.HandleFunc(path, handler)
}

func (rg *RouteGroup) Inject(root *http.ServeMux) {
	root.Handle(rg.prefix+"/", http.StripPrefix(rg.prefix, rg.mux))
}

func (rg *RouteGroup) HandleFunc(path string, handlerFunc http.HandlerFunc) {
	rg.mux.HandleFunc(path, handlerFunc)
}

func (rg *RouteGroup) AddGroup(path string) *RouteGroup {
	return &RouteGroup{
		prefix: rg.prefix + "/" + strings.Trim(path, "/"),
		mux:    http.NewServeMux(),
	}
}
