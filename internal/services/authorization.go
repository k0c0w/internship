package services

import (
	"avito/internal/config"
	domain "avito/internal/domain"
	jwt "avito/pkg/authorization"
	"errors"
	"log"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type (
	authroizationServiceImpl struct {
		userRepository domain.UserRepository
		jwtManager     jwt.JWTManager
		cfg            config.AuthConfig
	}
)

func NewAuthorizationService(jwtManager jwt.JWTManager, userRepository domain.UserRepository, cfg config.AuthConfig) domain.AuthorizationService[jwt.JWT] {
	return authroizationServiceImpl{
		userRepository: userRepository,
		cfg:            cfg,
		jwtManager:     jwtManager,
	}
}

func (s authroizationServiceImpl) SignIn(ctx context.Context, email domain.Email, password string) (jwt.JWT, error) {
	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	} else if match, err := argon2id.ComparePasswordAndHash(password, user.Password); err != nil || !match {
		if err != nil {
			log.Println(err)
		}
		return "", errors.New(domain.BadUserCredentialError)
	}

	token, err := s.jwtManager.GenerateToken(user.ID.String())

	if err != nil {
		log.Println(err)
		return "", errors.New("could not authorize user")
	}

	return token, nil
}

func (s authroizationServiceImpl) SignUp(ctx context.Context, email domain.Email, password string, role domain.UserRoleID) (*domain.User, error) {
	_, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if err.Error() != domain.UserDoesNotExistsError {
			return nil, err
		}
	}

	passwordHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	newUser, err := domain.NewUser(email, passwordHash)
	if err != nil {
		return nil, err
	}

	if role == domain.ModeratorUserRoleID {
		domain.GrantModeratorRole(&newUser)
	}

	err = s.userRepository.Add(ctx, newUser)

	return &newUser, err
}

func (s authroizationServiceImpl) UserFromCredentials(ctx context.Context, credentials jwt.JWT) (*domain.User, error) {
	if credentials == "" {
		return nil, errors.New(domain.InsufficientPrivilegesError)
	}

	claims, err := s.jwtManager.ExtractClaimsFrom(credentials)
	if err != nil {
		log.Println(err)
		return nil, errors.New(domain.InsufficientPrivilegesError)
	}

	userId, err := uuid.Parse(claims.UserID)
	if err != nil {
		log.Println(err)
		return nil, errors.New(domain.InsufficientPrivilegesError)
	}

	user, err := s.userRepository.FindByID(ctx, userId)

	if err != nil {
		if msg := err.Error(); msg != domain.UserDoesNotExistsError {
			log.Panicln(msg)
		}

		return nil, errors.New(domain.InsufficientPrivilegesError)
	}

	return &user, nil
}
