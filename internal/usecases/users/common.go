package users

import (
	"avito/internal/domain"
	"errors"
	"net/mail"
)

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)

	return err == nil
}

func parseRole(role string) (domain.UserRoleID, error) {
	switch role {
	case "moderator":
		return domain.ModeratorUserRoleID, nil
	case "employee":
		return domain.ClientUserRoleID, nil
	default:
		return 0, errors.New(domain.UnknownRoleNameError)
	}
}

const (
	PasswordIsRequiredError string = "password is required"
)
