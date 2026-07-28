package execution

import (
	"context"

	"github.com/google/uuid"
)

type contextKey = struct{}

func NewID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func Get(ctx context.Context) uuid.UUID {
	return ctx.Value(contextKey{}).(uuid.UUID)
}

func Set(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}
