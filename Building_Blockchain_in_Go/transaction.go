package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"

	"fmt"
	"log"
)

const subsidy = 10

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// Coinbase 즉 GenessisBlock인지 아닌지 판명
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// 새로운 코인베이스, 21,000개까지의 채굴
func NewCoinbaseTX(to, data string) *Transaction { //누구에게, Data를 받는다
	if data == "" {
		randData := make([]byte, 20) //20바이트의 랜덤 데이터
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData) //data = randerData
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}            //ID 배열 생성, BTC = -1, 개인 키 = 랜덤
	txout := NewTXOutput(subsidy, to)                           //BTC값 = 10, 공개 키 =to
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}} // ID는 없고, 인풋, 아웃풋 생성
	tx.ID = tx.Hash()                                           //다 합쳐서 Hash로 계산
	return &tx
}

// TXID를 주기 위한 코드
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// 직렬화를 위한 코드
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}
