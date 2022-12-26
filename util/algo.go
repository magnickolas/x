package util

import "golang.org/x/exp/constraints"

func Min[T constraints.Ordered](args ...T) T {
	return FoldLeft1(Min2[T], args)
}

func Max[T constraints.Ordered](args ...T) T {
	return FoldLeft1(Max2[T], args)
}

func Min2[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max2[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func FoldLeft[T, U any](f func(U, T) U, init U, xs []T) U {
	for _, x := range xs {
		init = f(init, x)
	}
	return init
}

func FoldLeft1[T any](f func(T, T) T, xs []T) T {
	return FoldLeft(f, xs[0], xs[1:])
}

func Map[T, U any](f func(T) U, xs []T) []U {
	ys := make([]U, len(xs))
	for i, x := range xs {
		ys[i] = f(x)
	}
	return ys
}

func Len(x string) int {
	return len(x)
}

func Keys[T comparable, U any](m map[T]U) []T {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func MapFromKV[T comparable, U any](keys []T, values []U) map[T]U {
	m := make(map[T]U, len(keys))
	for i, k := range keys {
		m[k] = values[i]
	}
	return m
}
