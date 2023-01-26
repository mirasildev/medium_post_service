package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	pb "github.com/mirasildev/medium_post_service/genproto/post_service"
	"github.com/mirasildev/medium_post_service/storage"
	"github.com/mirasildev/medium_post_service/storage/repo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PostService struct {
	pb.UnimplementedPostServiceServer
	storage storage.StorageI
	logger  *logrus.Logger
}

func NewPostService(strg storage.StorageI, logger *logrus.Logger) *PostService {
	return &PostService{
		storage: strg,
		logger:  logger,
	}
}

func (s *PostService) Create(ctx context.Context, req *pb.Post) (*pb.Post, error) {

	post, err := s.storage.Post().Create(&repo.Post{
		Title:       req.Title,
		Description: req.Description,
		ImageUrl:    req.ImageUrl,
		UserID:      req.UserId,
		CategoryID:  req.CategoryId,
	})
	if err != nil {
		s.logger.WithError(err).Error("failed to create post")
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}
	return parsePostModel(post), nil
}

func parsePostModel(p *repo.Post) *pb.Post {
	return &pb.Post{
		Id:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		ImageUrl:    p.ImageUrl,
		UserId:      p.UserID,
		CategoryId:  p.CategoryID,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		ViewsCount:  p.ViewsCount,
	}
}

func (s *PostService) Get(ctx context.Context, req *pb.GetPost) (*pb.Post, error) {
	post, err := s.storage.Post().Get(req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WithError(err).Error("failed to get post")
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to get post: %v", err)
	}

	return parsePostModel(post), nil
}

func (s *PostService) GetAll(ctx context.Context, req *pb.GetAllPostsRequest) (*pb.GetAllPostsResponse, error) {
	res, err := s.storage.Post().GetAll(&repo.GetAllPostsParams{
		Limit:      req.Limit,
		Page:       req.Page,
		Search:     req.Search,
		CategoryID: req.CategoryId,
		UserID:     req.UserId,
		SortByDate: req.SortByDate,
	})
	if err != nil {
		s.logger.WithError(err).Error("failed to get all posts")
		return nil, status.Errorf(codes.Internal, "failed to get all posts: %v", err)
	}

	response := pb.GetAllPostsResponse{
		Count: res.Count,
		Posts: make([]*pb.Post, 0),
	}

	for _, post := range res.Posts {
		response.Posts = append(response.Posts, parsePostModel(post))
	}

	return &response, nil
}

func (s *PostService) Update(ctx context.Context, req *pb.UpdatePostRequest) (*pb.Post, error) {
	post, err := s.storage.Post().UpdatePost(&repo.Post{
		ID:          req.Id,
		UserID:      req.UserId,
		Title:       req.Title,
		Description: req.Description,
		ImageUrl:    req.ImageUrl,
		CategoryID:  req.CategoryId,
	})
	if err != nil {
		s.logger.WithError(err).Error("failed to update post")
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WithError(err).Error("you can't update other user's post")
			return nil, status.Errorf(codes.NotFound, "you can't update other user's post")
		}
		return nil, status.Errorf(codes.Internal, "failed to update: %v", err)
	}

	return &pb.Post{
		Id:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		ImageUrl:    post.ImageUrl,
		UserId:      post.UserID,
		CategoryId:  post.CategoryID,
		CreatedAt:   post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   post.UpdatedAt.Format(time.RFC3339),
		ViewsCount:  post.ViewsCount,
	}, nil

}

func (s *PostService) Delete(ctx context.Context, req *pb.DeletePost) (*emptypb.Empty, error) {
	err := s.storage.Post().DeletePost(req.Id, req.UserId)
	if err != nil {
		s.logger.WithError(err).Error("failed to delete post")
		return nil, status.Errorf(codes.Internal, "failed to delete post: %v", err)
	}

	return &emptypb.Empty{}, nil
}
