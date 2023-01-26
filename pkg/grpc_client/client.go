package grpc_client

import (
	"fmt"

	"github.com/mirasildev/medium_post_service/config"
	pbu "github.com/mirasildev/medium_post_service/genproto/user_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClientI interface {
	UserService() pbu.UserServiceClient
}

type GrpcClient struct {
	cfg config.Config
	connections map[string]interface{}
}

func New(cfg config.Config) (GrpcClientI, error) {
	connUserService, err := grpc.Dial(
		fmt.Sprintf("%s%s", cfg.UserServiceHost, cfg.UserServiceGrpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("post service dial host:%s port:%s err: %v", 
			cfg.UserServiceHost, cfg.UserServiceGrpcPort, err)	
	}

	return &GrpcClient{
		cfg: cfg,
		connections: map[string]interface{}{
			"user_service": pbu.NewUserServiceClient(connUserService),
		},
	}, nil
	
}

func (g *GrpcClient) UserService() pbu.UserServiceClient {
	return g.connections["user_service"].(pbu.UserServiceClient)
}