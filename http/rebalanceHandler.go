package http

import (
	"bytes"
	"log"
	"net/http"
)

type rebalanceHandler struct {
	*Server
}

func (handler *rebalanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	log.Println("开始平衡节点")
	go handler.rebalance()
}

func (handler *rebalanceHandler) rebalance() {
	scanner := handler.NewScanner()
	defer scanner.Close()
	c := &http.Client{}
	for scanner.Scan() {
		k := scanner.Key()
		node, ok := handler.ShouldProcess(k)
		if !ok {
			url := "http://" + node + "/cache/" + k
			r, _ := http.NewRequest(http.MethodPut,
				url,
				bytes.NewReader(scanner.Value()))
			//log.Println(url)
			c.Do(r)
			handler.Del(k)
		}
	}
}
