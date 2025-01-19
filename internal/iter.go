package internal

import "iter"

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
