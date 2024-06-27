# slog/meld: A Logging Handler for log/slog

[![Go Reference](https://pkg.go.dev/badge/snqk.dev/slog/meld.svg)](https://pkg.go.dev/snqk.dev/slog/meld)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

`slog/meld` provides a simple `slog.Handler` designed to recursively merge and de-duplicate log attributes, ensuring
clean, concise, and informative log entries.

## Installation

```bash
go get -u snqk.dev/slog/meld
```

## Key Features

* **Attribute Merging:** Combines attributes with the same key, while preserving context.
* **Recursive Group Merging:**  Handles nested `slog.Group` attributes, ensuring proper merging at all levels.
* **De-duplication:** Eliminates duplicate keys within groups, preventing cluttered logs.
* **Order Preservation:** Maintains the original order of attributes as they were defined, even after merging or replacing.
* **Lightweight:** A pure go library without any dependencies.
* **Simplified Bootstrap:** Doesn't require any configuration options.

## Considerations

* **Thread Safety:** A `sync.RWMutex` ensures merge operations do not conflict with each other.
* **Greedy Merge:** Attributes are merged ahead-of-time vs when logging, where possible. IE when calling `Logger.With()`
  or `Logger.WithGroup()`.

## Usage

### Wrap `slog.Default()`

For most implementations, the following wrapper would suffice.

```go
package main

import (
	"log/slog"

	"snqk.dev/slog/meld"
)

func init() {
	slog.SetDefault(slog.New(meld.NewHandler(slog.Default().Handler())))
}
```

### Example with Comparisons

**Play:** https://go.dev/play/p/7r6D-reA4Pd

```go
package main

import (
	"log/slog"
	"os"

	"github.com/veqryn/slog-dedup"
	"snqk.dev/slog/meld"
)

func main() {
	hello(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	hello(slog.New(slogdedup.NewOverwriteHandler(slog.NewJSONHandler(os.Stdout, nil), nil)))
	hello(slog.New(meld.NewHandler(slog.NewJSONHandler(os.Stdout, nil))))
}

func hello(log *slog.Logger) {
	log = log.With(slog.Group("alice", slog.String("foo", "initial_attr"), slog.String("bar", "old_attr")))
	// do stuff...
	log = log.With(slog.Group("alice", slog.String("foo", "overwritten_attr"), slog.String("qux", "new_attr")))
	// do stuff...
	log = log.WithGroup("bob")
	// do stuff...
	log.Info("hello_world", "foo", "inline_attr")
}
```

#### Output using `slog/meld`

Newer attributes are overlay on top of the older tree, and merged together.

```json
{
  "time": "2009-11-10T23:00:00Z",
  "level": "INFO",
  "msg": "hello_world",
  "alice": {
    "foo": "overwritten_attr",
    "bar": "old_attr",
    "qux": "new_attr"
  },
  "bob": {
    "foo": "inline_attr"
  }
}
```

This approach means we only ever append &/ update older values, never deleting any attributes.
This makes `slog/meld` partially immutable, where the keys cannot be changed, but their values are mutable.

Alternatively, consider implementing `slog.LogValuer` for dynamically rendering attributes for different situations.

#### Output using `log/slog`

The fully immutable nature of `log/slog` means we can't modify previously added attributes (keys and values), only add more of the same as their duplicates.

```json
{
  "time": "2009-11-10T23:00:00Z",
  "level": "INFO",
  "msg": "hello_world",
  "alice": {
    "foo": "initial_attr",
    "bar": "old_attr"
  },
  "alice": {
    "foo": "overwritten_attr",
    "qux": "new_attr"
  },
  "bob": {
    "foo": "inline_attr"
  }
}
```

While this is legal JSON, parsers like `jq` will process it by overwriting earlier values with later ones.
Doing so results in the same output as from `veqryn/slog-dedup.NewOverwriteHandler`.

#### Output using `veqryn/slog-dedup.NewOverwriteHandler`

Here, `{"bar": "old_attr"}` is missing due to its parent group being completely overwritten, as deduplication alone isn't meant to handle attribute merging.

```json
{
  "time": "2009-11-10T23:00:00Z",
  "level": "INFO",
  "msg": "hello_world",
  "alice": {
    "foo": "overwritten_attr",
    "qux": "new_attr"
  },
  "bob": {
    "foo": "inline_attr"
  }
}
```

## Benchmarks

Benchmarks were also performed against the scenario above, comparing `log/slog`, `slog/meld`, and
`veqryn/slog-dedup.NewOverwriteHandler`.

```
$ go test -bench=. -benchtime 2s -benchmem -cpu 1,2,4,8,16 -run notest
goos: linux
goarch: amd64
pkg: snqk.dev/slog/meld
cpu: Intel(R) Core(TM) i7-5960X CPU @ 3.00GHz
BenchmarkDefaultLogger            721747              3671 ns/op            1272 B/op         25 allocs/op
BenchmarkDefaultLogger-2         1000000              2146 ns/op            1272 B/op         25 allocs/op
BenchmarkDefaultLogger-4         2117403              1132 ns/op            1272 B/op         25 allocs/op
BenchmarkDefaultLogger-8         2331649              1053 ns/op            1274 B/op         25 allocs/op
BenchmarkDefaultLogger-16        1939368              1335 ns/op            1278 B/op         25 allocs/op
BenchmarkMeldLogger               369272              6351 ns/op            2736 B/op         70 allocs/op
BenchmarkMeldLogger-2             658176              3626 ns/op            2736 B/op         70 allocs/op
BenchmarkMeldLogger-4            1216179              1962 ns/op            2737 B/op         70 allocs/op
BenchmarkMeldLogger-8            1000000              2017 ns/op            2741 B/op         70 allocs/op
BenchmarkMeldLogger-16            979236              2453 ns/op            2750 B/op         70 allocs/op
BenchmarkOverwriteLogger          131522             15580 ns/op           13099 B/op         65 allocs/op
BenchmarkOverwriteLogger-2        259272              9128 ns/op           14019 B/op         65 allocs/op
BenchmarkOverwriteLogger-4        245490              8277 ns/op           15768 B/op         65 allocs/op
BenchmarkOverwriteLogger-8        201349             10045 ns/op           20256 B/op         65 allocs/op
BenchmarkOverwriteLogger-16       199776             12011 ns/op           28338 B/op         65 allocs/op
PASS
ok      snqk.dev/slog/meld      42.994s
```
