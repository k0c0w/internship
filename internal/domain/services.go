package domain

import "context"

type AuthorizationService[TCredentials any] interface {
	SignIn(ctx context.Context, email Email, password string) (TCredentials, error)
	SignUp(ctx context.Context, email Email, password string, role UserRoleID) (*User, error)
	UserFromCredentials(ctx context.Context, credentials TCredentials) (*User, error)
}
