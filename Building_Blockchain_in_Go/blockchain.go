/*
package main

import (

	"bytes"
	"encoding/hex"
	"errors"
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
/*

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

		cbtx := NewCoinbaseTX(address, genesisCoinbaseData) //주소, 처음의 데이터 input 0, output 0 생성
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
*/
package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// 채굴 작업
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l")) //마지막 블럭의 해시값

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash) //TX, 마지막 해쉬(이전 블럭으로 설정) 새로운 블럭 생ㅅ어

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize()) //새로운 블럭 해쉬 ,직렬화 값
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash) //마지막 블럭 교체
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash //마지막 블럭 교체

		return nil
	})
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil { //쓴적이 있다면 그러니까 받은것을 누군가에게 이체를 조금이라도 했다면
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx { //검사하자 다 Input으로 받은것을 전부 다 Output했는지
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) { //키가 있냐(지금은 사용자 주소로만 한다)?
					unspentTXs = append(unspentTXs, *tx) //TX~~에 남은 돈이 있어요 -> UTXO가 있어요 알려준다
				}
			}

			if tx.IsCoinbase() == false { //Genesis Block이면
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// 사용가능한 잔액, 즉 UTXO를 찾자
// 사용할 UTXO값과 UTXO덩어리들(Output_Idx)를 반환한다.
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	//만약 Ivan에게 0
	unspentOutputs := make(map[string][]int)          //UTXO
	unspentTXs := bc.FindUnspentTransactions(address) //address에 그러니까 A가 B에게 돈을 보낼 때 A의 UTXO가 남아있는지 조사해서 TX를 넘김
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID) //인코딩한다

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value                                    //UTXO를 더한다 A->B에게 보낼보다 크다면 멈춘다
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx) //UTXO[TXID] = Output의 Idx를 저장한다.
				//해당 UTXO의 덩어리들을 append! 왜냐하면 UTXO덩어리를 두 개 쓸 수 도 있기 때문이다.

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs // 사용할 UTXO값과 UTXO덩어리들(Output_Idx)를 보낸다
}

// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l")) //마지막 블록의 HAsh를 가져오기

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// 말그대로 블록체인을 생성한다 Genesis 블록을 생성하고 Output 생성 (Genesis, Output, LastHash-> 블록 연결점 생성)
func CreateBlockchain(address string) *Blockchain {
	if dbExists() { //DB예외처리
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData) //GenesisCoinbaseData로 첫 채굴할 것 생성
		genesis := NewGenesisBlock(cbtx)                    //Transaction 내용은 cbtx로  output만 10BTC생성함

		b, err := tx.CreateBucket([]byte(blocksBucket)) //스키마 생성
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize()) // 버킷에 Gensis블록의 해시, 직렬화해서 값을 넣음
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash) //마지막 블록은 l =GEnesis블록 해시값이라고 명시해줌
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db} //GenesisHash와 ,DBd연결값 넘김

	return &bc
}
