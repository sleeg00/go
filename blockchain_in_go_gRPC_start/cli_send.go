package main

import (
	"context"
	"fmt"
	"log"

	blockchain "github.com/sleeg00/blockchain_go/proto"
)

func (cli *CLI) send(from, to string, amount int, node_id string, mineNow bool) {
	request := &blockchain.SendRequest{ //클라이언트가 sendRequest호출
		From:    from,
		To:      to,
		Amount:  int32(amount),
		NodeId:  node_id,
		MineNow: mineNow,
	}
	//누가, 누구에게, 얼마를, 어떤 노드 포트번호에게, 마이닝을 하고 있는지 보냄

	response, err := cli.blockchain.Send(context.Background(), request)
	if err != nil {
		log.Fatalf("Error calling Send RPC: %v", err)
	}

	if response.Response == "Success" {
		fmt.Println("Success!")
	} else {
		fmt.Println("Transaction failed.")
	}
}

/*
func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		cbTx := NewCoinbaseTX(from, "")
		txs := []*Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}

*/
