package util

import "golang.org/x/exp/constraints"

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// return type U, which equals T if not specified
func Max[T constraints.Ordered](a, b T) T {
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

func IdErr[T any](x T) (T, error) {
	return x, nil
}
