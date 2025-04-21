package http_profile

import (
	"context"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

type contextKey string

func (c contextKey) String() string {
	return fmt.Sprintf("app context key %s", string(c))
}

const bearerTokenContextKey contextKey = "authorization:bearer"

func BearerTokenMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			token := ""

			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}

			newCtx := context.WithValue(c.Request().Context(), bearerTokenContextKey, token)

			c.SetRequest(c.Request().WithContext(newCtx))

			return next(c)
		}
	}
}
