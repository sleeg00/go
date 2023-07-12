package main

import (
	"bytes"
	"crypto/ecdsa"
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

/*
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
// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile)
	if dbExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

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

// * UTXO를 포함하는 TX를 찾는 작업
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)  //map[string][TXOutputs]
	spentTXOs := make(map[string][]int) //map[string][int] //소비된 출력의 정보이다
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID) //Block의 TXID를 가져옴

		Outputs:
			for outIdx, out := range tx.Vout { //TXoutput[]   Output기록이 없다는 것은 누군가에게 지불한 적이 없기 때문에 돌지 않는다!
				// Was the output spent?
				if spentTXOs[txID] != nil { //나의 Input을 출력한 적 있는 경우
					//!! 여기서 Idx가 같으면 Continue하는 이유는 BC에서 금액을 지불할때는 받은것을 일단 그대로 보내고 거스름돈 개념으로 Output을
					//추가하기 때문에 Idx가 일단 같다면 Continue하고 밑에 있는 Output_Idx를 추가한다.
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx { //만약 Output Idx 0과 out의 Idx가 0이면 패스한다. ?
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]                       //UTXO[TransactionID] = TXoutput[]이기 때문에 성립
				outs.Outputs = append(outs.Outputs, out) //더한다 Output을 UTXO기 때문에
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false { //Genesis Block이 아니면
				for _, in := range tx.Vin { //Input은 누구에게 받은 것에 대한 정보이다!
					inTxID := hex.EncodeToString(in.Txid) //Input의 TXID = 이전 TxID
					//즉 Output에서 참조하기 때문에 이전 TXID를 참조한다
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout) //그냥 인덱스 뽑기이다.
					//그러니까 spentTXOs[0] = append spentTXos[0] = Output의 0번인 덱스

				}
			}
		}

		if len(block.PrevBlockHash) == 0 { //이전 블록이 0 이면 오류 GenesisBlock이거나
			break
		}
	}

	return UTXO
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}
func main() {
	cli := CLI{}
	cli.Run()
}
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		// TODO: ignore transaction if it's not valid
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}
