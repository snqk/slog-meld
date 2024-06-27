package tree

import (
	"log/slog"
	"sync"
)

type Group struct {
	m  sync.RWMutex
	vs []*value
}

func (g *Group) LogValue() slog.Value {
	return slog.GroupValue(g.Render()...)
}

func (g *Group) Clone() *Group {
	n := &Group{vs: make([]*value, len(g.vs))}

	g.m.RLock()
	defer g.m.RUnlock()

	for i, v := range g.vs {
		n.vs[i] = v.clone()
	}

	return n
}

func (g *Group) Render() []slog.Attr {
	out := make([]slog.Attr, len(g.vs))

	g.m.RLock()
	defer g.m.RUnlock()

	for i, v := range g.vs {
		out[i] = slog.Attr{Key: v.name, Value: v.LogValue()}
	}

	return out
}

func (g *Group) Merge(stack []string, attrs ...slog.Attr) {
	g.m.Lock()
	defer g.m.Unlock()

	for i := range attrs {
		merge(g, stack, attrs[i])
	}
}

// merge applies attr on stack to g, iterating recursively and merging where necessary.
func merge(in *Group, stack []string, attr slog.Attr) {
	l := last(in, stack)

	var match *int
	for i := range l.vs {
		if l.vs[i].name == attr.Key {
			match = &i
		}
	}

	if match == nil {
		if attr.Value.Kind() != slog.KindGroup {
			l.vs = append(l.vs, &value{name: attr.Key, v: attr.Value})
		} else {
			ng := &Group{vs: make([]*value, 0)}
			ng.Merge(nil, attr.Value.Group()...)
			l.vs = append(l.vs, &value{name: attr.Key, g: ng})
		}
		return
	}

	if attr.Value.Kind() != slog.KindGroup {
		l.vs[*match] = &value{name: attr.Key, v: attr.Value}
		return
	}

	if l.vs[*match].g == nil {
		ng := &Group{vs: make([]*value, 0)}
		l.vs[*match] = &value{name: attr.Key, g: ng} // override
	}

	l.vs[*match].g.Merge(nil, attr.Value.Group()...)
}

func last(in *Group, stack []string) *Group {
	switch len(stack) {
	case 0:
		return in
	default:
		for _, v := range in.vs {
			if v.name == stack[0] {
				return last(v.g, stack[1:])
			}
		}
		panic("slog/meld: invalid or misconfigured stack")
	}
}
