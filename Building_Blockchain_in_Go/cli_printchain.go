package main

import (
	"fmt"
	"strconv"
)

func (cli *CLI) printChain() { //Print BlockChain
	// TODO: Fix this
	bc := NewBlockchain("")
	defer bc.db.Close()

	bci := bc.Iterator() //최근 거래내역 부터

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)         //이전 블록
		fmt.Printf("Hash: %x\n", block.Hash)                        //블록 해쉬
		pow := NewProofOfWork(block)                                //작업증명 구조체 (블록, 타겟값)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate())) //이것이 존재하는 블록 즉 유효한 블럭인지 판별
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
