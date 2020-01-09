package cache

type State struct {
	Count int//缓存数量
	KeySize int64
	ValueSize int64
}

func (s *State) Add(k string,v []byte){
	s.Count++
	s.KeySize += int64(len(k))
	s.ValueSize += int64(len(v))
}

func (s *State) Del(k string,v []byte)  {
	s.Count--
	s.KeySize -= int64(len(k))
	s.ValueSize -= int64(len(v))
}
