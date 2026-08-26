package api

import (
	"net/http"
	"os"
	"path/filepath"
)

// SetupRouter creates an http.Handler with all API routes and static asset serving.
func SetupRouter(h *Handler, webDir string) http.Handler {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/data", h.HandleData)
	mux.HandleFunc("/api/plant", h.HandlePlant)
	mux.HandleFunc("/api/inverters", h.HandleInverters)
	mux.HandleFunc("/api/inverters/", h.HandleInverters)
	mux.HandleFunc("/api/health", h.HandleHealth)
	mux.HandleFunc("/api/stream", h.HandleStream)

	// Static Web Dashboard
	absWebDir, err := filepath.Abs(webDir)
	if err != nil {
		absWebDir = webDir
	}

	fileServer := http.FileServer(http.Dir(absWebDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			indexPath := filepath.Join(absWebDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	// Wrap in middleware chain
	var handler http.Handler = mux
	handler = CORSMiddleware(handler)
	handler = LoggingMiddleware(handler)
	handler = RecoveryMiddleware(handler)

	return handler
}
