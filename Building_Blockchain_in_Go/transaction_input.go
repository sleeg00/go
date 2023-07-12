package main

type TXInput struct {
	Ixid      []byte
	Vout      int
	ScriptSig string
}

/*
// TX Input

	type TXInput struct {
		Txid      []byte
		Vout      int
		Signature []byte
		PubKey    []byte
		ScriptSig string
	}

// UsesKey checks whether the address initiated the transaction

	func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
		lockingHash := HashPubKey(in.PubKey)

		return bytes.Compare(lockingHash, pubKeyHash) == 0
	}
*/
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}
