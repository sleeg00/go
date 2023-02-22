package simple_client_server

import (
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userpb "github.com/sleeg00/gRPC/protos/v1/user"
)

var (
	once sync.Once
	cli  userpb.UserClient
)

func GetUserClient(serviceHost string) userpb.UserClient {

	once.Do(func() {

		conn, _ := grpc.Dial(serviceHost,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)

		cli = userpb.NewUserClient(conn)
	})

	return cli
}
