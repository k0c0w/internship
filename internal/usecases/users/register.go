package users

import (
	"avito/internal/domain"
	jwt "avito/pkg/authorization"
	"context"
	"errors"
)

type RegisterUserUseCaseArgs struct {
	AuthorizationService domain.AuthorizationService[jwt.JWT]

	User RegisterUserDTO
}

type RegisterUserDTO struct {
	Email    string
	Password string
	Role     string
}

func RegisterUserUseCase(ctx context.Context, args RegisterUserUseCaseArgs) (*domain.User, error) {
	registerDto := args.User

	if !isValidEmail(registerDto.Email) {
		return nil, errors.New(domain.InvalidEmail)
	} else if registerDto.Password == "" {
		return nil, errors.New(PasswordIsRequiredError)
	}

	roleId, err := parseRole(registerDto.Role)
	if err != nil {
		roleId = domain.ClientUserRoleID
	}

	//todo: unit of work
	user, err := args.AuthorizationService.SignUp(ctx, domain.Email(registerDto.Email), registerDto.Password, roleId)

	return user, err
}
