package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mirasildev/medium_post_service/storage/repo"

	"github.com/jmoiron/sqlx"
)

type commentRepo struct {
	db *sqlx.DB
}

func NewComment(db *sqlx.DB) repo.CommentStorageI {
	return &commentRepo{
		db,
	}
}

func (cr *commentRepo) Create(comment *repo.Comment) (*repo.Comment, error) {
	query := `
		INSERT INTO comments (
			user_id,
		    post_id,
		    description
		) VALUES($1, $2, $3)
		RETURNING id, description, created_at
	`

	row := cr.db.QueryRow(
		query,
		comment.UserID,
		comment.PostID,
		comment.Description,
	)

	err := row.Scan(
		&comment.ID,
		&comment.Description,
		&comment.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (cr *commentRepo) Get(id int64) (*repo.Comment, error) {
	var (
		res repo.Comment
		updatedAt sql.NullTime
	)

	query := `
		SELECT 
			id,
			user_id,
			post_id,
			description,
			created_at,
			updated_at
		FROM comments WHERE id=$1
	`

	err := cr.db.QueryRow(query, id).Scan(
		&res.ID,
		&res.UserID,
		&res.PostID,
		&res.Description,
		&res.CreatedAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	res.UpdatedAt = updatedAt.Time
	return &res, nil
}

func (cr *commentRepo) GetAll(params *repo.GetAllCommentsParams) (*repo.GetAllCommentsResult, error) {
	result := repo.GetAllCommentsResult{
		Comments: make([]*repo.Comment, 0),
	}

	offset := (params.Page - 1) * params.Limit

	limit := fmt.Sprintf(" LIMIT %d OFFSET %d ", params.Limit, offset)

	filter := " WHERE true "
	if params.UserID != 0 {
		filter += fmt.Sprintf(" AND c.user_id=%d ", params.UserID)
	}

	if params.PostID != 0 {
		filter += fmt.Sprintf(" AND c.post_id=%d ", params.PostID)
	}

	query := `
		SELECT
			id,
			user_id,
			post_id,
			description,
			created_at,
			updated_at
		FROM comments 
		` + filter + `
		ORDER BY created_at desc` + limit

	rows, err := cr.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			c         repo.Comment
			updatedAt sql.NullTime
		)

		err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.PostID,
			&c.Description,
			&c.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}
		c.UpdatedAt = updatedAt.Time
		result.Comments = append(result.Comments, &c)
	}

	queryCount := "SELECT count(1) FROM comments " + filter

	err = cr.db.QueryRow(queryCount).Scan(&result.Count)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (cr *commentRepo) Update(com *repo.Comment) (*repo.Comment, error) {
	query := `
		UPDATE comments SET description=$1, updated_at=$2 WHERE id=$3
		RETURNING id, description, updated_at
	`

	var (
		res repo.Comment
		updatedAt sql.NullTime
	)

	err := cr.db.QueryRow(query, com.Description, time.Now(), com.ID).Scan(
		&res.ID,
		&res.Description,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	res.UpdatedAt = updatedAt.Time
	return &res, nil
}

func (cr *commentRepo) Delete(id int64) error {
	query := "DELETE FROM comments WHERE id=$1"
	res, err := cr.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsCount, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsCount == 0 {
		return sql.ErrNoRows
	}

	return nil
}
