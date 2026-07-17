package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/arthurblanchet59/korean-learning-go/internal/backend/domain"
	"github.com/arthurblanchet59/korean-learning-go/internal/backend/repository"
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
	if len([]rune(name)) < 2 {
		return AuthResult{}, validationErrorf("le nom doit contenir au moins 2 caractères")
	}
	if !validEmail(email) {
		return AuthResult{}, validationErrorf("l'adresse email n'est pas valide")
	}
	if err := validatePassword(input.Password); err != nil {
		return AuthResult{}, err
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

	if registration, ok := service.users.(repository.UserRegistrationRepository); ok {
		if err := registration.CreateUserWithSeed(ctx, user); err != nil {
			return AuthResult{}, err
		}
	} else if err := service.users.CreateUser(ctx, user); err != nil {
		return AuthResult{}, err
	} else if seeder, ok := service.users.(repository.UserDataSeeder); ok {
		if err := seeder.SeedUser(ctx, user.ID); err != nil {
			return AuthResult{}, fmt.Errorf("seed user study data: %w", err)
		}
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

func (service *AuthService) ListUsers(ctx context.Context) ([]domain.PublicUser, error) {
	users, err := service.users.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.PublicUser, 0, len(users))
	for _, user := range users {
		result = append(result, user.Public())
	}
	return result, nil
}

func (service *AuthService) UpdateSelf(ctx context.Context, userID string, input UpdateUserInput) (domain.PublicUser, error) {
	user, err := service.users.FindUserByID(ctx, userID)
	if err != nil {
		return domain.PublicUser{}, err
	}

	updated, err := service.applyUserUpdate(user, input)
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

	updated, err := service.applyUserUpdate(user, input)
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

func (service *AuthService) applyUserUpdate(user domain.User, input UpdateUserInput) (domain.User, error) {
	if strings.TrimSpace(input.Name) != "" {
		user.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Email) != "" {
		email := normalizeEmail(input.Email)
		if !validEmail(email) {
			return domain.User{}, validationErrorf("l'adresse email n'est pas valide")
		}
		user.Email = email
	}
	if input.Password != "" {
		if err := validatePassword(input.Password); err != nil {
			return domain.User{}, err
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

// maxPasswordBytes is the hard limit bcrypt enforces: bytes past the 72nd are
// ignored, so we reject longer passwords instead of silently truncating them.
const maxPasswordBytes = 72

func validatePassword(password string) error {
	if len([]rune(password)) < 8 {
		return validationErrorf("le mot de passe doit contenir au moins 8 caractères")
	}
	if len(password) > maxPasswordBytes {
		return validationErrorf("le mot de passe ne doit pas dépasser 72 octets")
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
