package storage

import (
	"context"
	"database/sql"
)

type Storage struct {
	Posts interface {
		Create(ctx context.Context, post *Post) error
		List(ctx context.Context) ([]*Post, error)
	}
	Users interface {
		Create(ctx context.Context, user *User) error
	}
}

func NewPostgresql(db *sql.DB) Storage {
	return Storage{
		Posts: &PostsStorage{db},
		Users: &UsersStorage{db},
	}
}
