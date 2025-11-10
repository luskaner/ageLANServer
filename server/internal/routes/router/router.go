package router

import (
	"net/http"
)

type Group struct {
	parent *Group
	path   string
	mux    *http.ServeMux
}

func (g *Group) fullPath() string {
	if g.parent == nil {
		return g.path
	}
	return g.parent.fullPath() + g.path
}

func (g *Group) Subgroup(path string) *Group {
	return &Group{
		parent: g,
		path:   path,
		mux:    g.mux,
	}
}

func (g *Group) HandleFunc(method string, path string, handler http.HandlerFunc) {
	g.mux.HandleFunc(method+" "+g.fullPath()+path, handler)
}

func (g *Group) Handle(method string, path string, handler http.Handler) {
	g.mux.Handle(method+" "+g.fullPath()+path, handler)
}

func (g *Group) HandlePath(path string, handler http.Handler) {
	g.mux.Handle(g.fullPath()+path, handler)
}

type Router struct {
	group *Group
}

func (r *Router) initialize() {
	r.group = &Group{
		path: "",
		mux:  http.NewServeMux(),
	}
}

func (r *Router) Check(_ *http.Request) bool {
	return true
}

func (r *Router) Initialize(_ string) bool {
	return true
}

type Initializer interface {
	InitializeRoutes(gameId string, next http.Handler) http.Handler
	Name() string
}
