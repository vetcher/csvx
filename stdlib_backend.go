package csvx

import (
	"encoding/csv"
	"errors"
	"io"
)

type stdlibReader struct{ r *csv.Reader }

func newStdlibReader(r io.Reader, o options) recordReader {
	cr := csv.NewReader(r)
	if o.comma != 0 {
		cr.Comma = o.comma
	}
	if o.comment != 0 {
		cr.Comment = o.comment
	}
	cr.LazyQuotes = o.lazyQuotes
	cr.TrimLeadingSpace = o.trimLeadingSpace
	cr.ReuseRecord = true
	return &stdlibReader{r: cr}
}

func (s *stdlibReader) Read() ([]string, error) {
	rec, err := s.r.Read()
	if err != nil {
		var pe *csv.ParseError
		if errors.As(err, &pe) {
			return nil, &SyntaxError{Msg: err.Error()}
		}
		return nil, err
	}
	return rec, nil
}

type stdlibWriter struct{ w *csv.Writer }

func newStdlibWriter(w io.Writer, o options) recordWriter {
	cw := csv.NewWriter(w)
	if o.comma != 0 {
		cw.Comma = o.comma
	}
	return &stdlibWriter{w: cw}
}

func (s *stdlibWriter) Write(record []string) error { return s.w.Write(record) }
func (s *stdlibWriter) Flush() error {
	s.w.Flush()
	return s.w.Error()
}
