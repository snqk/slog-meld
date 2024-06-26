# slog/meld: A Logging Handler for log/slog

[![Go Reference](https://pkg.go.dev/badge/snqk.dev/slog/meld.svg)](https://pkg.go.dev/snqk.dev/slog/meld)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

`slog/meld` provides a simple `slog.Handler` designed to recursively merge and de-duplicate log attributes, ensuring clean, concise, and informative log entries.

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
* **Greedy Merge:** Attributes are merged ahead-of-time vs when logging, where possible. IE when calling `Logger.With()` or `Logger.WithGroup()`.
## Usage
### Wrap `slog.Default()`
For most implementations, the following wrapper would suffice.
```go
slog.SetDefault(meld.NewHandler(slog.Default()))
```

### Custom Logger
For custom loggers, the bootstrap is similar.
```go
log := slog.New(meld.NewHandler(slog.NewJSONHandler(os.Stdout, nil)))
// do stuff...
log1 := log.With(slog.Group("alice", slog.String("foo", "initial_attr"), slog.String("bar", "old_attr")))
// do stuff...
log2 := log1.With(slog.Group("alice", slog.String("foo", "overwritten_attr"), slog.String("qux", "new_attr")))
// do stuff...
log3 := log2.WithGroup("bob")
// do stuff...
log3.Info("hello_world", "foo", "inline_attr")
```

```json
{
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
