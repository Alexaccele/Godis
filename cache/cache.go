package cache

type Cache interface {
	Get(k string)([]byte,error)
	Set(k string,v []byte) error
	Del(k string) error
	GetState() State
}


