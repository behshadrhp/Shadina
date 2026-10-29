package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

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
	RoleID    int       `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) HashedPassword(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hash)

	return nil
}

func (u *User) ComparePassword(text string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(text))
	if err != nil {
		return false
	}

	return true
}

func (us *UsersStorage) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users(username, email, password, role_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := us.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
		user.RoleID,
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
			role_id,
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
		&user.RoleID,
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

func (us *UsersStorage) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT 
			id,
			username,
			email,
			password,
			role_id,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`

	var user User

	err := us.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.RoleID,
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
