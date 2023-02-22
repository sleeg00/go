package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/sleeg00/gRPC/data"
	userpb "github.com/sleeg00/gRPC/protos/v1/user"
)

const portNumber = "9000"

type userServer struct {
	userpb.UserServer
}

// GetUser returns user message by user_id
// user_id로 유저의 정보를 갖고오는 GetUser
func (s *userServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	//(Context)는 작업 명세서와 같은 역할로, 작업 가능한 시간, 작업 취소 등 작업의 흐름을 제어하는데 사용됨
	userID := req.UserId

	var userMessage *userpb.UserMessage
	for _, u := range data.Users {
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
// rpc와 유저들의 정보 모두를 갖고오는 ListUsers
func (s *userServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	userMessages := make([]*userpb.UserMessage, len(data.Users))
	for i, u := range data.Users {
		userMessages[i] = u
	}

	return &userpb.ListUsersResponse{
		UserMessages: userMessages,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":"+portNumber) //TCP 프로토콜에 9000 호트로 연결을 받음 (현재 연결 대기상태)
	if err != nil {
		log.Fatalf("failed to listen: %v", err) //에러면 출력 후 종료
	} //형식은 %v -> 아무거나 가능
	//연결 대기 닫을 거면 lis.Close() lis.Accep() -> 연결되면 리턴 conn, err :=

	grpcServer := grpc.NewServer() //gRPC server를 만든다
	userpb.RegisterUserServer(grpcServer, &userServer{})
	//user.pb에 있는 RegisterUserServer 메소드를 불러와
	// user서비스를 등록 -> user서비스를 담당하는 gRPC server생성
	log.Printf("start gRPC server on %s port", portNumber)
	if err := grpcServer.Serve(lis); err != nil { //listener connection을 위해 Serve()라는 함수 인자로 넣어줌
		log.Fatalf("failed to serve: %s", err)
	}
}
