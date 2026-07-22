# csvx

Fast, struct-tag-driven CSV encoding and decoding for Go (`encoding/json`-shaped API with explicit options).

## Install

```bash
go get github.com/vetcher/csvx
```

## Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/vetcher/csvx"
)

type Client struct {
	ID   string `csv:"client_id"`
	Name string `csv:"client_name"`
	Age  int    `csv:"client_age"`
}

func main() {
	in := []Client{
		{ID: "1", Name: "Jose", Age: 42},
		{ID: "2", Name: "Daniel", Age: 26},
	}

	data, err := csvx.Marshal(in)
	if err != nil {
		log.Fatal(err)
	}

	var out []Client
	if err := csvx.Unmarshal(data, &out); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", out)
}
```

## Benchmarks

See [bench/README.md](bench/README.md) for comparison results.

**CSV backend:** `encoding/csv` (stdlib). An experimental parser in `internal/fastcsv` did not pass the parse benchmark gate vs `encoding/csv`, so there is no public `csvtext` package.
