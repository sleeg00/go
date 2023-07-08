package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "github.com/sleeg00/blockchain/protos" // protobuf 파일의 경로에 맞게 수정
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewBlockChainClient(conn)

	// 새로운 트랜잭션 추가 요청
	addTransactionReq := &pb.TransactionRequest{
		Data: "New Transaction",
	}
	addTransactionRes, err := client.AddTransaction(context.Background(), addTransactionReq)
	if err != nil {
		log.Fatalf("Failed to add transaction: %v", err)
	}
	log.Println(addTransactionRes.Message)

	// 체인 유효성 검증 요청
	validateChainReq := &pb.ValidateChainRequest{}
	validateChainRes, err := client.ValidateChain(context.Background(), validateChainReq)
	if err != nil {
		log.Fatalf("Failed to validate chain: %v", err)
	}
	log.Println("Chain is valid:", validateChainRes.IsValid)
}
