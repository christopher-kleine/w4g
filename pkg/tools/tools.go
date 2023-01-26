package tools

import "golang.org/x/exp/constraints"

func Min[T constraints.Integer](a, b T) T {
	if a <= b {
		return a
	}

	return b
}

func Max[T constraints.Integer](a, b T) T {
	if a >= b {
		return a
	}

	return b
}

func Ternary[T any](eq bool, a, b T) T {
	if eq {
		return a
	}

	return b
}
