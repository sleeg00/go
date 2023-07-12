package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"math/big"

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

/*
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

		txin := TXInput{[]byte{}, -1, nil, []byte(data)}            //ID 배열 생성, Idx = -1, Signature = -nil, pubkey = []
		txout := NewTXOutput(subsidy, to)                           //BTC값 = 10, 공개 키 =to -> 누구에게 준 것이냐
		tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}} // ID는 없고, Input 0 , Output 0 생성 (누구에게 줄 것인지)
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

// NewUTXOTransaction creates a new transaction

	func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
		var inputs []TXInput
		var outputs []TXOutput

		pubKeyHash := HashPubKey(wallet.PublicKey)
		acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

		if acc < amount {
			log.Panic("ERROR: Not enough funds")
		}

		// Build a list of inputs
		for txid, outs := range validOutputs {
			txID, err := hex.DecodeString(txid)
			if err != nil {
				log.Panic(err)
			}

			for _, out := range outs {
				input := TXInput{txID, out, nil, wallet.PublicKey}
				inputs = append(inputs, input)
			}
		}

		// Build a list of outputs
		from := fmt.Sprintf("%s", wallet.GetAddress())
		outputs = append(outputs, *NewTXOutput(amount, to))
		if acc > amount {
			outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
		}

		tx := Transaction{nil, inputs, outputs}
		tx.ID = tx.Hash()
		UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

		return &tx
	}

// DeserializeTransaction deserializes a transaction
*/
func (tx Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())

	tx.ID = hash[:]
}

// GetHash hashes the transaction and returns the hash
func (tx Transaction) GetHash() []byte {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())

	return hash[:]
}
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TXInput([]byte{}, -1, data)
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()
	return &tx
}
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Vin[inID].PubKey = nil
	}

	return true
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}
