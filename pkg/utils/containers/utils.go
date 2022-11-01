package containers

func Reverse[T any](arr []T) []T {
	length := len(arr)
	res := make([]T, length)

	for ind, val := range arr {
		res[length-ind-1] = val
	}

	return res
}

func Map[T any, V any](arr []T, f func(T) V) []V {
	res := make([]V, len(arr))
	for ind, val := range arr {
		res[ind] = f(val)
	}

	return res
}

func Filter[T any](arr []T, f func(T) bool) []T {
	res := make([]T, 0)
	for _, v := range arr {
		if f(v) {
			res = append(res, v)
		}
	}

	return res
}

func DFS[T comparable](initial T, neighbors func(T) ([]T, error)) ([]T, error) {
	res := make([]T, 0)
	seen := map[T]bool{}
	s := NewStack[T]()
	s.Push(initial)

	for s.Len() > 0 {
		r, err := s.Pop()
		if err != nil {
			return res, err
		}

		if _, ok := seen[r]; ok {
			continue
		}

		seen[r] = true
		res = append(res, r)
		nebs, err := neighbors(r)
		if err != nil {
			return res, err
		}

		for _, neb := range nebs {
			s.Push(neb)
		}
	}

	return res, nil
}
