package csvx

import (
	"io"
	"iter"
)

func Rows[T any](r io.Reader, opts ...Options) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		d := NewDecoder(r, opts...)
		for {
			var row T
			err := d.Decode(&row)
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(row, err)
				return
			}
			if !yield(row, nil) {
				return
			}
		}
	}
}
