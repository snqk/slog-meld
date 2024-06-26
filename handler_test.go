package meld

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"strconv"
	"testing"
	"testing/slogtest"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	for i, test := range []struct {
		mock     func(log *slog.Logger)
		expected string
	}{
		{
			expected: "level=INFO msg=hello_world\n",
			mock: func(log *slog.Logger) {
				log.Info("hello_world")
			},
		},
		{
			expected: "level=INFO msg=hello_world foo=bar\n",
			mock: func(log *slog.Logger) {
				log.With("foo", "bar").Info("hello_world")
			},
		},
		{
			expected: "level=INFO msg=hello_world foo=baz\n",
			mock: func(log *slog.Logger) {
				log1 := log.With("foo", "bar")
				log2 := log1.With("foo", "baz")
				log2.Info("hello_world")
			},
		},
		{
			expected: "level=INFO msg=hello_world alice.foo=boo alice.bar=baz alice.qux=quux bob=lorem_ipsum\n",
			mock: func(log *slog.Logger) {
				log1 := log.With(slog.Group("alice", slog.String("foo", "goo"), slog.String("bar", "baz")))
				log2 := log1.With(slog.Group("alice", slog.String("foo", "boo"), slog.String("qux", "quux")))
				log3 := log2.With("bob", "lorem_ipsum")
				log3.Info("hello_world")
			},
		},
		{
			expected: "level=INFO msg=hello_world alice.foo=boo alice.qux=quux bob=lorem_ipsum\n",
			mock: func(log *slog.Logger) {
				log1 := log.With(slog.String("alice", "snafu"))
				log2 := log1.With(slog.Group("alice", slog.String("foo", "boo"), slog.String("qux", "quux")))
				log3 := log2.With("bob", "lorem_ipsum")
				log3.Info("hello_world")
			},
		},
		{
			expected: "level=INFO msg=hello_world alice.foo=boo alice.bar=baz alice.qux=quux bob=lorem_ipsum\n",
			mock: func(log *slog.Logger) {
				log1 := log.With(slog.Group("alice", slog.String("foo", "goo"), slog.String("bar", "baz")))
				log2 := log1.With(slog.Group("alice", slog.String("foo", "boo"), slog.String("qux", "quux")))
				log2.Info("hello_world", "bob", "lorem_ipsum")
			},
		},
		{
			expected: "level=INFO msg=hello_world alice.foo=boo alice.bar=baz alice.qux=quux bob.foo=lorem_ipsum\n",
			mock: func(log *slog.Logger) {
				log1 := log.With(slog.Group("alice", slog.String("foo", "goo"), slog.String("bar", "baz")))
				log2 := log1.With(slog.Group("alice", slog.String("foo", "boo"), slog.String("qux", "quux")))
				log3 := log2.WithGroup("bob")
				log3.Info("hello_world", "foo", "lorem_ipsum")
			},
		},
		{
			expected: "level=INFO msg=hello_world alice=newtext bob.foo=lorem_ipsum\n",
			mock: func(log *slog.Logger) {
				log1 := log.With(slog.Group("alice", slog.String("foo", "goo"), slog.String("bar", "baz")))
				log2 := log1.With(slog.Group("alice", slog.String("foo", "boo"), slog.String("qux", "quux")))
				log2a := log2.With("alice", "newtext")
				log3 := log2a.WithGroup("bob")
				log3.Info("hello_world", "foo", "lorem_ipsum")
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)
			test.mock(slog.New(NewHandler(slog.NewTextHandler(buf, &slog.HandlerOptions{
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey && len(groups) == 0 {
						return slog.Attr{}
					}
					return a
				},
			}))))

			assert.Equal(t, test.expected, buf.String())
		})
	}
}

func Test_slogtest_json(t *testing.T) {
	var buf bytes.Buffer

	newHandler := func(*testing.T) slog.Handler {
		buf.Reset()
		return NewHandler(slog.NewJSONHandler(&buf, nil))
	}
	result := func(t *testing.T) map[string]any {
		m := map[string]any{}
		if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
			t.Fatal(err)
		}
		return m
	}

	slogtest.Run(t, newHandler, result)
}

func BenchmarkDefaultLogger(b *testing.B) {
	b.ReportAllocs()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.With(slog.Group("alice", slog.String("foo", "goo"), slog.String("bar", "baz"))).
				With(slog.Group("alice", slog.String("foo", "boo"), slog.String("qux", "quux"))).
				WithGroup("bob").
				Info("hello_world", "foo", "lorem_ipsum")
		}
	})
}

func BenchmarkMeldLogger(b *testing.B) {
	b.ReportAllocs()

	log := slog.New(NewHandler(slog.NewTextHandler(io.Discard, nil)))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.With(slog.Group("alice", slog.String("foo", "goo"), slog.String("bar", "baz"))).
				With(slog.Group("alice", slog.String("foo", "boo"), slog.String("qux", "quux"))).
				WithGroup("bob").
				Info("hello_world", "foo", "lorem_ipsum")
		}
	})
}
