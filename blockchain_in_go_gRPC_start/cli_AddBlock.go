package main

import (
	"context"
	"log"

	"github.com/sleeg00/blockchain_go/proto"
	blockchain "github.com/sleeg00/blockchain_go/proto"
	"google.golang.org/grpc"
)

func AddBlockRequest(transactions []*Transaction, node_id string) bool {

	conn, err := grpc.Dial("localhost:"+node_id, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial node %s: %v", node_id, err)
	}
	defer conn.Close()
	client := proto.NewBlockchainServiceClient(conn)

	var protoTransactions []*blockchain.Transaction

	for _, tx := range transactions {
		protoTx := &blockchain.Transaction{
			Id:   tx.ID,
			Vin:  []*blockchain.TXInput{},
			Vout: []*blockchain.TXOutput{},
		}

		for _, input := range tx.Vin {
			protoInput := &blockchain.TXInput{
				Txid:      input.Txid,
				Vout:      input.Vout,
				Signature: input.Signature,
				PubKey:    input.PubKey,
			}
			protoTx.Vin = append(protoTx.Vin, protoInput)
		}

		for _, output := range tx.Vout {
			protoOutput := &blockchain.TXOutput{
				Value:      output.Value,
				PubKeyHash: output.PubKeyHash,
			}
			protoTx.Vout = append(protoTx.Vout, protoOutput)
		}

		protoTransactions = append(protoTransactions, protoTx)
	}

	req := &blockchain.AddBlockRequest{
		Transactions: protoTransactions,
		NodeId:       node_id, // 직렬화된 데이터를 Payload 필드에 직접 할당
	}
	response, err := client.AddBlock(context.Background(), req)
	log.Println(response)
	return true
}
