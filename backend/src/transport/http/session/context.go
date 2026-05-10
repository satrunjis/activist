package session

import (
	"context"
	"net/http"
	"time"

	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
)

type actorContextKey string

const actorKey actorContextKey = "auth_actor"

type Actor struct {
	User                 domainuser.User
	UserID               shared.UserID
	SessionID            string
	SessionToken         string
	SessionExpiresAt     time.Time
	SessionIdleExpiresAt time.Time
	CSRFToken            string
}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorKey, actor)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
	if ctx == nil {
		return Actor{}, false
	}
	value := ctx.Value(actorKey)
	actor, ok := value.(Actor)
	return actor, ok
}

func ActorFromRequest(r *http.Request) (Actor, bool) {
	if r == nil {
		return Actor{}, false
	}
	return ActorFromContext(r.Context())
}
