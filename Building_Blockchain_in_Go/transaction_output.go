package main

import (
	"bytes"
)

// TxOutput
type TXOutput struct {
	Value      int    //값
	PubKeyHash []byte //공개키 HAsh
}

// 서명을 잠군다.
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0 //두개가 비교해서 같다면 공개키가 같다면 true리턴
}

// 새로운 Output을 만든다.
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}
