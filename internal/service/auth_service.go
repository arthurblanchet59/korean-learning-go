package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/arthurblanchet59/korean-learning-go/internal/domain"
	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrForbidden = errors.New("forbidden")

type AuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
	now       func() time.Time
}

type AuthResult struct {
	Token string            `json:"token"`
	User  domain.PublicUser `json:"user"`
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type UpdateUserInput struct {
	Name     string
	Email    string
	Password string
}

type TokenClaims struct {
	UserID  string `json:"userId"`
	IsAdmin bool   `json:"isAdmin"`
	jwt.RegisteredClaims
}

func NewAuthService(users repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (service *AuthService) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	name := strings.TrimSpace(input.Name)
	email := normalizeEmail(input.Email)
	if name == "" || email == "" || len(input.Password) < 8 {
		return AuthResult{}, fmt.Errorf("name, valid email and password with at least 8 characters are required")
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return AuthResult{}, err
	}

	now := service.now()
	user := domain.User{
		ID:           uuid.NewString(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		IsAdmin:      false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := service.users.CreateUser(ctx, user); err != nil {
		return AuthResult{}, err
	}

	return service.authResult(user)
}

func (service *AuthService) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	user, err := service.users.FindUserByEmail(ctx, normalizeEmail(input.Email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	return service.authResult(user)
}

func (service *AuthService) UserByID(ctx context.Context, userID string) (domain.PublicUser, error) {
	user, err := service.users.FindUserByID(ctx, userID)
	if err != nil {
		return domain.PublicUser{}, err
	}

	return user.Public(), nil
}

func (service *AuthService) UpdateSelf(ctx context.Context, userID string, input UpdateUserInput) (domain.PublicUser, error) {
	user, err := service.users.FindUserByID(ctx, userID)
	if err != nil {
		return domain.PublicUser{}, err
	}

	updated, err := service.applyUserUpdate(user, input, false)
	if err != nil {
		return domain.PublicUser{}, err
	}

	if err := service.users.UpdateUser(ctx, updated); err != nil {
		return domain.PublicUser{}, err
	}

	return updated.Public(), nil
}

func (service *AuthService) AdminUpdateUser(ctx context.Context, userID string, input UpdateUserInput) (domain.PublicUser, error) {
	user, err := service.users.FindUserByID(ctx, userID)
	if err != nil {
		return domain.PublicUser{}, err
	}
	if user.IsAdmin {
		return domain.PublicUser{}, ErrForbidden
	}

	updated, err := service.applyUserUpdate(user, input, false)
	if err != nil {
		return domain.PublicUser{}, err
	}

	if err := service.users.UpdateUser(ctx, updated); err != nil {
		return domain.PublicUser{}, err
	}

	return updated.Public(), nil
}

func (service *AuthService) ParseToken(tokenValue string) (TokenClaims, error) {
	claims := TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenValue, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return service.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return TokenClaims{}, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (service *AuthService) EnsureAdmin(ctx context.Context, name string, email string, password string) error {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return err
	}

	now := service.now()
	return service.users.EnsureAdmin(ctx, domain.User{
		ID:           "admin",
		Name:         strings.TrimSpace(name),
		Email:        normalizeEmail(email),
		PasswordHash: passwordHash,
		IsAdmin:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (service *AuthService) authResult(user domain.User) (AuthResult, error) {
	expiresAt := service.now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		UserID:  user.ID,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(service.now()),
		},
	})

	signedToken, err := token.SignedString(service.jwtSecret)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{Token: signedToken, User: user.Public()}, nil
}

func (service *AuthService) applyUserUpdate(user domain.User, input UpdateUserInput, allowAdminChange bool) (domain.User, error) {
	if strings.TrimSpace(input.Name) != "" {
		user.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Email) != "" {
		user.Email = normalizeEmail(input.Email)
	}
	if input.Password != "" {
		if len(input.Password) < 8 {
			return domain.User{}, fmt.Errorf("password must contain at least 8 characters")
		}
		passwordHash, err := hashPassword(input.Password)
		if err != nil {
			return domain.User{}, err
		}
		user.PasswordHash = passwordHash
	}
	user.UpdatedAt = service.now()
	return user, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
