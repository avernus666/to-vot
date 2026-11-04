package httpserver

import (
	"context"

	"github.com/avernus666/averna/apps/api/internal/auth"
)

type authContextKey struct{}

func withUser(ctx context.Context, user auth.User) context.Context {
	return context.WithValue(ctx, authContextKey{}, user)
}

func userFromContext(ctx context.Context) (auth.User, bool) {
	user, ok := ctx.Value(authContextKey{}).(auth.User)
	return user, ok
}
