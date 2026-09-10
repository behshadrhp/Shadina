package storage

import (
	"context"
	"database/sql"
	"time"
)

type UsersStorage struct {
	db *sql.DB
}

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (us *UsersStorage) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users(username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := us.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}
