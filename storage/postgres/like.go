package postgres

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/mirasildev/medium_post_service/storage/repo"
)

type likeRepo struct {
	db *sqlx.DB
}

func NewLike(db *sqlx.DB) repo.LikeStorageI {
	return &likeRepo{
		db: db,
	}
}

func (lr *likeRepo) CreateOrUpdate(l *repo.Like) (*repo.Like, error) {
	like, err := lr.Get(l.UserID, l.PostID)
	if errors.Is(err, sql.ErrNoRows) {
		query := `
			INSERT INTO likes(user_id, post_id, status) 
			VALUES($1, $2, $3) RETURNING id
		`

		_, err := lr.db.Exec(query, l.UserID, l.PostID, l.Status)
		if err != nil {
			return nil, err
		}
	} else if like != nil {
		if like.Status == l.Status {
			_, err := lr.db.Exec(`DELETE FROM likes WHERE id=$1`, like.ID)
			if err != nil {
				return nil, err
			}
		} else {
			query := `UPDATE likes SET status=$1 WHERE id=$2`
			_, err := lr.db.Exec(query, l.Status, l.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	return l, nil
}

func (lr *likeRepo) Get(userID, postID int64) (*repo.Like, error) {
	var result repo.Like

	query := `
		SELECT
			id,
			user_id,
			post_id,
			status
		FROM likes
		WHERE user_id = $1 AND post_id = $2
	`

	err := lr.db.QueryRow(query, userID, postID).Scan(
		&result.ID,
		&result.UserID,
		&result.PostID,
		&result.Status,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (lr *likeRepo) GetLikesDislikesCount(postID int64) (*repo.LikesDislikesCountsResult, error) {
	var result repo.LikesDislikesCountsResult

	query := `
		SELECT 
				COUNT(1) FILTER (WHERE status=true) as likes_count,
				COUNT(1) FILTER (WHERE status=false) as dislikes_count
		FROM likes
		WHERE post_id=$1
	`

	row := lr.db.QueryRow(query, postID)
	err := row.Scan(
		&result.LikesCount,
		&result.DislikesCount,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
