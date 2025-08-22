package mux

import "iter"

func FlagsOfType[T any](route Route) iter.Seq[T] {
	return func(yield func(T) bool) {
		for flag := range route.Flags() {
			if val, ok := flag.(T); ok && !yield(val) {
				break
			}
		}
	}
}
