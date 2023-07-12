package main

/*
type TXOutput struct {
	Value        int
	ScriptPubKey string
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

/*
	type TXOutput struct {
		Value        int    // BTC값
		PubKeyHash   []byte //공개 키
		ScriptPubKey string
	}

// 새로운 TXOutput 생성자

	func NewTXOutput(value int, address string) *TXOutput {
		txo := &TXOutput{value, nil, "0"} //Value랑 누구에게 보내는지 받는다
		//txo.Lock([]byte(address))

		return txo
	}

type TXOutputs struct { //TX_Output을 구조체로서 가져옴

		Outputs []TXOutput
	}

	func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
		return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0 //두개가 비교해서 같다면 공개키가 같다면 true리턴
	}

func DeserializeOutputs(data []byte) TXOutputs { //역직렬화 배열 아님 Key-Value rnwh

		var outputs TXOutputs

		dec := gob.NewDecoder(bytes.NewReader(data))
		err := dec.Decode(&outputs)
		if err != nil {
			log.Panic(err)
		}

		return outputs
	}

// Serialize serializes TXOutputs

	func (outs TXOutputs) Serialize() []byte {
		var buff bytes.Buffer

		enc := gob.NewEncoder(&buff)
		err := enc.Encode(outs)
		if err != nil {
			log.Panic(err)
		}

		return buff.Bytes()
	}
*/
