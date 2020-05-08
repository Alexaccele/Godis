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
	http.Handle("/rebalance", s.rebalanceHandle())
	http.Handle("/keys", s.keysAndValuesHandle())
	addr := fmt.Sprintf(":%v", port)
	server := &http.Server{Addr: addr, Handler: nil}
	go func() {
		select {
		case <-ctx.Done():
			server.Shutdown(ctx)
		}
	}()
	//log.Printf("http监听地址：%v\n",addr)
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

func (s *Server) rebalanceHandle() http.Handler {
	return &rebalanceHandler{s}
}

func (s *Server) keysAndValuesHandle() http.Handler {
	return &keysAndValuesHandler{s}
}
