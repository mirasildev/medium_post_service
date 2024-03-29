package main

import (
	"fmt"
	"log"
	"net"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/mirasildev/medium_post_service/config"
	pb "github.com/mirasildev/medium_post_service/genproto/post_service"
	"github.com/mirasildev/medium_post_service/pkg/logger"
	grpcPkg "github.com/mirasildev/medium_post_service/pkg/grpc_client"
	"github.com/mirasildev/medium_post_service/service"
	"github.com/mirasildev/medium_post_service/storage"
)

func main() {
	cfg := config.Load(".")

	psqlUrl := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
	)

	psqlConn, err := sqlx.Connect("postgres", psqlUrl)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	strg := storage.NewStoragePg(psqlConn)
	logrus := logger.New()

	grpcConn, err := grpcPkg.New(cfg)
	if err != nil {
		log.Fatalf("failed to get grpc connections: %v", err)
	}

	postService := service.NewPostService(strg, logrus)
	categoryService := service.NewCategoryService(strg, logrus)
	commentService := service.NewCommentService(strg, logrus, grpcConn)
	likeService := service.NewLikeService(strg, logrus)

	lis, err := net.Listen("tcp", cfg.GrpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	pb.RegisterPostServiceServer(s, postService)
	pb.RegisterCategoryServiceServer(s, categoryService)
	pb.RegisterCommentServiceServer(s, commentService)
	pb.RegisterLikeServiceServer(s, likeService)

	log.Println("Grpc server started in port ", cfg.GrpcPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error while listening: %v", err)
	}

}
