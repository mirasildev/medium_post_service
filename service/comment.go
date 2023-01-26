package service

import (
	"context"
	"database/sql"
	"errors"

	// "database/sql"
	// "errors"
	"time"

	pb "github.com/mirasildev/medium_post_service/genproto/post_service"
	"github.com/mirasildev/medium_post_service/genproto/user_service"
	grpcPkg "github.com/mirasildev/medium_post_service/pkg/grpc_client"
	"github.com/mirasildev/medium_post_service/storage"
	"github.com/mirasildev/medium_post_service/storage/repo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CommentService struct {
	pb.UnimplementedCommentServiceServer
	storage    storage.StorageI
	grpcClient grpcPkg.GrpcClientI
	logger     *logrus.Logger
}

func NewCommentService(strg storage.StorageI, logger *logrus.Logger, grpcConn grpcPkg.GrpcClientI) *CommentService {
	return &CommentService{
		storage:    strg,
		grpcClient: grpcConn,
		logger:     logger,
	}
}

func (s *CommentService) Create(ctx context.Context, req *pb.Comment) (*pb.Comment, error) {
	com, err := s.storage.Comment().Create(&repo.Comment{
		UserID:      req.UserId,
		PostID:      req.PostId,
		Description: req.Description,
	})
	if err != nil {
		s.logger.WithError(err).Error("failed to create comment")
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}

	user, err := s.grpcClient.UserService().Get(context.Background(), &user_service.IdRequest{Id: req.UserId})
	if err != nil {
		s.logger.WithError(err).Error("failed to get user")
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}

	return parseCommentModel(com, user), nil
}

func (s *CommentService) Get(ctx context.Context, req *pb.GetComment) (*pb.Comment, error) {
	user, err := s.grpcClient.UserService().Get(context.Background(), &user_service.IdRequest{Id: req.UserId})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WithError(err).Error("failed to get user in comment_service")
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
	}

	com, err := s.storage.Comment().Get(req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WithError(err).Error("failed to get comment in comment_service")
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
	}

	return parseCommentModel(com, user), nil

}

func parseCommentModel(c *repo.Comment, u *user_service.User) *pb.Comment {
	return &pb.Comment{
		Id:          c.ID,
		UserId:      u.Id,
		PostId:      c.PostID,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		User: &pb.CommentUser{
			Id:           u.Id,
			FirstName:    u.FirstName,
			LastName:     u.LastName,
			Email:        u.Email,
			ProfileImage: u.ProfileImageUrl,
		},
	}
}

func (s *CommentService) GetAll(ctx context.Context, req *pb.GetAllCommentsRequest) (*pb.GetAllCommentsResponse, error) {
	res, err := s.storage.Comment().GetAll(&repo.GetAllCommentsParams{
		Limit:  req.Limit,
		Page:   req.Page,
		UserID: req.UserId,
		PostID: req.PostId,
	})
	if err != nil {
		s.logger.WithError(err).Error("failed to get all comments in comment_service")
		return nil, status.Errorf(codes.Internal, "failed to get all comments: %v", err)
	}

	response := pb.GetAllCommentsResponse{
		Count:    res.Count,
		Comments: make([]*pb.Comment, 0),
	}

	for _, com := range res.Comments {
		user, err := s.grpcClient.UserService().Get(context.Background(), &user_service.IdRequest{Id: com.UserID})
		if err != nil {
			s.logger.WithError(err).Error("failed to get user in get-all comments")
			return nil, status.Errorf(codes.Internal, "failed to get user in get-all comments; %v", err)
		}

		response.Comments = append(response.Comments, parseCommentModel(com, user))
	}

	return &response, nil
}

func (s *CommentService) Update(ctx context.Context, req *pb.Comment) (*pb.Comment, error) {
	com, err := s.storage.Comment().Update(&repo.Comment{
		ID:          req.Id,
		Description: req.Description,
	})
	if err != nil {
		s.logger.WithError(err).Error("failed to update comment")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.NotFound, "failed to update: %v", err)
	}

	user, err := s.grpcClient.UserService().Get(context.Background(), &user_service.IdRequest{Id: req.UserId})
	if err != nil {
		s.logger.WithError(err).Error("failed to get user when updating the comment")
		return nil, status.Errorf(codes.Internal, "failed to get user when updating the comment: %v", err)
	}

	return &pb.Comment{
		Id:          com.ID,
		Description: com.Description,
		UpdatedAt:   com.UpdatedAt.Format(time.RFC3339),
		User: &pb.CommentUser{
			Id:           user.Id,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Email:        user.Email,
			ProfileImage: user.ProfileImageUrl,
		},
	}, nil
}

func (s *CommentService) Delete(ctx context.Context, req *pb.GetComment) (*emptypb.Empty, error) {
	err := s.storage.Comment().Delete(req.Id)
	if err != nil {
		s.logger.WithError(err).Error("failed to delete comment")
		return nil, status.Errorf(codes.Internal, "failed to delete comment	: %v", err)
	}

	return &emptypb.Empty{}, nil
}
