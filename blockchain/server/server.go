package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/sleeg00/blockchain/protos" // protobuf 파일의 경로에 맞게 수정

	"sync"
)

type Block struct { //Response Dto라고 볼 수 있다 나중에 코드를 나누는 것이 더 좋을 듯
	PrevHash  string //이전 블록의 Hash value
	Hash      string //현재 블록의 Hash value
	Data      string //블록에 저장되는 Data
	Timestamp string //블록이 생성된 시간
}

type BlockChain struct { //위와 같은 맥락이다
	chain  []Block      //Block의 구조체의 배열(말 그대로 생각하자 그럼 편하다), 순차적으로 저장
	lock   sync.Mutex   //동시성 관리를 위한 lock => Mutex
	server *grpc.Server //gRPC서버 인스턴스 통신을 관리하자
}

func (bc *BlockChain) createBlock(prevHash, data string) Block { //블록 생성
	block := Block{ //Block 구조체 생성
		PrevHash:  prevHash,
		Data:      data,
		Timestamp: "2023-07-08 12:00:00",
	}

	block.Hash = bc.calculateHash(block) //인코딩한 Hash값을 저장합니다.
	return block
}

func (bc *BlockChain) calculateHash(block Block) string { //Hash값을 암호화합니다
	record := block.PrevHash + block.Data + block.Timestamp
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

func (bc *BlockChain) addBlock(data string) {
	bc.lock.Lock()         //동시성 관리를 위해 먼저 Lock을 겁니다.
	defer bc.lock.Unlock() //defer = > bc.lock.Unlock()을 지연 실행하도록 설정
	//함수 반환전에 락을 푼다.

	prevBlock := bc.chain[len(bc.chain)-1]           //이전 블록을 가져옵니다
	newBlock := bc.createBlock(prevBlock.Hash, data) //새로운 블록을 생성합니다.
	//여기서 체인처럼 연결합니다 이전 블록은 무엇인지
	//createBlock에서 저장합니다.
	bc.chain = append(bc.chain, newBlock) //chain에 append 더합니다.
}

func (bc *BlockChain) initializeChain() { //블럭 초기화 및 생성
	genesisBlock := bc.createBlock("", "Genesis Block")
	bc.chain = append(bc.chain, genesisBlock)
}
func (bc *BlockChain) isChainValid() bool { //체인이 연결이 되어있고 해쉬값이 맞는지 확인
	for i := 1; i < len(bc.chain); i++ {
		currentBlock := bc.chain[i]
		prevBlock := bc.chain[i-1]

		// Check the hash of the current block
		if currentBlock.Hash != bc.calculateHash(currentBlock) {
			return false
		}

		// Check the previous hash of the current block
		if currentBlock.PrevHash != prevBlock.Hash {
			return false
		}
	}
	return true
}

type BlockChainService struct {
	blockchain *BlockChain
	pb.UnimplementedBlockChainServer
}

// 컨트롤러와 비즈니스 로직의 역할
func (s *BlockChainService) AddTransaction(ctx context.Context, req *pb.TransactionRequest) (*pb.AddTransactionResponse, error) {
	data := req.GetData()       //클라이언트에게 받은 데이터를 가져옵니다
	s.blockchain.addBlock(data) //블록에 데이터를 더해줍니다.

	return &pb.AddTransactionResponse{ //응답으로는 이렇게 전송합니다.
		Message: "Transaction added successfully",
	}, nil
}

func (s *BlockChainService) ValidateChain(ctx context.Context, req *pb.ValidateChainRequest) (*pb.ValidateChainResponse, error) {
	isValid := s.blockchain.isChainValid() //체읹 확인 연결되어있고 해쉬값이 맞는지

	return &pb.ValidateChainResponse{
		IsValid: isValid,
	}, nil
}

func main() {
	blockchain := &BlockChain{}
	blockchain.initializeChain()

	server := grpc.NewServer()
	blockchainService := &BlockChainService{
		blockchain: blockchain,
	}

	pb.RegisterBlockChainServer(server, blockchainService)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	fmt.Println("Server listening on port 50051")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
