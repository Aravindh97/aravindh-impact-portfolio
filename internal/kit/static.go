package kit

import (
	_ "embed"
	"net/http"
)

//go:embed kit.css
var css []byte

// CSS returns the shared stylesheet.
func CSS() []byte { return css }

// HandleCSS registers GET /kit.css (the shared stylesheet) on mux.
func HandleCSS(mux *http.ServeMux) {
	mux.HandleFunc("GET /kit.css", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write(css)
	})
}

// HTML returns a handler that serves a static HTML page.
func HTML(page []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page)
	}
}
