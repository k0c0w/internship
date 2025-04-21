package usecases

import (
	"avito/internal/domain"
	jwt "avito/pkg/authorization"
	"context"
	"errors"
)

type AuthenticationArgs struct {
	AuthorizationService domain.AuthorizationService[jwt.JWT]
	jwt.JWT
}

const (
	IdIsRequiredArgError string = "id is required"
)

func (args *AuthenticationArgs) ValidatePrivelegies(ctx context.Context, roleIds ...domain.UserRoleID) (*domain.User, error) {
	user, err := args.AuthorizationService.UserFromCredentials(ctx, args.JWT)
	if err != nil {
		return nil, err
	}

	if len(roleIds) == 0 {
		return user, nil
	}

	for _, roleId := range roleIds {
		if user.UserRole.ID == roleId {
			return user, nil
		}
	}

	return user, errors.New(domain.InsufficientPrivilegesError)
}
