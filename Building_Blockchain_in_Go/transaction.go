/*package main

import (
	"bytes"
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
} /*
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

*/
// Verify verifies signatures of Transaction inputs

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
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

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// SetID sets ID of a transaction
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

// TXInput represents a transaction input
type TXInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

// TXOutput represents a transaction output
type TXOutput struct {
	Value        int
	ScriptPubKey string
}

// CanUnlockOutputWith checks whether the address initiated the transaction
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// CanBeUnlockedWith checks if the output can be unlocked with the provided data
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction { //시작 비트코인
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to} //여기서 address => to 이다
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount) //사용할 UTXO값, UTXO덩어리들의 Idx가 왔다(Output_Idx)

	if acc < amount { //돈이 충분치 않을 경우 error
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid) //txid를 디코딩해서 가져오고
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs { //Output_IDX => UTXO의 IDX
			input := TXInput{txID, out, from} //Input구조체에 기록한다 (어떤 TXID에게 받았고, 얼마를, 누구에게 받았는지)
			inputs = append(inputs, input)    //배열에 저장
		}
	}

	// Build a list of outputs
	outputs = append(outputs, TXOutput{amount, to}) //기록한다 얼마를 누구한테 줌
	if acc > amount {                               //거스름돈을 챙기자
		outputs = append(outputs, TXOutput{acc - amount, from}) //나한테 뺀 가격만큼 거스름돈을 보내자
	}

	tx := Transaction{nil, inputs, outputs} //새로운 트랜잭션을 생성하자 ! Input, Output을 넣고
	tx.SetID()

	return &tx
}
