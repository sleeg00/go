package main

import (
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block) //pow = 생성된 블록 , 시피트 연산한 값을 가져옴
	nonce, hash := pow.Run()     // 작업 증명

	block.Hash = hash[:] // 작업 증명에 통과한 Hash VAlue 저장
	block.Nonce = nonce  // 작업 증명에 사용한 nonce 저장

	return block
}
