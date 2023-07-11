package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block { //TX를 받고, 이전 블록을 받는다, 높이도 받는다
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0, height}
	pow := NewProofOfWork(block) //pow = 생성된 블록 , 시피트 연산한 값을 가져옴
	nonce, hash := pow.Run()     // 작업 증명

	block.Hash = hash[:] // 작업 증명에 통과한 Hash VAlue 저장
	block.Nonce = nonce  // 작업 증명에 사용한 nonce 저장

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block { //TX를 받고 새로운 재네시스 블록을 생성한다.
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0) //TX구조체를 넘기고, 이전 블록은 없어서 그냥 생성자만 넘긴다, height도 0이다.
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
