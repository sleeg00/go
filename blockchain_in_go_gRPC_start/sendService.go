package main

import (
	"context"
	"log"

	"github.com/sleeg00/blockchain_go/proto"
	blockchain "github.com/sleeg00/blockchain_go/proto"
	"google.golang.org/grpc"
)

type tx struct {
	AddFrom     string
	Transaction []byte
}

// 마이낭하고 있는 사람이 없다면 모든 노드에게 TX를 보내서 일관성을 유지하자.
func sendTxToNode(address string, tnx *Transaction) string {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	// 서버와의 연결 설정
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial node %s: %v", address, err)
	}
	defer conn.Close()

	// 클라이언트 생성
	client := proto.NewBlockchainServiceClient(conn)

	req := &blockchain.SendTxRequest{
		Address: address,
		Payload: request, // 직렬화된 데이터를 Payload 필드에 직접 할당
	}

	response, err := client.SendTx(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to send transaction to node %s: %v", address, err)
	}
	if len(response.Response) > 0 {
	} else {
		return "error"
	}
	return "pass"
}

/*
	sendData(address, request)
*/
