package csvx

// Options configures marshal/unmarshal. Zero value means "unset"; joinOptions merges with last-write-wins per field.
type Options struct {
	noHeader                  *bool
	allowUnknownColumns       *bool
	matchCaseInsensitiveNames *bool
	allocEmptyPointers        *bool
	comma                     *rune
	comment                   *rune
	lazyQuotes                *bool
	trimLeadingSpace          *bool
	marshalers                *Marshalers
	unmarshalers              *Unmarshalers
}

type options struct {
	noHeader                  bool
	allowUnknownColumns       bool
	matchCaseInsensitiveNames bool
	allocEmptyPointers        bool
	comma                     rune // 0 = unset → backend default ','
	comment                   rune // 0 = unset → backend default 0
	lazyQuotes                bool
	trimLeadingSpace          bool
	marshalers                *Marshalers
	unmarshalers              *Unmarshalers
}

func boolOpt(v bool) *bool { return &v }
func runeOpt(v rune) *rune { return &v }

func NoHeader(v bool) Options                  { return Options{noHeader: boolOpt(v)} }
func AllowUnknownColumns(v bool) Options       { return Options{allowUnknownColumns: boolOpt(v)} }
func MatchCaseInsensitiveNames(v bool) Options { return Options{matchCaseInsensitiveNames: boolOpt(v)} }
func AllocEmptyPointers(v bool) Options        { return Options{allocEmptyPointers: boolOpt(v)} }
func Comma(r rune) Options                     { return Options{comma: runeOpt(r)} }
func Comment(r rune) Options                   { return Options{comment: runeOpt(r)} }
func LazyQuotes(v bool) Options                { return Options{lazyQuotes: boolOpt(v)} }
func TrimLeadingSpace(v bool) Options          { return Options{trimLeadingSpace: boolOpt(v)} }

func joinOptions(opts []Options) options {
	var out options
	for _, o := range opts {
		if o.noHeader != nil {
			out.noHeader = *o.noHeader
		}
		if o.allowUnknownColumns != nil {
			out.allowUnknownColumns = *o.allowUnknownColumns
		}
		if o.matchCaseInsensitiveNames != nil {
			out.matchCaseInsensitiveNames = *o.matchCaseInsensitiveNames
		}
		if o.allocEmptyPointers != nil {
			out.allocEmptyPointers = *o.allocEmptyPointers
		}
		if o.comma != nil {
			out.comma = *o.comma
		}
		if o.comment != nil {
			out.comment = *o.comment
		}
		if o.lazyQuotes != nil {
			out.lazyQuotes = *o.lazyQuotes
		}
		if o.trimLeadingSpace != nil {
			out.trimLeadingSpace = *o.trimLeadingSpace
		}
		if o.marshalers != nil {
			out.marshalers = o.marshalers
		}
		if o.unmarshalers != nil {
			out.unmarshalers = o.unmarshalers
		}
	}
	return out
}
