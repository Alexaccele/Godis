package http

import (
	"fmt"
	"net/http"
)

type keysAndValuesHandler struct {
	*Server
}

func (s *keysAndValuesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	Keys, Values := s.KeysAndValues()
	for i := 0; i < len(Keys); i++ {
		w.Write([]byte(fmt.Sprintf("key:%v,value:%v,create time:%v,TTL:%v\n", Keys[i], string(Values[i].Val), Values[i].Created.Format("15:04:05"), Values[i].TTL)))
	}
}
