package fastcsv

import (
	"bufio"
	"io"
)

// Reader parses RFC 4180-style CSV (comma-separated, optional quotes).
type Reader struct {
	br    *bufio.Reader
	comma byte
	rec   []string
	field []byte
}

// NewReader returns a CSV reader with comma ','.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		br:    bufio.NewReaderSize(r, 256*1024),
		comma: ',',
		rec:   make([]string, 0, 16),
		field: make([]byte, 0, 64),
	}
}

// Read returns the next record or io.EOF at end of input.
func (r *Reader) Read() ([]string, error) {
	r.rec = r.rec[:0]
	for {
		endRecord, err := r.readField()
		if err == io.EOF {
			if len(r.rec) == 0 {
				return nil, io.EOF
			}
			if len(r.field) > 0 || len(r.rec) > 0 {
				r.rec = append(r.rec, string(r.field))
			}
			return r.rec, nil
		}
		if err != nil {
			return nil, err
		}
		r.rec = append(r.rec, string(r.field))
		if endRecord {
			return r.rec, nil
		}
	}
}

func (r *Reader) readField() (endRecord bool, err error) {
	r.field = r.field[:0]
	quoted := false
	first := true

	for {
		c, e := r.br.ReadByte()
		if e == io.EOF {
			if first {
				return false, io.EOF
			}
			return true, io.EOF
		}
		if e != nil {
			return false, e
		}
		first = false

		if quoted {
			if c == '"' {
				peek, e2 := r.br.ReadByte()
				if e2 == io.EOF {
					return true, nil
				}
				if e2 != nil {
					return false, e2
				}
				if peek == '"' {
					r.field = append(r.field, '"')
					continue
				}
				_ = r.br.UnreadByte()
				quoted = false
				continue
			}
			r.field = append(r.field, c)
			continue
		}

		switch c {
		case '"':
			quoted = true
		case r.comma:
			return false, nil
		case '\r':
			peek, e2 := r.br.ReadByte()
			if e2 == io.EOF {
				return true, nil
			}
			if e2 != nil {
				return false, e2
			}
			if peek != '\n' {
				_ = r.br.UnreadByte()
			}
			return true, nil
		case '\n':
			return true, nil
		default:
			r.field = append(r.field, c)
		}
	}
}
