package main

import "bytes"

// TXInput
type TXInput struct {
	Txid      []byte //Trasaction ID( before )
	Vout      int    //Output Idx
	Signature []byte //서명
	PubKey    []byte //공개키
}

// 공개키가 맞는지 확인하는 용도 같다 ? 입력이 출력을 해체할 수 있는 특정한 키를 가지고 있는지
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
