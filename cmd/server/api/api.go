package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"mirror-sync/cmd/server/core/storage"
	"mirror-sync/pkg/constants"
	"mirror-sync/pkg/project"
	"mirror-sync/pkg/remote/obj"
	"net/http"
	"runtime"

	cronruntime "mirror-sync/cmd/server/core/runtime"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type (
	HTTPServer struct {
		Server    *http.Server
		data      *storage.Repository
		scheduler *cronruntime.Scheduler
	}
)

func NewServer(data *storage.Repository, scheduler *cronruntime.Scheduler, addr string, port int) *HTTPServer {
	s := &HTTPServer{
		data:      data,
		scheduler: scheduler,
	}
	router := chi.NewRouter()
	router.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		notFound("id not found", writer, request)
	})
	router.MethodNotAllowed(func(writer http.ResponseWriter, request *http.Request) {
		methodNotAllowed(writer, request)
	})
	chi.RegisterMethod("EXECUTE")
	router.Use(middleware.Logger)
	router.Use(recoverMiddleware)
	router.Use(middleware.GetHead)
	router.Use(middleware.Compress(5, "application/gzip"))
	router.Use(middleware.Heartbeat("/heartbeat"))
	router.Route("/api", func(routerAPI chi.Router) {
		routerAPI.Route("/v1", func(r chi.Router) {
			// Get information about the server
			r.Get("/version", s.Information)
			r.MethodFunc("EXECUTE", "/run", s.RunProjectHandler)
			r.Route("/projects", func(r chi.Router) {
				r.Get("/all", s.ProjectsGetHandler)
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {})
				r.Post("/", s.ProjectPostHandler)
				r.Delete("/", s.ProjectDeleteHandler)
			})
		})
	})
	s.Server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
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

func (s *HTTPServer) ProjectPostHandler(w http.ResponseWriter, r *http.Request) {
	var pr project.Project
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&pr); err != nil {
		slog.Error("failed to parse project description", "err", err)
		internalServerError(err, w, r)
		return
	}

	if err := s.data.Save(pr); err != nil {
		slog.Error("failed to save project to the database", "err", err)
		internalServerError(err, w, r)
		return
	}

	s.scheduler.Remove(pr)
	if err := s.scheduler.Add(pr); err != nil {
		slog.Error("failed to run project", "err", err)
		internalServerError(err, w, r)
		return
	}

	w.WriteHeader(201)
}

func (s *HTTPServer) ProjectDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var pr project.Project
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&pr); err != nil {
		slog.Error("failed to parse project description", "err", err)
		internalServerError(err, w, r)
		return
	}

	s.scheduler.Remove(pr)
	if err := s.data.Remove(pr); err != nil {
		slog.Error("failed to remove project", "err", err, "uuid", pr.UUID)
		internalServerError(err, w, r)
		return
	}

	w.WriteHeader(204)
}

func (s *HTTPServer) ProjectsGetHandler(w http.ResponseWriter, r *http.Request) {
	prs, err := s.data.List()
	if err != nil {
		slog.Error("failed to fetch all the projects from the database", "err", err)
		internalServerError(err, w, r)
		return
	}

	ok(prs, w, r)
}

func (s *HTTPServer) RunProjectHandler(w http.ResponseWriter, r *http.Request) {
	var pr project.Project
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&pr); err != nil {
		slog.Error("failed to parse project description", "err", err)
		internalServerError(err, w, r)
		return
	}

	if err := s.scheduler.RunOnce(pr); err != nil {
		slog.Error("failed to run the project", "err", err)
		internalServerError(err, w, r)
		return
	}

	ok("ok", w, r)
}
