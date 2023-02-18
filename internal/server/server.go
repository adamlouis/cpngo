package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adamlouis/cpngo/cpngo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Net *cpngo.Net
}

func (s *Server) Serve() error {
	if s.Net == nil {
		return fmt.Errorf("cannot run server - net is nil")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		s := s.Net.Summary()
		j, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(j)
	})
	r.Get("/fire", func(w http.ResponseWriter, r *http.Request) {
		if err := s.Net.FireAny(); err != nil {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			return
		}
		s := s.Net.Summary()
		j, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(j)
	})
	return http.ListenAndServe(":8080", r)
}
