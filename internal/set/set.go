package set

type Set[T comparable] map[T]struct{}

func (s Set[T]) Insert(v T) {
	s[v] = struct{}{}
}

func (s Set[T]) InsertMany(v []T) {
	for _, i := range v {
		s.Insert(i)
	}
}

func (s Set[T]) Slice() []T {
	res := make([]T, 0, len(s))
	for k := range s {
		res = append(res, k)
	}
	return res
}
