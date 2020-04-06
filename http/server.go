package http

import (
	"Godis/cache"
	"Godis/cluster"
	"context"
	"fmt"
	"net/http"
)

type Server struct {
	cache.Cache
	cluster.Node
}

func NewServer(c cache.Cache, node cluster.Node) *Server {
	return &Server{c, node}
}

func (s *Server) Listen(port string, ctx context.Context) {
	http.Handle("/cache/", s.cacheHandle())
	http.Handle("/status", s.statusHandle())
	http.Handle("/cluster", s.clusterHandle())
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

func (s *Server) clusterHandle() http.Handler {
	return &clusterHandler{s}
}
