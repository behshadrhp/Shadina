package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/behshadrhp/Shadina/internal/consts"
)

type CommentsStorage struct {
	db *sql.DB
}

type Comment struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	PostID    int64     `json:"post_id"`
	User      User      `json:"user"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cs CommentsStorage) Create(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (user_id, post_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, consts.CTX_TIMEOUT)
	defer cancel()

	err := cs.db.QueryRowContext(
		ctx,
		query,
		comment.UserID,
		comment.PostID,
		comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (cs CommentsStorage) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	query := `
		SELECT c.id, c.user_id, c.post_id, c.content, c.created_at, c.updated_at, users.id, users.username from comments c
		JOIN users on users.id = c.user_id
		where c.post_id = $1
		ORDER BY c.created_at DESC;
	`

	row, err := cs.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	comments := []Comment{}

	for row.Next() {
		var c Comment
		c.User = User{}
		err := row.Scan(
			&c.ID, &c.UserID, &c.PostID, &c.Content, &c.CreatedAt, &c.UpdatedAt,
			&c.User.ID, &c.User.Username,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}
