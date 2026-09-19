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
	// initial fields
	ID      int64    `json:"id"`
	UserID  int64    `json:"user_id"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`

	// timestamp
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// relations
	User     User      `json:"user"`
	Comments []Comment `json:"comments"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int64 `json:"comment_count"`
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

func (ps *PostsStorage) GetFeed(ctx context.Context, userID int64) ([]PostWithMetadata, error) {
	query := `
		SELECT
			p.id,
			p.user_id,
			p.title,
			p.content,
			p.created_at,
			p.updated_at,
			p.tags,
			u.username,
			COUNT(c.id) AS comments_count
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
		WHERE f.user_id = $1 OR p.user_id = $1
		GROUP BY p.id, u.username
		ORDER BY p.created_at DESC
	`

	ctx, cancel := context.WithTimeout(ctx, consts.CTX_TIMEOUT)
	defer cancel()

	rows, err := ps.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var feed []PostWithMetadata

	for rows.Next() {
		var p PostWithMetadata
		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.UpdatedAt,
			pq.Array(&p.Tags),
			&p.User.Username,
			&p.CommentsCount,
		)
		if err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	return feed, nil
}
