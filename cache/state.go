package cache

type State struct {
	Count     int //缓存数量
	KeySize   int64
	ValueSize int64
}

func (s *State) Add(k string, v []byte) {
	s.Count++
	s.KeySize += int64(len(k))
	s.ValueSize += int64(len(v))
}

func (s *State) Del(k string, v []byte) {
	s.Count--
	s.KeySize -= int64(len(k))
	s.ValueSize -= int64(len(v))
}

//计算出占用内存
func (s *State) Memory() int64 {
	return 8 * (s.KeySize + s.ValueSize)
}
