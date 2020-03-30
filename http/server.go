package http

import (
	"Godis/cache"
	"context"
	"fmt"
	"net/http"
)

type Server struct {
	cache.Cache
}

func NewServer(c cache.Cache) *Server {
	return &Server{c}
}

func (s *Server) Listen(port string, ctx context.Context) {
	http.Handle("/cache/", s.cacheHandle())
	http.Handle("/status", s.statusHandle())
	server := &http.Server{Addr: fmt.Sprintf(":%v", port), Handler: nil}
	go func() {
		select {
		case <-ctx.Done():
			server.Shutdown(ctx)
		}
	}()
	server.ListenAndServe()
}

func (s *Server) cacheHandle() http.Handler {
	return &cacheHandle{s}
}

func (s *Server) statusHandle() http.Handler {
	return &statusHandler{s}
}
