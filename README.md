# slog/meld: A Logging Handler for log/slog

[![Go Reference](https://pkg.go.dev/badge/snqk.dev/slog/meld.svg)](https://pkg.go.dev/snqk.dev/slog/meld)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

`slog/meld` is a slog handler designed to enhance your structured logging experience in Go. It recursively merges and de-duplicates log attributes, ensuring clean, concise, and informative log entries.

## Installation

```bash
go get snqk.dev/slog/meld
```

## Key Features

* **Zero Dependency:** A pure go library without any dependencies.
* **Simplified Bootstrap:** Doesn't require any configuration options.
* **Attribute Merging:** Combines attributes with the same key, while preserving context.
* **Recursive Group Merging:**  Handles nested `slog.Group` attributes, ensuring proper merging at all levels.
* **De-duplication:** Eliminates duplicate keys within groups, preventing cluttered logs.
* **Order Preservation:** Maintains the original order of attributes, even after merging or replacing.

## Considerations

* **Thread Safety:** A `sync.RWMutex` ensures merge operations do not conflict with each other. 
* **Greedy Merge:** Attributes are merged ahead-of-time instead of at the time of logging, where possible. IE when calling `Logger.With()` or `Logger.WithGroup()`.
* **Performance:** Handler is not necessarily optimised for performance.

## Usage
### Wrap `slog.Default()`
For most implementations, the following wrapper would suffice.
```go
slog.SetDefault(meld.NewHandler(slog.Default()))
```

### Custom Logger
For custom loggers, 
```go
log := slog.New(meld.NewHandler(slog.NewTextHandler(os.Stdout, nil)))
// do stuff...
log1 := log.With(slog.Group("alice", slog.String("foo", "goo"), slog.String("bar", "baz")))
// do stuff...
log2 := log1.With(slog.Group("alice", slog.String("foo", "boo"), slog.String("qux", "quux")))
// do stuff...
log3 := log2.WithGroup("bob")
// do stuff...
log3.Info("hello_world", "foo", "lorem_ipsum")
```

```console
level=INFO msg=hello_world alice.foo=boo alice.bar=baz alice.qux=quux bob.foo=lorem_ipsum
```