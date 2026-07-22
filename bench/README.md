# csvx benchmarks

Comparison suite for `encoding/csv` and [csvx](../) (this repo).

## Fixture

- ~10,000 data rows × 10 string columns (plus header)
- Struct tags: `csv:"col_N"`

## Run locally

```bash
cd bench
go mod tidy
go test -bench=. -benchmem -count=1
```

## Published results

| Environment | Value |
|-------------|-------|
| Go | go1.24.4 linux/amd64 |
| OS | Linux 6.12.94+ x86_64 |
| CPU | Intel(R) Xeon(R) Processor |
| Date | 2026-07-22 |

| Benchmark | ns/op | MB/s | B/op | allocs/op |
|-----------|------:|-----:|-----:|----------:|
| `BenchmarkEncodingCSV_Parse` | 2,231,071 | 263.98 | 2,227,654 | 20,019 |
| `BenchmarkCSVx_UnmarshalStruct` | 8,646,525 | 68.11 | 12,582,805 | 40,051 |
| `BenchmarkCSVx_MarshalStruct` | 3,439,301 | — | 3,698,158 | 10,014 |
| `BenchmarkCSVx_DecodeAllocs` | 4,298,930 | 137.00 | 2,229,242 | 20,032 |

Results from `go test -bench=. -benchmem -count=1` in this directory. Re-run on your machine to refresh; numbers vary by hardware.

Field converters are classified once in the type plan (builtin / `MarshalCSV` / `TextMarshaler` / slow). Builtin fields skip per-cell interface probes — Decode allocs are near raw `encoding/csv` parse (~2/row).

## Task 12: `csvtext` gate (parse-only)

`internal/fastcsv` was compared to `encoding/csv` on the same ~10k×10 fixture (`go test ./internal/fastcsv -bench=Parse -benchmem -count=5`, go1.24 linux/amd64):

| Benchmark | ns/op | MB/s | B/op | allocs/op |
|-----------|------:|-----:|-----:|----------:|
| `BenchmarkParse_EncodingCSV` | ~1,459,000 | ~404 | 627,640 | 10,019 |
| `BenchmarkParse_FastCSV` | ~2,739,000 | ~215 | 781,730 | 100,005 |

**Gate: LOSE** — custom parser is ~1.9× slower and allocates ~10× more. Public `csvtext` was **not** exported; csvx stays on the stdlib backend.
