package postgres_test

import (
	"testing"

	"github.com/mirasildev/medium_post_service/storage/repo"
	"github.com/stretchr/testify/require"
)

func createLike(t *testing.T) *repo.Like {
	like, err := strg.Like().CreateOrUpdate(&repo.Like{
		PostID: 6,
		UserID: 2,
		Status: true,
	})
	require.NoError(t, err)

	return like
}

func TestCreate(t *testing.T) {
	createLike(t)
}

func TestGetLike(t *testing.T) {
	like := createLike(t)
	like2, err := strg.Like().Get(like.UserID, like.PostID)
	require.NoError(t, err)
	require.Equal(t, like.PostID, like2.PostID)

}
