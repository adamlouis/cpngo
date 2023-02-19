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
	Net *cpngo.Net
}

func (s *Server) Serve() error {
	if s.Net == nil {
		return fmt.Errorf("cannot run server - net is nil")
	}
	webFS, err := fs.Sub(rootFS, "web")
	if err != nil {
		return err
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/fire", s.fireTransitionHandler)
	r.Handle("/*", http.FileServer(http.FS(webFS)))
	return http.ListenAndServe(":8080", r)
}

type RequestFire struct {
	Net cpngo.Summary `json:"net"`
}

func (s *Server) fireTransitionHandler(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println(string(body))

	req := &RequestFire{}
	if err := json.Unmarshal(body, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	net, err := cpngo.NewNet(
		req.Net.Places,
		req.Net.Transitions,
		req.Net.InputArcs,
		req.Net.OutputArcs,
		req.Net.Tokens,
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if err := net.FireAny(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	ret, err := json.Marshal(net.Summary())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(ret)
}
