package repo

import "time"

type Post struct {
	ID          int64
	Title       string
	Description string
	ImageUrl    string
	UserID      int64
	CategoryID  int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ViewsCount  int32
}

type GetAllPostsParams struct {
	Limit      int32
	Page       int32
	Search     string
	CategoryID int64
	UserID     int64
	SortByDate string
}

type GetAllPostsResult struct {
	Posts []*Post
	Count int32
}

type PostStorageI interface {
	Create(p *Post) (*Post, error)
	Get(id int64) (*Post, error)
	GetAll(params *GetAllPostsParams) (*GetAllPostsResult, error)
	UpdatePost(p *Post) (*Post, error)
	DeletePost(id int64, UserID int64) error
}
