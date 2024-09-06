package utils

func Map[T any](arr []T, f func(T) T) []T {
	ret := []T{}
	for _, v := range arr {
		ret = append(ret, f(v))
	}
	return ret
}

func Filter[T any](arr []T, f func(T) bool) []T {
	ret := []T{}
	for _, v := range arr {
		if f(v) {
			ret = append(ret, v)
		}
	}
	return ret
}

func Find[T any](arr []T, f func(T) bool) *T {
	for i := range arr {
		v := &arr[i]
		if f(*v) {
			return v
		}
	}
	return nil
}
