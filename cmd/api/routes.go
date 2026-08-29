package main

import (
	"io/fs"
	"net/http"

	"visitor_analytics/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)
	mux.HandleFunc("POST /v1/links", app.createLinkHandler)
	mux.HandleFunc("GET /v1/links", app.listLinksHandler)
	mux.HandleFunc("GET /v1/links/{slug}", app.getLinkAnalyticsHandler)

	// Tracking redirect endpoint
	mux.HandleFunc("GET /r/{slug}", app.redirectLinkHandler)

	// Static UI assets via embed.FS
	staticFS, err := fs.Sub(ui.Files, "static")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// Serve index.html at root
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		indexHTML, err := staticFS.Open("index.html")
		if err != nil {
			app.notFoundResponse(w, r)
			return
		}
		defer indexHTML.Close()

		stat, err := indexHTML.Stat()
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		http.ServeContent(w, r, "index.html", stat.ModTime(), indexHTML.(ioReadSeeker))
	})

	return app.recoverPanic(app.enableCORS(app.logRequest(mux)))
}

type ioReadSeeker interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}
