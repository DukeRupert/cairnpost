package auth

import (
	"context"

	"github.com/dukerupert/cairnpost/internal/model"
)

type contextKey string

const userContextKey contextKey = "auth_user"

func ContextWithUser(ctx context.Context, user model.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (model.User, bool) {
	user, ok := ctx.Value(userContextKey).(model.User)
	return user, ok
}

func MustUserFromContext(ctx context.Context) model.User {
	user, ok := UserFromContext(ctx)
	if !ok {
		panic("auth: no user in context")
	}
	return user
}
