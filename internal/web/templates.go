package web

import (
	"embed"
	"html/template"
	"io"
	"net/http"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

var (
	homeTmpl    *template.Template
	elementTmpl *template.Template
	errorTmpl   *template.Template
	badFormTmpl *template.Template
	badLargeTmpl *template.Template
	streamTmpls *template.Template
)

func init() {
	homeTmpl = mustParseWithLayout("templates/home.html")
	elementTmpl = mustParseWithLayout("templates/element.html")
	errorTmpl = mustParseWithLayout("templates/error.html")
	badFormTmpl = mustParseWithLayout("templates/bad.html")
	badLargeTmpl = mustParseWithLayout("templates/bad_large.html")
	streamTmpls = template.Must(
		template.ParseFS(templateFS,
			"templates/bad_head.html",
			"templates/bad_item.html",
			"templates/bad_done.html",
		),
	)
}

func mustParseWithLayout(page string) *template.Template {
	return template.Must(
		template.ParseFS(templateFS, "templates/layout.html", page),
	)
}

func render(w http.ResponseWriter, tmpl *template.Template, data templateData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "layout", data)
}

func renderStream(w io.Writer, name string, data templateData) {
	streamTmpls.ExecuteTemplate(w, name, data)
}
