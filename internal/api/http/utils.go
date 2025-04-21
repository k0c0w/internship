package http_profile

import (
	"avito/internal/usecases"
	jwt "avito/pkg/authorization"
	"context"
)

func (h *httpRequestHandlers) authArgs(ctx context.Context) usecases.AuthenticationArgs {
	return usecases.AuthenticationArgs{
		AuthorizationService: h.deps.AuthorizationService,
		JWT:                  authToken(ctx),
	}
}

func authToken(ctx context.Context) jwt.JWT {

	if tokenStr, ok := ctx.Value(bearerTokenContextKey).(string); ok {
		return tokenStr
	}

	return ""
}
