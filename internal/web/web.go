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

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type pageData struct {
	Subject string
}

func (p *Handler) Handle(mux *http.ServeMux) {
	fs := http.FileServerFS(staticFS)
	mux.Handle("/static/", fs)
	var tmpl = template.Must(template.ParseFS(tmplFS, "templates/*.html"))
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		data := &pageData{
			Subject: "phone number",
		}
		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			slog.ErrorContext(r.Context(), err.Error())
			return
		}
	})
}
