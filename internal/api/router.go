package api

import (
	"io/fs"
	"net/http"
	"strings"
)

// NewRouter constructs the HTTP handler for PlanForge.
func NewRouter(staticFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/v1/plans/analyze", corsMiddleware(HandleAnalyze))
	mux.HandleFunc("/api/v1/plans/compare", corsMiddleware(HandleCompare))
	mux.HandleFunc("/api/v1/plans/export", corsMiddleware(HandleExport))
	mux.HandleFunc("/api/v1/samples", corsMiddleware(HandleSamples))
	mux.HandleFunc("/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","version":"1.0.0"}`))
	}))

	// Static assets handler
	if staticFS != nil {
		fileServer := http.FileServer(http.FS(staticFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}

			// Check if file exists in staticFS
			f, err := staticFS.Open(strings.TrimPrefix(r.URL.Path, "/"))
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}

			// SPA fallback to index.html
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<!DOCTYPE html><html><head><title>PlanForge</title></head><body style="background:#090d16;color:#f3f4f6;font-family:sans-serif;padding:32px;"><h2>PlanForge API Server</h2><p>Vite frontend running on dev port or building embedded assets.</p></body></html>`))
		})
	}

	return mux
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
