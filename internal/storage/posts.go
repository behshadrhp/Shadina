package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"github.com/behshadrhp/Shadina/internal/consts"
)

type PostsStorage struct {
	db *sql.DB
}

type Post struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	Comments  []Comment `json:"comments"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ps *PostsStorage) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts(user_id, title, content, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := ps.db.QueryRowContext(
		ctx,
		query,
		post.UserID,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (ps *PostsStorage) List(ctx context.Context) ([]*Post, error) {
	query := `
		SELECT 
			id, 
			user_id, 
			title, 
			content, 
			tags, 
			created_at, 
			updated_at
		FROM posts
	`

	row, err := ps.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var posts []*Post

	for row.Next() {
		var post Post

		err := row.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			pq.Array(&post.Tags),
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		posts = append(posts, &post)
	}

	return posts, nil
}

func (ps *PostsStorage) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT 
			id, 
			user_id, 
			title, 
			content, 
			tags, 
			created_at, 
			updated_at
		FROM posts
		WHERE id=$1
	`

	var post Post

	err := ps.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, consts.NOT_FOUND
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (ps *PostsStorage) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM posts
		WHERE id=$1
	`

	res, err := ps.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return consts.NOT_FOUND
	}

	return nil
}

func (ps *PostsStorage) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title=$1, content=$2
		WHERE id=$3
	`

	_, err := ps.db.ExecContext(ctx, query, post.Title, post.Content, post.ID)
	if err != nil {
		return err
	}

	return nil
}
