package iterutil

import "iter"

func IterSeq[T any](next func() (T, bool)) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			v, ok := next()
			if !ok {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}
