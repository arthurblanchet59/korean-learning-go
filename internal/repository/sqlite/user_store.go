package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/arthurblanchet59/korean-learning-go/internal/domain"
	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
)

func (store *Store) EnsureAdmin(ctx context.Context, user domain.User) error {
	var existingID string
	err := store.db.QueryRowContext(ctx, `
		SELECT id
		FROM users
		WHERE is_admin = 1
		ORDER BY created_at ASC
		LIMIT 1
	`).Scan(&existingID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, password_hash, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, ?, ?)
	`, user.ID, user.Name, user.Email, user.PasswordHash, formatTime(user.CreatedAt), formatTime(user.UpdatedAt))
	return normalizeSQLiteError(err)
}

func (store *Store) CreateUser(ctx context.Context, user domain.User) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, password_hash, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, user.ID, user.Name, user.Email, user.PasswordHash, boolInt(user.IsAdmin), formatTime(user.CreatedAt), formatTime(user.UpdatedAt))
	return normalizeSQLiteError(err)
}

func (store *Store) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, name, email, password_hash, is_admin, created_at, updated_at
		FROM users
		WHERE is_admin = 0
		ORDER BY name COLLATE NOCASE, created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (store *Store) FindUserByID(ctx context.Context, id string) (domain.User, error) {
	row := store.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, is_admin, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id)

	return scanUser(row)
}

func (store *Store) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row := store.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, is_admin, created_at, updated_at
		FROM users
		WHERE email = ?
	`, strings.ToLower(strings.TrimSpace(email)))

	return scanUser(row)
}

func (store *Store) UpdateUser(ctx context.Context, user domain.User) error {
	result, err := store.db.ExecContext(ctx, `
		UPDATE users
		SET name = ?,
			email = ?,
			password_hash = ?,
			is_admin = ?,
			updated_at = ?
		WHERE id = ?
	`, user.Name, user.Email, user.PasswordHash, boolInt(user.IsAdmin), formatTime(user.UpdatedAt), user.ID)
	if err != nil {
		return normalizeSQLiteError(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func scanUser(scanner rowScanner) (domain.User, error) {
	var user domain.User
	var isAdmin int
	var createdAt string
	var updatedAt string

	err := scanner.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&isAdmin,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	user.IsAdmin = isAdmin == 1
	user.CreatedAt = parseTime(createdAt)
	user.UpdatedAt = parseTime(updatedAt)
	return user, nil
}
