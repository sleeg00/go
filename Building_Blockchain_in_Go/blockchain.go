package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket)) //버킷 가져옴
		blockInDb := b.Get(block.Hash)       //블록의 해시값

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()      //직렬화 디코딩
		err := b.Put(block.Hash, blockData) //해시값 넣기
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))               //마지막 블럭 해시 가져오기
		lastBlockData := b.Get(lastHash)             //블록 데이터 가져오기
		lastBlock := DeserializeBlock(lastBlockData) //역직렬화를 통해 가져오기

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func dbExists(dbFile string) bool { //DB열려있는지 확인
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// 새로운 블록체인과 DB를 생성한다.
func CreateBlockchain(address, nodeID string) *Blockchain { //주소와 노드ID를 받는다
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) { //예외처리
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData) //주소, 처음의 데이터
	genesis := NewGenesisBlock(cbtx)                    //새로운 GenesisBlock 생성

	db, err := bolt.Open(dbFile, 0600, nil) //DB Open
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket)) //버킷 생성 스키마?
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize()) //처음 생성된 Genesis Hash 넣기 인코딩해서
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash) //l값에 해시 넣기  (마지막 블럭이라는 뜻)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash //tip = hash값

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db} //마지막 Hash값 , db계속 연결

	return &bc
}
func NewBlockchain(nodeID string) *Blockchain { //새로운 블록 체인
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) { //예외처리
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// 모든 UTXO를 찾고 쓴 UTXO는 지우고 Return
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)  //map[string][TXOutputs]
	spentTXOs := make(map[string][]int) //map[string][int]
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID) //Block의 TXID를 가져옴

		Outputs:
			for outIdx, out := range tx.Vout { //TXoutput[]
				// Was the output spent?
				if spentTXOs[txID] != nil { //썻다면
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false { //Genesis Block이 아니면
				for _, in := range tx.Vin { //Input만큼 일단 여기서 내가 Input했다는 것 -> 무언가를 주문했다는 것
					inTxID := hex.EncodeToString(in.Txid)                  //Input의 TXID
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout) //spentTXOs[inTXID] = 인풋의 TXID 에 입력에서 입력한 값을 더한다
					//이차원 배열이다..
				}
			}
		}

		if len(block.PrevBlockHash) == 0 { //이전 블록이 0 이면 오류 GenesisBlock이거나
			break
		}
	}

	return UTXO
}
