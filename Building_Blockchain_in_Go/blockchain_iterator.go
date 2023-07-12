package main

/*
import (
	"log"

	"github.com/boltdb/bolt"
)

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB //키들은 바이트 순서로 정렬되어 있따
}

func (bc *Blockchain) Iterator() *BlockchainIterator { //반복자
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}
func (i *BlockchainIterator) Next() *Block {
	var block *Block //블럭 구조체 변수 생성

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)   //i의 현재Hash
		block = DeserializeBlock(encodedBlock) //역직렬화

		return nil
	})
	if err != nil {
		log.Print(err) //Panic = 에러메시지 출력 후 즉시중단
	}
	i.currentHash = block.PrevBlockHash //이전 블럭 Hash가져옴 => 반복문이라고 볼 수 있다.
	return block
}
*/
