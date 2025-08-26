package mux

import "iter"

// FlagsOfType allows you to get all flags of a certain type[T] from [Route].
func FlagsOfType[T any](route Route) iter.Seq[T] {
	return func(yield func(T) bool) {
		for flag := range route.Flags() {
			if val, ok := flag.(T); ok && !yield(val) {
				break
			}
		}
	}
}
