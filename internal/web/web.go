// Package web serves the embedded demo page and its static assets.
package web

import (
	"embed"
	"net/http"
)

//go:embed templates
var tmplFS embed.FS

//go:embed static
var staticFS embed.FS

// Web serves the demo page.
type Web struct{}

// New returns a Web.
func New() *Web {
	return &Web{}
}

// Routes registers GET / and GET /static/ on mux.
func (p *Web) Routes(mux *http.ServeMux) {
	fs := http.FileServerFS(staticFS)
	mux.Handle("GET /static/", fs)
	mux.HandleFunc(
        "GET /{$}",
        func(w http.ResponseWriter, r *http.Request) {
            http.ServeFileFS(w,r,tmplFS,"templates/index.html")
	    },
    )
}
