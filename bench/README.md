# csvx benchmarks

Comparison suite for `encoding/csv`, [csvx](../) (this repo), and [gocarina/gocsv](https://github.com/gocarina/gocsv).

## Fixture

- ~10,000 data rows × 10 string columns (plus header)
- Struct tags: `csv:"col_N"`

## Run locally

```bash
cd bench
go mod tidy
go test -bench=. -benchmem -count=1
```

If `go mod tidy` cannot resolve `gocsv` via the public module proxy, fetch it directly:

```bash
GOPROXY=direct go mod tidy
```

Without `gocsv`, build with the `nogocsv` tag (csvx and `encoding/csv` benches still run):

```bash
go test -tags nogocsv -bench=. -benchmem -count=1
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
| `BenchmarkEncodingCSV_Parse` | 2,233,872 | 263.65 | 2,227,696 | 20,020 |
| `BenchmarkCSVx_UnmarshalStruct` | 18,251,311 | 32.27 | 15,782,958 | 240,056 |
| `BenchmarkCSVx_MarshalStruct` | 12,761,617 | — | 6,898,109 | 210,018 |
| `BenchmarkCSVx_DecodeAllocs` | 14,060,610 | 41.89 | 5,429,359 | 220,035 |
| `BenchmarkGoCSV_UnmarshalStruct` | 8,297,291 | 70.98 | 9,611,754 | 140,059 |

Results from `go test -bench=. -benchmem -count=1` in this directory. Re-run on your machine to refresh; numbers vary by hardware.
