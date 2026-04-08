package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var spaFS embed.FS

// SPAHandler serves the embedded React SPA.
// Static files are served directly; all other routes fall back to index.html.
func SPAHandler() http.Handler {
	dist, _ := fs.Sub(spaFS, "dist")
	fileServer := http.FileServer(http.FS(dist))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try to serve the file directly.
		if path != "/" && !strings.HasSuffix(path, "/") {
			f, err := dist.(fs.ReadFileFS).ReadFile(strings.TrimPrefix(path, "/"))
			if err == nil && len(f) > 0 {
				// Cache static assets.
				if strings.HasPrefix(path, "/assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// Fallback to index.html for SPA routes.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
