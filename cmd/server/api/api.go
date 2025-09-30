package api

import (
	"fmt"
	"mirror-sync/pkg/constants"
	"mirror-sync/pkg/remote/obj"
	"net/http"
	"runtime"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type (
	HTTPServer struct {
		Server *http.Server
	}
)

func NewServer(port int) *HTTPServer {
	s := &HTTPServer{}
	router := chi.NewRouter()
	router.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		notFound("id not found", writer, request)
	})
	router.MethodNotAllowed(func(writer http.ResponseWriter, request *http.Request) {
		methodNotAllowed(writer, request)
	})
	router.Use(middleware.Logger)
	router.Use(recoverMiddleware)
	router.Use(middleware.GetHead)
	router.Use(middleware.Compress(5, "application/gzip"))
	router.Use(middleware.Heartbeat("/heartbeat"))
	router.Route("/api", func(routerAPI chi.Router) {
		routerAPI.Route("/v1", func(r chi.Router) {
			// Get information about the server
			r.Get("/version", s.Information)
			r.Route("/sync", func(r chi.Router) {
				
			})
		})
	})
	s.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	return s
}

func (s *HTTPServer) Information(w http.ResponseWriter, r *http.Request) {
	info := obj.SystemInformation{
		Version:        constants.Version,
		APIVersion:     constants.ApiVersion,
		GoVersion:      runtime.Version(),
		OSName:         runtime.GOOS,
		OSArchitecture: runtime.GOARCH,
	}
	ok(info, w, r)
}
