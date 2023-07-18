package main

import (
	"context"
	"log"

	"github.com/sleeg00/blockchain_go/proto"
	blockchain "github.com/sleeg00/blockchain_go/proto"
	"google.golang.org/grpc"
)

// 가장 높은 길이의 노드 반환
func sendVersion(node string, bestHeight int64) *proto.VersionResponse {
	payload := gobEncode(verzion{nodeVersion, int64(bestHeight), nodeAddress})

	request := append(commandToBytes("version"), payload...)

	conn, err := grpc.Dial("localhost:"+node, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial node %s: %v", node, err)
	}
	defer conn.Close()
	client := proto.NewBlockchainServiceClient(conn)
	req := &blockchain.VersionRequest{
		Address: node,
		Payload: request, // 직렬화된 데이터를 Payload 필드에 직접 할당
	}
	response, err := client.Version(context.Background(), req)

	return response
}
