package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/behshadrhp/Shadina/internal/consts"
)

type Follower struct {
	UserID     int64     `json:"user_id"`
	FollowerID int64     `json:"follower_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FollowersStorage struct {
	db *sql.DB
}

func (fs *FollowersStorage) Follow(ctx context.Context, userID, followerID int64) error {
	query := `
		INSERT INTO followers(user_id, follower_id)
		VALUES ($1, $2)
	`

	ctx, cancel := context.WithTimeout(ctx, consts.CTX_TIMEOUT)
	defer cancel()

	_, err := fs.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FollowersStorage) Unfollow(ctx context.Context, userID, followerID int64) error {
	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, consts.CTX_TIMEOUT)
	defer cancel()

	_, err := fs.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		return err
	}

	return nil
}
