package utils

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](array []T) Set[T] {
	set := make(Set[T], 0)
	for _, v := range array {
		set.Add(v)
	}
	return set
}

func (s Set[T]) Add(value T) {
	s[value] = struct{}{}
}

func (s Set[T]) Remove(value T) {
	delete(s, value)
}

func (s Set[T]) Contains(value T) bool {
	_, exists := s[value]
	return exists
}
