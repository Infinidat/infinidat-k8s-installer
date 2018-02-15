package lib

//map which maintains sequence
//nil key not supported
type LinkedMap struct {
	seq  []string
	data map[string]string
}

func (s *LinkedMap) Put(key, value string) {
	if s.data == nil {
		s.data = make(map[string]string)
	}
	s.seq = append(s.seq, key)
	s.data[key] = value
}
func (s *LinkedMap) Get(key string) (value string) {
	return s.data[key]
}
func (s *LinkedMap) GetIterator() func() (*string, *string) {
	i := -1
	return func() (k, v *string) {
		defer func() {
			r := recover()
			if r != nil {
				k, v = nil, nil
			}
		}()
		if len(s.seq) > i {
			i++
			k = &s.seq[i]
			value := s.data[*k]
			v = &value
		}
		return k, v
	}
}
