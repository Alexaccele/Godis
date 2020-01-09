package http

import (
	"Godis/cache"
	"net/http"
)

type Server struct {
	cache.Cache
}

func NewServer(c cache.Cache) *Server {
	return &Server{c}
}

func (s *Server) Listen(port string)  {
	http.Handle("/cache/",s.cacheHandle())
	http.Handle("/status",s.statusHandle())
	http.ListenAndServe(port,nil)
}

func (s *Server) cacheHandle() http.Handler {
	return &cacheHandle{s}
}

func (s *Server) statusHandle() http.Handler {
	return &statusHandler{s}
}