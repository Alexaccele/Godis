package http

import (
	"Godis/cache"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type cacheHandle struct {
	*Server
}

func (c *cacheHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := strings.Split(r.URL.EscapedPath(), "/")[2]
	switch r.Method {
	case http.MethodGet:
		bytes, err := c.Get(key)
		if err != nil{
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(bytes) == 0{
			w.WriteHeader(http.StatusNotFound)
		}
		w.Write(bytes)
		return
	case http.MethodPut:
		bytes, err := ioutil.ReadAll(r.Body)
		if err!=nil || len(bytes)==0{
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = c.Set(key, cache.Value{bytes,time.Now(),0})
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	case http.MethodDelete:
		err := c.Del(key)
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}


