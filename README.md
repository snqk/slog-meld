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
* **Order Preservation:** Maintains the original order of attributes, even after merging or replacing.
* **Zero Dependency:** A pure go library without any dependencies.
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

### Example
**Play:** https://play.golang.com/p/Tkj5CBiKnZt

```go
package main

import (
	"log/slog"
	"os"

	"snqk.dev/slog/meld"
)

func main() {
	log := slog.New(meld.NewHandler(slog.NewJSONHandler(os.Stdout, nil)))
	// do stuff...
	log1 := log.With(slog.Group("alice", slog.String("foo", "initial_attr"), slog.String("bar", "old_attr")))
	// do stuff...
	log2 := log1.With(slog.Group("alice", slog.String("foo", "overwritten_attr"), slog.String("qux", "new_attr")))
	// do stuff...
	log3 := log2.WithGroup("bob")
	// do stuff...
	log3.Info("hello_world", "foo", "inline_attr")
}
```

Output (Formatted):
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
**NOTE:** merging attributes means we're only ever appending &/ updating older values. We never delete older attributes. For conditionally removable attributes, implement `slog.LogValuer`.

## Benchmarks
Benchmarks were also performed against the scenario above, comparing `log/slog`, `slog/meld`, and `veqryn/slog-dedup.NewOverwriteHandler`.
```
$ go test -bench=. -benchtime 2s -benchmem -cpu 1,2,4 -run notest
goos: linux
goarch: amd64
BenchmarkDefaultLogger       	  481646	      4877 ns/op	    1272 B/op	      25 allocs/op
BenchmarkDefaultLogger-2     	  715798	      2876 ns/op	    1272 B/op	      25 allocs/op
BenchmarkDefaultLogger-4     	 1753418	      1432 ns/op	    1272 B/op	      25 allocs/op
BenchmarkMeldLogger          	  254832	      8926 ns/op	    3216 B/op	      73 allocs/op
BenchmarkMeldLogger-2        	  432796	      5033 ns/op	    3216 B/op	      73 allocs/op
BenchmarkMeldLogger-4        	  983511	      2457 ns/op	    3217 B/op	      73 allocs/op
BenchmarkOverwriteLogger     	  111040	     19365 ns/op	   13118 B/op	      65 allocs/op
BenchmarkOverwriteLogger-2   	  222187	     10740 ns/op	   14016 B/op	      65 allocs/op
BenchmarkOverwriteLogger-4   	  230452	     18207 ns/op	   15753 B/op	      65 allocs/op
PASS
ok  	snqk.dev/slog/meld	28.032s
```
