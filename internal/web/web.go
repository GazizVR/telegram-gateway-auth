package web

import (
	"embed"
	"net/http"
)

//go:embed templates
var tmplFS embed.FS

//go:embed static
var staticFS embed.FS

type Web struct {}

func New() *Web {
	return &Web{}
}

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
