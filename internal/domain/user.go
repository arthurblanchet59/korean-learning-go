package domain

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	IsAdmin      bool      `json:"isAdmin"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type PublicUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"isAdmin"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (user User) Public() PublicUser {
	return PublicUser{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
