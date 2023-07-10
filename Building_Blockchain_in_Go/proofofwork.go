package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

const targetBits = 24 //난이도

const maxNonce = 2147483647

type ProofOfWork struct {
	block  *Block   //블록 하나
	target *big.Int //요구사항의 다른 말 target
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits)) //시프트연산

	pow := &ProofOfWork{b, target} //구조체 생성
	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join( //바이트 배열을 하나로 결합
		[][]byte{
			pow.block.PrevBlockHash,       //현재 블록 이전 Hash Value
			pow.block.Data,                //Block Data
			IntToHex(pow.block.Timestamp), //TimeStamp of Binary
			IntToHex(int64(targetBits)),   // targetBits of Binary
			IntToHex(int64(nonce)),        // nonce of Binary
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)

	for nonce < maxNonce {
		data := pow.prepareData(nonce) //데이터를 준비함 nonce를 이용해서 엄청 긴 데이터를
		hash = sha256.Sum256(data)     //hash값 계산
		fmt.Printf("\r%x", hash)       //찍고

		hashInt.SetBytes(hash[:])          //hash값을 hashInt에 저장
		if hashInt.Cmp(pow.target) == -1 { //해시 값이 목표값 보다 작으면 break!
			break
		} else { //아니면 nonce를 더해서 더 찾기
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:] //nonce와 hash return
}
