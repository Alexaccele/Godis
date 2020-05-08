package cache

type Cache interface {
	Get(k string) ([]byte, error)
	Set(k string, v Value) error
	Del(k string) error
	KeysAndValues() ([]string, []Value)
	GetState() State
	NewScanner() Scanner
}
