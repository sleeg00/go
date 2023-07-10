package main

import "fmt"

type Blockchain struct {
	blocks []*Block
}

func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]   //이전 블록 변수에 추가
	newBlock := NewBlock(data, prevBlock.Hash) //새로운 블록 추가
	bc.blocks = append(bc.blocks, newBlock)    //Blocks에 추가
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{}) //첫 블록
}

func NewBlockchain() *Blockchain { //첫 블록체인
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}

func main() {
	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	for _, block := range bc.blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
