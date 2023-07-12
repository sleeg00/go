package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const subsidy = 10

// TX 구조체
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// 첫 코인인지
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// 직렬화
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// TX에 서명을 추가하는 메소드
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() { //Genesis 아니면 진행
		return
	}

	for _, vin := range tx.Vin { //그전 입력값들이 정상적인 입력들인지 확인
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy() //입력의 데이터가 포함된 잘려진 복사본

	for inID, vin := range txCopy.Vin { //Index , Input[Index]
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]            //Input의 참조갑 Output을 가져옵니다.
		txCopy.Vin[inID].Signature = nil                           //복사본 사인 X
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash //복사본 공개 키 = 이전 출력의 복사 키 해시 값
		txCopy.ID = txCopy.Hash()                                  //복사본의 ID를 해시를 얻고
		txCopy.Vin[inID].PubKey = nil                              //복사본의 공캐키는 null로바꿈

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID) //privKey로 txCopy.ID를 서명한다 .
		//A->B에게 보낸다고 생각하면 Output(금액 , 비밀 키 ) -> Input(누가, 공개 키 )-> 전자서명이다
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...) //signarute필드에 여러가지 조합해서 저장한다

		tx.Vin[inID].Signature = signature
	}
}

// 입력의 데이터가 포함된 잘려진 복사 복사
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin { //Input들을 가져옵니다.
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout { //Output들을 가져옵니다.
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs} //트랜잭션 자체를 복사합니다.

	return txCopy
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

// 서명을 검증
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy() //복사
	curve := elliptic.P256()   //키 쌍 생성할 때 사용된 것과 동일한 곡선

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil
		//----------- 서명 데이터를 가져오는 작업 -------------

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

		rawPubKey := ecdsa.PublicKey{curve, &x, &y} //그냥 검증이라고 생각하고 넘기자
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

// 새로운 코인베이스, 21,000개까지의 채굴
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

// Input, Output , TX기록 TX반납
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

	tx := Transaction{nil, inputs, outputs} //새로운 트랜잭션을 생성하자 ! Input, Output을 넣고 거스름돈이라면 누구에게 9원 줌 기록
	tx.SetID()

	return &tx
}
