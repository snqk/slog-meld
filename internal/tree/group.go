package tree

import (
	"log/slog"
)

type Group struct {
	vs []*value
}

func (g *Group) LogValue() slog.Value {
	return slog.GroupValue(g.Render()...)
}

func (g *Group) Clone() *Group {
	n := &Group{vs: make([]*value, len(g.vs))}

	for i, v := range g.vs {
		n.vs[i] = v.clone()
	}

	return n
}

func (g *Group) Render() []slog.Attr {
	out := make([]slog.Attr, len(g.vs))

	for i, v := range g.vs {
		out[i] = slog.Attr{Key: v.name, Value: v.LogValue()}
	}

	return out
}

func (g *Group) Merge(stack []string, attrs ...slog.Attr) {
	for i := range attrs {
		merge(g, stack, attrs[i])
	}
}

// merge applies attr on stack to g, iterating recursively and merging where necessary.
func merge(in *Group, stack []string, attr slog.Attr) {
	l := last(in, stack)

	for i := range l.vs {
		if l.vs[i].name == attr.Key {
			if attr.Value.Kind() != slog.KindGroup {
				l.vs[i].v = attr.Value
				return
			}

			if l.vs[i].g == nil {
				l.vs[i].g = new(Group)
			}

			l.vs[i].g.Merge(nil, attr.Value.Group()...)
			return
		}
	}

	if attr.Value.Kind() != slog.KindGroup {
		l.vs = append(l.vs, &value{name: attr.Key, v: attr.Value})
	} else {
		ng := new(Group)
		ng.Merge(nil, attr.Value.Group()...)
		l.vs = append(l.vs, &value{name: attr.Key, g: ng})
	}

	return
}

func last(in *Group, stack []string) *Group {
	if len(stack) == 0 {
		return in
	}

	for _, v := range in.vs {
		if v.name == stack[0] {
			return last(v.g, stack[1:])
		}
	}

	panic("slog/meld: invalid or misconfigured stack")
}
