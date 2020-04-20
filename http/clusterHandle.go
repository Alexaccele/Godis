package http

import (
	"encoding/json"
	"log"
	"net/http"
)

type clusterHandler struct {
	*Server
}

func (handler *clusterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	members := handler.Members()
	b, err := json.Marshal(members)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
