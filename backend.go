package csvx

type recordReader interface {
	Read() (record []string, err error)
}

type recordWriter interface {
	Write(record []string) error
	Flush() error
}
