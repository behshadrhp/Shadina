package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/behshadrhp/Shadina/internal/consts"
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

func (us *UsersStorage) List(ctx context.Context) ([]*User, error) {
	query := `
		SELECT
			id,
			username,
			email,
			password,
			created_at,
			updated_at
		FROM users
	`

	rows, err := us.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	return users, nil
}

func (us *UsersStorage) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT 
			id,
			username,
			email,
			password,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	var user User

	err := us.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, consts.NOT_FOUND
		default:
			return nil, err
		}
	}

	return &user, err
}
