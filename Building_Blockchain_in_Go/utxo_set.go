package main

import (
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

type UTXOSet struct { //UTXO 집합은 Blockchain과 결합되어있다.
	Blockchain *Blockchain
}

// UTXO 세트를 재구성하는 메소드
func (u UTXOSet) Reindex() {
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error { //버킷이 열려있으면 닫는다
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO() //블록체인의 모든 UTXO들을 가져온다. //TXID-TXOUTPUT 의 맵을 가져온다

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

// UTXO의 집합들에서 사용할 수 있는 금액들을 가져온다.
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) { //공개키, 계좌 총 양을 받음

	unspentOutputs := make(map[string][]int) //Key: Value로 정의 (UTXO)
	accumulated := 0                         //일단은 0
	db := u.Blockchain.db                    //db가져옴 //UTXO들의 집합 DB이다.

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket)) //버킷 가져옴
		c := b.Cursor()                    // 현재 버킷 선택

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k) //TXID가져오기, Key로 사용될 것임
			outs := DeserializeOutputs(v) //역직렬화하기 사용할 수 있도록 배열로서 가져옴 TX_output을 UTXO들이 존재한다

			for outIdx, out := range outs.Outputs { //어떤 UTXO인지 Idx에 저장, UTXO는 out에저장
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount { //공캐키가 맞고 계좌 잔액이 0 BTC보다 크다면
					//조건이 중요하다 지금은 UTXO를 사용하려고 ex) 원하는 지불 값 5천원 UTXO가 >5,000보다 크면 종료
					//그것을 지불하기 위해서 지금 모으는 중이다
					accumulated += out.Value                                    //계좌에 출력 금액을 더해준다
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx) //배열에 TXID: TX_output의 Index 몇번째인데 추가
					//UTXO[TransactionId] += outIdx -> UTXOIndex를 더한다.
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs //총 UTXO => 즉 작액, 계좌들의 [Key,Value] = [TransactionId, UTXOIndex]
}

// UTXO가져오기 ex) 잔액조회 등등..
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput { //공개 키

	var UTXOs []TXOutput  //UTXOs = > UTXO의 배열
	db := u.Blockchain.db //db연결

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v) //

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) { //나의 잔고를 가져온다.
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// 몇개의 트랜잭션이 있는지 반환

func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return counter
}

// 새로운 블럭을 받아
func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.db //UTXO DB

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, vin := range tx.Vin { //입력을 가져온다
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid) //Input전 Output이 발생한 위치를 찾는다 UTXO를 찾는다

					outs := DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout { //내가 받은것을 다시 쓰지 않았다면 즉 UTXO가 남아있다면
							updatedOuts.Outputs = append(updatedOuts.Outputs, out) //UTXO업데이트 지갑에 있는 구조체들 더함
						}
					}

					if len(updatedOuts.Outputs) == 0 { //비었따면
						err := b.Delete(vin.Txid) //방금 트랜잭션에서 사용했다는 것이므로 TXID자체를 지운다 하나도 남은 게 없기때문 그 계좌 지움
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.Txid, updatedOuts.Serialize()) //남아있는 잔액으로만 업데이트한다.
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.Vout { //CoinBase일때 애기다 이건
				newOutputs.Outputs = append(newOutputs.Outputs, out) //새로운 Output을 더한다.
			}

			err := b.Put(tx.ID, newOutputs.Serialize()) //
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
