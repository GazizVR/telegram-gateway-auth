package web

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
)

//go:embed templates
var tmplFS embed.FS

//go:embed static
var staticFS embed.FS

type Web struct {
	pages map[string]*template.Template
}

func New() *Web {
	pages := make(map[string]*template.Template)
	var mainPage = template.Must(template.ParseFS(tmplFS, "templates/index.html"))
	pages["index.html"] = mainPage
	return &Web{
		pages: pages,
	}
}

type pageData struct {
	Subject string
}

func (p *Web) Routes(mux *http.ServeMux) {
	fs := http.FileServerFS(staticFS)
	mux.Handle("GET /static/", fs)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		data := &pageData{
			Subject: "номер телефона",
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page, ok := p.pages["index.html"]
		if ok {
			if err := page.ExecuteTemplate(w, "index.html", data); err != nil {
				slog.ErrorContext(r.Context(), err.Error())
				return
			}
		}
	})
}
