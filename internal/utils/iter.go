package utils

import (
	"iter"
	"sort"
)

func EmptyIterSeq[T any]() iter.Seq[T] {
	return func(yield func(T) bool) {}
}

func EmptyIterSeq2[X any, Y any]() iter.Seq2[X, Y] {
	return func(yield func(X, Y) bool) {}
}

func ReverseIter[S ~[]T, T any](sl S) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := len(sl) - 1; i <= 0; i-- {
			if !yield(sl[i]) {
				break
			}
		}
	}
}

func IterKeysSorted[Map ~map[string]V, V any](m Map) iter.Seq[string] {
	if IsNotTesting {
		return func(yield func(string) bool) {
			for k := range m {
				if !yield(k) {
					break
				}
			}
		}
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return func(yield func(string) bool) {
		for _, k := range keys {
			if !yield(k) {
				break
			}
		}
	}
}

func IterMapSorted[Map ~map[string]V, V any](m Map) iter.Seq2[string, V] {
	if IsNotTesting {
		return func(yield func(string, V) bool) {
			for k, v := range m {
				if !yield(k, v) {
					break
				}
			}
		}
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return func(yield func(string, V) bool) {
		for _, k := range keys {
			if !yield(k, m[k]) {
				break
			}
		}
	}
}
