package service

import (
	"context"
	"database/sql"
	"errors"

	pb "github.com/mirasildev/medium_post_service/genproto/post_service"
	"github.com/mirasildev/medium_post_service/storage"
	"github.com/mirasildev/medium_post_service/storage/repo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LikeService struct {
	pb.UnimplementedLikeServiceServer
	storage storage.StorageI
	logger  *logrus.Logger
}

func NewLikeService(strg storage.StorageI, logger *logrus.Logger) *LikeService {
	return &LikeService{
		storage: strg,
		logger:  logger,
	}
}

func (s *LikeService) CreateOrUpdate(ctx context.Context, req *pb.Like) (*pb.Like, error) {
	like, err := s.storage.Like().CreateOrUpdate(&repo.Like{
		PostID: req.PostId,
		UserID: req.UserId,
		Status: req.Status,
	})
	if err != nil {
		s.logger.WithError(err).Error("failed to create-update like")
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}

	return parseLikeModel(like), nil
}

func parseLikeModel(like *repo.Like) *pb.Like {
	return &pb.Like{
		Id:     like.ID,
		PostId: like.PostID,
		UserId: like.UserID,
		Status: like.Status,
	}
}

func (s *LikeService) Get(ctx context.Context, req *pb.GetLike) (*pb.Like, error) {
	like, err := s.storage.Like().Get(req.UserId, req.PostId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WithError(err).Error("failed to get like")
			return nil, status.Errorf(codes.Internal, "failed to get like: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get like: %v", err)
	}

	return parseLikeModel(like), nil
}

func (s *LikeService) GetAllLikesCount(ctx context.Context, req *pb.GetLike) (*pb.AllLikesCount, error) {
	res, err := s.storage.Like().GetLikesDislikesCount(req.PostId)
	if err != nil {
		s.logger.WithError(err).Error("failed to get likes and dislikes count")
		return nil, status.Errorf(codes.Internal, "failed to get likes and dislikes count: %v", err)
	}

	return &pb.AllLikesCount{
		LikesCount:    res.LikesCount,
		DislikesCount: res.DislikesCount,
	}, nil
}
