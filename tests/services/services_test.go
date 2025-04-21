package services_test

import (
	"avito/internal/domain"
	"avito/internal/services"
	jwt "avito/pkg/authorization"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/stretchr/testify/assert"
)

var jwtManager jwt.JWTManager = *jwt.NewJWTManager("secret", "testing", time.Hour)

var ctx = context.TODO()

func TestAuthorizationService_SignIn(t *testing.T) {
	tests := []struct {
		name        string
		email       domain.Email
		password    string
		userSetup   func(repo FakeUserRepository) domain.User
		expectToken bool
		expectErr   string
	}{
		{
			name:     "Success",
			email:    "test@example.com",
			password: "password123",
			userSetup: func(repo FakeUserRepository) domain.User {
				hash := hash("password123")
				user, _ := domain.NewUser("test@example.com", hash)
				repo.Add(ctx, user)
				return user
			},
			expectToken: true,
			expectErr:   "",
		},
		{
			name:     "User not found",
			email:    "unknown@example.com",
			password: "password123",
			userSetup: func(repo FakeUserRepository) domain.User {
				return domain.User{}
			},
			expectToken: false,
			expectErr:   domain.UserDoesNotExistsError,
		},
		{
			name:     "Wrong password",
			email:    "test@example.com",
			password: "wrongpassword",
			userSetup: func(repo FakeUserRepository) domain.User {
				hash := hash("password123")
				user, _ := domain.NewUser("test@example.com", hash)
				repo.Add(ctx, user)
				return user
			},
			expectToken: false,
			expectErr:   domain.BadUserCredentialError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := NewFakeUserRepository()
			jwtManager := jwtManager
			svc := services.NewAuthorizationService(jwtManager, repo)
			ctx := context.Background()
			user := tt.userSetup(repo)

			// Act
			token, err := svc.SignIn(ctx, tt.email, tt.password)

			// Assert
			if tt.expectErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectErr, err.Error())
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				if tt.expectToken {
					assert.NotEmpty(t, token)
					claims, _ := jwtManager.ExtractClaimsFrom(token)
					assert.Equal(t, user.ID.String(), claims.UserID)
				}
			}
		})
	}
}

func TestAuthorizationService_SignUp(t *testing.T) {
	tests := []struct {
		name      string
		email     domain.Email
		password  string
		role      domain.UserRoleID
		userSetup func(repo FakeUserRepository)
		expectErr bool
	}{
		{
			name:     "Success - Regular user",
			email:    "new@example.com",
			password: "password123",
			role:     domain.ClientUserRoleID,
			userSetup: func(repo FakeUserRepository) {
			},
			expectErr: false,
		},
		{
			name:     "Success - Moderator",
			email:    "mod@example.com",
			password: "password123",
			role:     domain.ModeratorUserRoleID,
			userSetup: func(repo FakeUserRepository) {
			},
			expectErr: false,
		},
		{
			name:     "Email already exists",
			email:    "existing@example.com",
			password: "password123",
			role:     domain.ClientUserRoleID,
			userSetup: func(repo FakeUserRepository) {
				hash := hash("password123")
				user, _ := domain.NewUser("existing@example.com", hash)
				repo.Add(ctx, user)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := NewFakeUserRepository()
			svc := services.NewAuthorizationService(jwtManager, repo)
			ctx := context.Background()

			// Setup user
			tt.userSetup(repo)

			// Act
			user, err := svc.SignUp(ctx, tt.email, tt.password, tt.role)

			// Assert
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				if tt.role == domain.ModeratorUserRoleID {
					assert.Equal(t, domain.ModeratorUserRoleID, user.UserRole.ID)
				} else {
					assert.NotEqual(t, domain.ModeratorUserRoleID, user.UserRole.ID)
				}
				match := compare(tt.password, user.Password)
				assert.True(t, match)
				stored, _ := repo.FindByEmail(ctx, tt.email)
				assert.Equal(t, user.ID, stored.ID)
			}
		})
	}
}

func TestAuthorizationService_UserFromCredentials(t *testing.T) {
	tests := []struct {
		name        string
		credentials jwt.JWT
		userSetup   func(repo FakeUserRepository) (domain.User, jwt.JWT)
		expectErr   string
	}{
		{
			name:        "Success",
			credentials: "",
			userSetup: func(repo FakeUserRepository) (domain.User, jwt.JWT) {
				user, _ := domain.NewUser("test@example.com", "hash")
				repo.Add(context.Background(), user)
				token, _ := jwtManager.GenerateToken(user.ID.String())
				return user, token
			},
			expectErr: "",
		},
		{
			name:        "Empty credentials",
			credentials: "",
			userSetup: func(repo FakeUserRepository) (domain.User, jwt.JWT) {
				return domain.User{}, ""
			},
			expectErr: domain.InsufficientPrivilegesError,
		},
		{
			name:        "Invalid token",
			credentials: "invalid-token",
			userSetup: func(repo FakeUserRepository) (domain.User, jwt.JWT) {
				return domain.User{}, ""
			},
			expectErr: domain.InsufficientPrivilegesError,
		},
		{
			name:        "User not found",
			credentials: "",
			userSetup: func(repo FakeUserRepository) (domain.User, jwt.JWT) {
				user, _ := domain.NewUser("test@example.com", "hash")
				token, _ := jwtManager.GenerateToken(user.ID.String())

				return user, token
			},
			expectErr: domain.InsufficientPrivilegesError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := NewFakeUserRepository()
			svc := services.NewAuthorizationService(jwtManager, repo)
			ctx := context.Background()

			// Setup user and token
			user, token := tt.userSetup(repo)
			if tt.credentials == "" {
				tt.credentials = token
			}

			// Act
			result, err := svc.UserFromCredentials(ctx, tt.credentials)

			// Assert
			if tt.expectErr != "" {
				assert.Error(t, err, "Expected an error")
				assert.Equal(t, tt.expectErr, err.Error(), "Error message mismatch")
				assert.Nil(t, result, "User should be nil on error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.NotNil(t, result, "Expected a user")
				assert.Equal(t, user.ID, result.ID, "User ID should match")
				assert.Equal(t, user.Email, result.Email, "User email should match")
			}
		})
	}
}

func hash(value string) string {
	hash, _ := argon2id.CreateHash("password123", argon2id.DefaultParams)

	return hash
}

func compare(value, hash string) bool {
	match, _ := argon2id.ComparePasswordAndHash(value, hash)
	return match
}

type FakeUserRepository struct {
	emailMap map[domain.Email]*domain.User
	idMap    map[domain.UserID]*domain.User
}

func (f FakeUserRepository) Add(ctx context.Context, user domain.User) error {
	ref := &user
	f.idMap[user.ID] = ref
	f.emailMap[user.Email] = ref

	return nil
}

func (f FakeUserRepository) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	user, exists := f.emailMap[email]
	if !exists {
		return domain.User{}, errors.New(domain.UserDoesNotExistsError)
	}

	return *user, nil
}

func (f FakeUserRepository) FindByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	user, exists := f.idMap[id]
	if !exists {
		return domain.User{}, errors.New(domain.UserDoesNotExistsError)
	}

	return *user, nil
}

func NewFakeUserRepository() FakeUserRepository {
	return FakeUserRepository{
		emailMap: make(map[domain.Email]*domain.User),
		idMap:    make(map[domain.UserID]*domain.User),
	}
}
