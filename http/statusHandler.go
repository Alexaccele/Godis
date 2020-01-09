package http

import (
	"encoding/json"
	"net/http"
)

type statusHandler struct {
	*Server
}

func (s *statusHandler) ServeHTTP(w http.ResponseWriter,r *http.Request) {
	if r.Method != http.MethodGet{
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	bytes, err := json.Marshal(s.GetState())
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

