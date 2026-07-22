# csvx Design Spec

**Date:** 2026-07-22  
**Module:** `github.com/vetcher/csvx`  
**Status:** Draft — pending user review before implementation plan

## Goal

Build a high-performance Go CSV library with an API shaped like `encoding/json` / `encoding/json/v2`: semantic `Marshal`/`Unmarshal` plus Options, and an optional low-level `csvtext` layer. Prefer speed without sacrificing a clear, test-first public API.

## Constraints (decisions)

| Decision | Choice |
|----------|--------|
| Architecture approach | Semantic-first; public `csvtext` only if it beats/ties `encoding/csv` |
| Go version | 1.24+ / tip |
| Module path | `github.com/vetcher/csvx` |
| Package layout | Root semantic package + `csvtext` subpackage (gated) |
| Struct tags | `` `csv` `` |
| Speed bar | Beat `encoding/csv` (parse) and popular binders (struct); else drop public `csvtext` |
| Streaming | `Decoder`/`Encoder` + `iter.Seq` / `iter.Seq2` helpers; no channel APIs in v1 |
| Headers | Required by default; `NoHeader` supported (positional / index tags) |
| Unknown columns | Reject by default; `AllowUnknownColumns` opt-in (also under `NoHeader` for extra cells) |
| Interfaces / `any` | Disallowed as field or unmarshal targets |
| Stdlib special types | Not built in (`time.Time`, etc. only via hooks) |
| Process | Test-driven development for all production code |

## Architecture

```
csvx/                 # semantic: Marshal*, Unmarshal*, Options, tags, Encoder/Decoder
csvx/csvtext/         # PUBLIC only if benches ≥ encoding/csv; else absent
csvx/internal/reader/ # private backend iface (stdlib now; custom later)
bench/                # published suite; README links results
```

### Layers

1. **Backend** — thin interface: read/write records + dialect. Initial implementation wraps `encoding/csv`. A custom tokenizer may replace it behind the same interface.
2. **Semantic** — map records ↔ Go values (structs, slices, maps). Variadic Options (json/v2-style).
3. **Streaming** — `Decoder.Decode`, `Encoder.Encode`, iterator helpers.

### csvtext export gate

- CI/bench suite compares parse-only performance to `encoding/csv` on agreed RFC 4180 fixtures.
- **Win or tie (within noise):** export public `csvtext` and document it.
- **Lose:** keep custom code private or delete it; root package continues on stdlib backend; do **not** publish `csvtext`.
- README always links benchmark results regardless of gate outcome.

## Public API

### Root package `csvx`

```go
Marshal(v any, opts ...Options) ([]byte, error)
Unmarshal(data []byte, v any, opts ...Options) error
MarshalWrite(w io.Writer, v any, opts ...Options) error
UnmarshalRead(r io.Reader, v any, opts ...Options) error

NewEncoder(w io.Writer, opts ...Options) *Encoder
NewDecoder(r io.Reader, opts ...Options) *Decoder
// Encoder.Encode(v any) error
// Decoder.Decode(v any) error

Rows[T any](r io.Reader, opts ...Options) iter.Seq2[T, error]
```

Notes:

- Top-level `any` in signatures means “any supported concrete value,” not interface-typed CSV fields.
- `Decoder.Decode` target rules:
  - `*T` where `T` is struct or `map[string]string`: decode **one** data row; return `io.EOF` when none left.
  - `*[]T` / `*[]*T` / `*[]map[string]string`: decode **all remaining** rows into the slice (appends; does not clear existing elements unless documented helper says otherwise — v1: **overwrite** by setting `len=0` then append).
  - Mixing “one row” and “all rows” in one Decoder is supported sequentially only for one-row targets; after an all-rows Decode, next Decode returns `io.EOF`.
- Prefer `Unmarshal*` for full-document slice loads; use `Decode` for streaming one-row targets and `Rows`.
- No process-wide `SetCSVReader` / `SetCSVWriter` mutators. Configure via Options per call or per Encoder/Decoder.

### Supported value shapes

| Shape | Mode |
|-------|------|
| `[]T`, `[]*T` | Full table (header by default) |
| `T`, `*T` | One row (`Decode` / single-record marshal) |
| `map[string]string`, `[]map[string]string` | Header mode |
| `[][]string` | `NoHeader` |
| Structs with `` `csv:"N"` `` | `NoHeader` positional |

### Explicitly unsupported

- Interface-typed fields and `any` / `interface{}` fields (plan-time `SemanticError`).
- Unmarshal into an interface value.
- `complex64` / `complex128` in v1 (use a user marshaler if needed later).
- Field-level slices/arrays/maps as multi-cell records by default (single cell only via marshaler/`format`).
- Channel-based APIs in v1.

## Builtin types

No Register and no Text\* required:

| Go type | Behavior |
|---------|----------|
| `string` | As-is |
| `bool` | `true`/`false`; unmarshal via `strconv.ParseBool` (also `1`/`0`) |
| `int`, `int8`…`int64` | `strconv` Format/Parse with bit size |
| `uint`, `uint8`…`uint64` | Same |
| `float32`, `float64` | `FormatFloat`/`ParseFloat` |
| `byte` / `rune` | Numeric (not character literals) |
| `[]byte`, `[N]byte` | Treated as UTF-8 string bytes; array length must match on unmarshal |
| `*T` | Dereference on marshal; allocate on unmarshal as needed |
| Named types over builtins | Underlying builtin unless hooks/interfaces apply |
| Structs | Exported fields; tags control names / skip / inline |

**Not builtins:** any stdlib composite special case (`time.Time`, `net.IP`, `url.URL`, …). Those work only through type methods, local Options hooks, or global registry.

**Empty cells:** unmarshal to zero value for builtins. For pointer fields, default is: empty cell → set pointer to `nil`. Option `AllocEmptyPointers(true)` changes this to allocate and write the zero value. (Replaces earlier `NilOnEmpty` draft name.)

## Tags

Struct tag name: `csv`.

Baseline tag features (full option laundry list deferred to a dedicated options review with the author):

- Name override: `` `csv:"column_name"` ``
- Ignore: `` `csv:"-"` ``
- `omitzero`, `omitempty`
- `format:...` for specialized formatting when applicable
- Inline nested struct: `` `csv:"."` `` (no parent prefix)
- Under `NoHeader`: index tags `` `csv:"0"` `` etc., or declaration order

## Marshalers and unmarshalers

### Interfaces (concrete types only)

```go
type Marshaler interface {
    MarshalCSV() (string, error)
}

type Unmarshaler interface {
    UnmarshalCSV(string) error
}
```

Also honor `encoding.TextMarshaler` / `encoding.TextUnmarshaler` in the fallback chain (generic mechanism, not a stdlib type bundle).

### Func hooks (types the user cannot edit)

```go
MarshalFunc[T any](fn func(T) (string, error)) *Marshalers
UnmarshalFunc[T any](fn func(string) (T, error)) *Unmarshalers
// Join to compose sets
```

### Attachment points

1. **Local Option** — `WithMarshalers` / `WithUnmarshalers` on the call or Encoder/Decoder.
2. **Global index** — process-wide registry:

```go
func RegisterMarshaler[T any](fn func(T) (string, error))
func RegisterUnmarshaler[T any](fn func(string) (T, error))
```

Libraries should prefer local Options to avoid hidden globals. Tests may need unregister/clear helpers.

No default global registrations ship with the library.

### Lookup order (first match wins)

**Marshal (value → cell string):**

1. Local `WithMarshalers` for the concrete dynamic type (including documented pointer/value matching rules)
2. `Marshaler` (`MarshalCSV`); if only pointer receiver and value is addressable, use `*T`
3. Global registry for that type
4. `encoding.TextMarshaler` (same addressability rules)
5. Builtin defaults / `format:` tag

**Unmarshal (cell string → value):**

1. Local `WithUnmarshalers`
2. `Unmarshaler` (`UnmarshalCSV`)
3. Global registry
4. `encoding.TextUnmarshaler`
5. Builtin defaults / `format:` tag

Precedence summary: **local Options > type methods > global registry > Text\* > builtins**.

## Options (starter set)

Opaque, composable, variadic. No process-wide dialect globals.

**Dialect (syntactic):** `Comma`, `Comment`, `LazyQuotes`, `TrimLeadingSpace`, record reuse / newline behavior as needed for backend parity with `encoding/csv`.

**Semantic:** `NoHeader`, `AllowUnknownColumns`, `MatchCaseInsensitiveNames`, omit helpers, `WithMarshalers`, `WithUnmarshalers`, `AllocEmptyPointers`.

**Keep:** per-type converters via marshalers; nested/inline `.`; ignore `-`.

**Drop:** global `SetCSVReader` / `SetCSVWriter` style configuration.

**Deferred:** author must approve the **complete** Options list (json/v2 analogues plus CSV dialect/semantic knobs) before implementation plan freezes every name. Starter set above is normative until that review adds/renames entries.

## Data flow

### Unmarshal

1. Open backend reader with dialect from Options.
2. Header mode: read first record as names; build name→field index (case-sensitive by default).
3. Each data record: map cells → fields; convert; apply unmarshal lookup order.
4. Unknown column/header → error unless `AllowUnknownColumns`.
5. Missing expected column → error (unless an approved omit/optional policy from the options pass).
6. `NoHeader`: skip header; bind by index tag or declaration order; trailing extra cells are unknown.

### Marshal

1. Build column plan from `reflect.Type` + Options fingerprint (cached).
2. Write header unless `NoHeader`.
3. Each element → cells → backend write.

## Errors

Typed, wrap-friendly:

- `SyntaxError` — parse/quote issues from backend
- `SemanticError` — bad target kind, disallowed interface/`any`, type conversion, unknown/missing columns
- `FieldError` — includes row/column name or index context (e.g. `row 3 column "age"`)

`Decoder.Decode` returns `io.EOF` when no more rows (json-like). Mid-stream failure: **fail-fast, leave partial slice as-is, return error** (caller discards if needed).

## Performance

- Cache type plans (`sync.Map` or equivalent) keyed by type + option fingerprint.
- Hot path: `strconv`, no `fmt`.
- Reuse buffers in Encoder/Decoder.
- Aim for low allocs on warm `Decode` into `*T`.
- Custom tokenizer work is justified only by the export gate.

## Testing (TDD mandatory)

All production code follows red → green → refactor:

1. Write one failing test for one behavior.
2. Run it; confirm failure is due to missing feature (not typo).
3. Write minimal code to pass.
4. Refactor while staying green.
5. Repeat.

No production code without a failing test first (see project TDD skill).

### Test map

- Dialect/parse edge cases (quotes, embedded delim, multiline, CRLF)
- Semantic: tags, `NoHeader`, unknown reject/allow, omit rules, inline
- Marshaler lookup order (table-driven)
- Reject `any`/interfaces at plan time
- Fuzz after core behaviors exist
- Race detector on type-plan cache; document Encoder/Decoder as not concurrency-safe per instance

### Benchmarks

Suite under `bench/`:

1. Parse-only vs `encoding/csv`
2. Struct unmarshal vs other popular binders if useful later (stdlib parse-only remains primary peer)
3. Struct marshal vs same
4. Streaming `Decode` with `ReportAllocs`

Publish results in `bench/README.md` with Go version and machine notes. Root `README.md` links to that document. CI runs benches; regression baselines as appropriate; `csvtext` export gated on parse-only results.

## Out of scope (v1)

- Channel unmarshal/marshal APIs
- Built-in `time.Time` / other stdlib special codecs
- Interface/`any` field support
- Guaranteeing public `csvtext` if it cannot match `encoding/csv`
- Global CSV reader/writer setters

## Implementation approach reminder

Ship the semantic API on an `encoding/csv` backend first (TDD). Prototype a faster backend behind `internal/`. Promote to public `csvtext` only after the benchmark gate passes.

## Open follow-ups (before or during planning)

1. Author review of **full Options list** (json/v2 analogues + CSV dialect/semantic knobs) — required before plan freezes names.
2. Whether a `csvx/register` side-effect subpackage is ever desired (not required for v1; default: no).
