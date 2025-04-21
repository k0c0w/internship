package users

import (
	"avito/internal/domain"
	jwt "avito/pkg/authorization"
	"context"
	"errors"
)

type LoginUserUseCaseArgs struct {
	AuthorizationService domain.AuthorizationService[jwt.JWT]

	User LoginUserDTO
}

type LoginUserDTO struct {
	Email    string
	Password string
}

func LoginUserUseCase(ctx context.Context, args LoginUserUseCaseArgs) (jwt.JWT, error) {
	loginDto := args.User

	if !isValidEmail(loginDto.Email) {
		return "", errors.New(domain.InvalidEmail)
	} else if loginDto.Password == "" {
		return "", errors.New(PasswordIsRequiredError)
	}

	token, err := args.AuthorizationService.SignIn(ctx, domain.Email(loginDto.Email), loginDto.Password)

	return token, err
}
