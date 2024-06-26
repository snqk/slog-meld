package tree

import "log/slog"

type value struct {
	name string
	v    slog.Value // must not be slog.KindGroup
	g    *Group
}

func (v *value) LogValue() slog.Value {
	if v.g != nil {
		return slog.GroupValue(v.g.Render()...)
	}
	return v.v
}

func (v *value) clone() *value {
	n := &value{name: v.name}

	if v.g != nil {
		n.g = v.g.Clone()
		return n
	}

	n.v = v.v
	return n
}
