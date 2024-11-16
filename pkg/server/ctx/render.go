package ctx

import (
	"context"
	"io"
)

type Component interface {
	Render(ctx context.Context, wr io.Writer) error
}

func (c Ctx) Render(f Component) error {
	return f.Render(c.Context(), c.wr)
}
