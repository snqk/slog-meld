package meld

import (
	"context"
	"log/slog"

	"snqk.dev/slog/meld/internal/tree"
)

type handler struct {
	next  slog.Handler
	stack []string
	root  *tree.Group
}

// NewHandler returns a slog.Handler which melds (joins) older slog.Attr(s) with newer updates, making it a mutable slog.Handler.
// It is thread-safe by immutability; handler state is never updated in-place after creation.
// It is recursive; merging slog.KindGroup, and replacing / updating all other types as appropriate.
// It is ordered; slog.Attr(s) configured first will appear in order. Replacing an attribute does not change its position in the order.
func NewHandler(next slog.Handler) slog.Handler {
	return &handler{
		next:  next,
		stack: nil,
		root:  new(tree.Group),
	}
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *handler) Handle(ctx context.Context, rec slog.Record) error {
	newRec := slog.NewRecord(rec.Time, rec.Level, rec.Message, rec.PC)

	if rec.NumAttrs() == 0 { // optimisation
		newRec.AddAttrs(h.root.Render()...)
	} else {
		newRoot := h.root.Clone()
		rec.Attrs(func(attr slog.Attr) bool {
			newRoot.Merge(h.stack, attr)
			return true
		})
		newRec.AddAttrs(newRoot.Render()...)
	}

	return h.next.Handle(ctx, newRec)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newRoot := h.root.Clone()
	newRoot.Merge(h.stack, attrs...)
	return &handler{
		next:  h.next,
		stack: h.stack,
		root:  newRoot,
	}
}

func (h *handler) WithGroup(name string) slog.Handler {
	newRoot := h.root.Clone()
	newRoot.Merge(h.stack, slog.Group(name))
	return &handler{
		next:  h.next,
		stack: append(h.stack, name),
		root:  newRoot,
	}
}
