package cache

import "unsafe"

type State struct {
	Count     int //缓存数量
	KeySize   int64
	ValueSize int64
}

func (s *State) Add(k string, v []byte) {
	s.Count++
	s.KeySize += int64(unsafe.Sizeof(k))
	s.ValueSize += int64(unsafe.Sizeof(v))
}

func (s *State) Del(k string, v []byte) {
	s.Count--
	s.KeySize -= int64(unsafe.Sizeof(k))
	s.ValueSize -= int64(unsafe.Sizeof(v))
}

//计算出占用内存
func (s *State) Memory() int64 {
	return s.KeySize + s.ValueSize
}
