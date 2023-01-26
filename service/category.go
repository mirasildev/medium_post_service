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

type CategoryService struct {
	pb.UnimplementedCategoryServiceServer
	storage storage.StorageI
	logger  *logrus.Logger
}

func NewCategoryService(strg storage.StorageI, logger *logrus.Logger) *CategoryService {
	return &CategoryService{
		storage: strg,
		logger: logger,
	}
}

func (s *CategoryService) Create(ctx context.Context, req *pb.Category) (*pb.Category, error) {
	category, err := s.storage.Category().Create(&repo.Category{
		Title: req.Title,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}
	return parseCategoryModel(category), nil
}

func parseCategoryModel(c *repo.Category) *pb.Category {
	return &pb.Category{
		Id:        c.ID,
		Title:     c.Title,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}

func (s *CategoryService) Get(ctx context.Context, req *pb.GetCategory) (*pb.Category, error) {
	resp, err := s.storage.Category().Get(req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WithError(err).Error("failed to get user")
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &pb.Category{
		Id: resp.ID,
		Title: resp.Title,
		CreatedAt: resp.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *CategoryService) GetAll(ctx context.Context, req *pb.GetAllCategoriesRequest) (*pb.GetAllCategoriesResponse, error) {
	res, err := s.storage.Category().GetAll(&repo.GetAllCategoriesParams{
		Page:   req.Page,
		Limit:  req.Limit,
		Search: req.Search,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal error: %v", err)
	}

	response := pb.GetAllCategoriesResponse{
		Count:      res.Count,
		Categories: make([]*pb.Category, 0),
	}

	for _, user := range res.Categories {
		response.Categories = append(response.Categories, parseCategoryModel(user))
	}

	return &response, nil
}

func (s *CategoryService) Update(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error) {
	category, err := s.storage.Category().Update(&repo.Category{
		ID:    req.Id,
		Title: req.Title,
	})
	if err != nil {
		s.logger.WithError(err).Error("failed to update user")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to update: %v", err)
	}

	return parseCategoryModel(category), nil
}

func (s *CategoryService) Delete(ctx context.Context, req *pb.GetCategory) (*emptypb.Empty, error) {
	err := s.storage.Category().Delete(req.Id)
	if err != nil {
		s.logger.WithError(err).Error("failed to delete user")
		return nil, status.Errorf(codes.Internal, "failed to delete: %v", err)
	}

	return &emptypb.Empty{}, nil
}
