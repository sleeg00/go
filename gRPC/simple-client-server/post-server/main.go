// simple-client-server/post-server/main.go
package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/sleeg00/gRPC/data"
	postpb "github.com/sleeg00/gRPC/protos/v1/post"
	userpb "github.com/sleeg00/gRPC/protos/v1/user"
	client "github.com/sleeg00/gRPC/simple-client-server"
)

const portNumber = "9001"

type postServer struct {
	postpb.PostServer

	userCli userpb.UserClient //user 서비스 이용 가능하도록 설정
}

// ListPostsByUserId returns post messages by user_id
func (s *postServer) ListPostsByUserId(ctx context.Context, req *postpb.ListPostsByUserIdRequest) (*postpb.ListPostsByUserIdResponse, error) {
	userID := req.UserId

	resp, err := s.userCli.GetUser(ctx, &userpb.GetUserRequest{UserId: userID})

	//user gRPC server의 GetUSer rpc 호출
	if err != nil {
		return nil, err
	}

	var postMessages []*postpb.PostMessage

	for _, up := range data.UserPosts {
		if up.UserID != userID {
			continue
		}

		for _, p := range up.Posts {
			p.Author = resp.UserMessage.Name //id에 담긴 이름
		}

		postMessages = up.Posts
		break
	}

	return &postpb.ListPostsByUserIdResponse{
		PostMessages: postMessages,
	}, nil
}

// ListPosts returns all post messages
func (s *postServer) ListPosts(ctx context.Context, req *postpb.ListPostsRequest) (*postpb.ListPostsResponse, error) {
	var postMessages []*postpb.PostMessage
	for _, up := range data.UserPosts {
		resp, err := s.userCli.GetUser(ctx, &userpb.GetUserRequest{UserId: up.UserID})
		if err != nil {
			return nil, err
		}

		for _, p := range up.Posts {
			p.Author = resp.UserMessage.Name
		}

		postMessages = append(postMessages, up.Posts...)
	}

	return &postpb.ListPostsResponse{
		PostMessages: postMessages,
	}, nil
}

func main() {

	lis, err := net.Listen("tcp", ":"+portNumber) //9001번 포트로 연결 대기
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	userCli := client.GetUserClient("localhost:9000") //User gRPC 서버와 연결을 맺는 userCli 선언\

	//grpc패키지로 보안을 거친 후 이제 통신이 가능함
	grpcServer := grpc.NewServer() //서버 생성 9001포트로~

	postpb.RegisterPostServer(grpcServer, &postServer{
		userCli: userCli,
	})

	log.Printf("start gRPC server on %s port", portNumber)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
