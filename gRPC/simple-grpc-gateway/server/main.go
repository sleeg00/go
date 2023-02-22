package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/sleeg00/gRPC/data"
	userpb "github.com/sleeg00/gRPC/protos/v2/user"
)

const portNumber = "9001"

type userServer struct {
	userpb.UserServer
}

// GetUser returns user message by user_id
func (s *userServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	userID := req.UserId

	var userMessage *userpb.UserMessage
	for _, u := range data.UsersV2 {
		if u.UserId != userID {
			continue
		}
		userMessage = u
		break
	}

	return &userpb.GetUserResponse{
		UserMessage: userMessage,
	}, nil
}

// ListUsers returns all user messages
func (s *userServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	userMessages := make([]*userpb.UserMessage, len(data.UserData))
	for i, u := range data.UsersV2 {
		userMessages[i] = u
	}

	return &userpb.ListUsersResponse{
		UserMessages: userMessages,
	}, nil
}

func customMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		log.Print("Requested at:", time.Now())

		resp, err := handler(ctx, req)
		return resp, err
	}
}

func main() {
	lis, err := net.Listen("tcp", ":"+portNumber)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// grpcServer := grpc.NewServer() 기존 gRPC middleware 없는 방식
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			customMiddleware(),
		)),
	)
	userpb.RegisterUserServer(grpcServer, &userServer{})

	log.Printf("start gRPC server on %s port", portNumber)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
