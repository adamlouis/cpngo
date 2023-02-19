package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"

	"github.com/adamlouis/cpngo/cpngo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	//go:embed web/*
	rootFS embed.FS
)

type Server struct {
	Port int
}

func (s *Server) Serve() error {
	webFS, err := fs.Sub(rootFS, "web")
	if err != nil {
		return err
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/fire", s.handlePostFire)
	r.Handle("/*", http.FileServer(http.FS(webFS)))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), r)
}

type RequestFire struct {
	Net cpngo.Net `json:"net"`
}

func (s *Server) handlePostFire(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	if err := r.Body.Close(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	req := &RequestFire{}
	if err := json.Unmarshal(body, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	rnr, err := cpngo.NewRunner(&cpngo.Net{
		Places:      req.Net.Places,
		Transitions: req.Net.Transitions,
		InputArcs:   req.Net.InputArcs,
		OutputArcs:  req.Net.OutputArcs,
		Tokens:      req.Net.Tokens,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if err := rnr.FireAny(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	ret, err := json.Marshal(rnr.Net())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(ret)
}
