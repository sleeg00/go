package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int) { //메시지 보내기  (Output작성)
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}
	bc := NewBlockchain(from) //블록체인에서  DB를 뒤져 마지막 블록 Hash를 가져옵니다
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, &UTXOSet) //TX를 기록하고 그 기록을 가져옵니다.
	cbTx := NewCoinbaseTX(from, "")
	txs := []*Transaction{cbTx, tx}

	newBlock := bc.MineBlock(txs) //새로운 블럭이 채굴되면 UTXO가 갱신된다 => TX가 블록이 생기면서 반영되기 때문이다
	UTXOSet.Update(newBlock)      //새로운 블럭의 TX를 뒤져서 UTXO DB를 Update
	fmt.Println("Success!")
	/*
		bc.MineBlock([]*Transaction{tx}) //코인은 전송한다는 건 TX를 만들고 블록 채굴을 통해 이를 블록체인에 추가한다는 것.
		//Tansaction이 생기고 비트코인을 채굴하면 그 Transaction을 기록한다. 평균 10분간격이지만 보상이 지금은 낮다.
		fmt.Println("Success!") //성공 돈을 보냈고 돈은 받은놈이 채굴을 했다.
	*/
}
