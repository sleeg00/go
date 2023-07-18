package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"github.com/boltdb/bolt"
	"github.com/sleeg00/blockchain_go/proto"
	blockchain "github.com/sleeg00/blockchain_go/proto"
	"google.golang.org/grpc"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type server struct {
	blockchain blockchain.BlockchainServiceServer
}

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type verzion struct {
	Version    int
	BestHeight int64
	AddrFrom   string
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

// StartServer starts a node

func (s *server) Version(ctx context.Context, req *proto.VersionRequest) (*proto.VersionResponse, error) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(req.Payload[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	bc, err := GetBlockchain(req.Address)

	myBestHeight := bc.GetBestHeight()

	return &proto.VersionResponse{
		NodeId: req.Address,
		Height: int32(myBestHeight),
	}, nil
}

func StartServer2(nodeID, minerAddress string) {
	LocalNode := "localhost:" + nodeID
	srv := grpc.NewServer()
	lis, err := net.Listen("tcp", fmt.Sprintf(LocalNode))
	if err != nil {
		log.Fatalf("서버 연결 안됨")
	}
	blockchainService := &server{}
	blockchain.RegisterBlockchainServiceServer(srv, blockchainService)

	log.Println("Server listening on localhost:", nodeID)

	srv.Serve(lis)

}

// mempool에서 TXID지움
func (s *server) DeleteTxInMempool(ctx context.Context, req *proto.DeleteTxInMempoolRequest) (*proto.DeleteTxInMempoolResponse, error) {
	delete(mempool, req.TxId)
	return &proto.DeleteTxInMempoolResponse{}, nil
}
func (s *server) CreateBlockchain(ctx context.Context, req *proto.CreateBlockchainRequest) (*proto.CreateBlockchainResponse, error) {
	if !ValidateAddress(req.Address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := CreateBlockchain(req.Address, req.NodeId)
	defer bc.db.Close()

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	return &proto.CreateBlockchainResponse{
		Response: "Success",
	}, nil
}
func (s *server) CreateWallet(ctx context.Context, req *proto.CreateWalletRequest) (*proto.CreateWalletResponse, error) {
	log.Println("!@#@")
	wallets, _ := NewWallets(req.NodeId)
	log.Println("!")
	address := wallets.CreateWallet()
	log.Println("2")
	wallets.SaveToFile(req.NodeId)
	log.Println("3")

	fmt.Printf("Your new address: %s\n", address)
	return &proto.CreateWalletResponse{
		Address: address,
	}, nil
}
func (s *server) GetBalance(ctx context.Context, req *proto.GetBalanceRequest) (*proto.GetBalanceResponse, error) {
	// CreateBlockchain 메소드의 구현 내용을 작성하세요
	return nil, nil
}
func (s *server) ListAddresses(ctx context.Context, req *proto.ListAddressesRequest) (*proto.ListAddressesResponse, error) {
	// CreateBlockchain 메소드의 구현 내용을 작성하세요
	return nil, nil
}
func (s *server) ReindexUTXO(ctx context.Context, req *proto.ReindexUTXORequest) (*proto.ReindexUTXOResponse, error) {
	// CreateBlockchain 메소드의 구현 내용을 작성하세요
	return nil, nil
}
func (s *server) PrintChain(ctx context.Context, req *proto.PrintChainRequest) (*proto.PrintChainResponse, error) {
	// CreateBlockchain 메소드의 구현 내용을 작성하세요
	return nil, nil
}
func (s *server) Send(ctx context.Context, req *proto.SendRequest) (*proto.SendResponse, error) {

	if !ValidateAddress(req.From) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(req.To) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	//지갑 정보를 확인한다 노드의 지갑이 존재하는지
	wallets, err := NewWallets(req.NodeId) //wallet_NodeId가 존재하는지 확인
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(req.From) //해당 address가 있는지 확인

	//NodeID의 블록체인을 가져온다.
	bc, err := GetBlockchain(req.NodeId)
	if err != nil {
		log.Println("못 가져오네")
	}
	//블록체인 동기화 한 번 해주자 -> 할 필요 없다 Mempool에서 TX가 다 차고 블럭이 생성되는 순간 어자피 동기화를 하기 때문이고,
	//Mempool이 다 안차있는데 블럭이 새로 생기는 경우가 없기 때문이다.
	UTXOSet := UTXOSet{bc}
	//UTXO를 찾아 있으면 TX생성해서 Output만듬 -> 여기서 서명도 했음
	tx := NewUTXOTransaction(&wallet, req.To, int64(req.Amount), &UTXOSet)
	//순서대로 from, to, amount, UTXOset이다.

	//마이닝하고 있는 사람이 없으므로 바로 TX 전송과 동시에 마이닝한다.
	//Genesis Block에서 마이닝한 사람들이기 때문에 Input, OUTPUT의 TX는 임의로 정해둔다
	if req.MineNow {
		log.Println("바로 마이닝합니다.")
		cbTx := NewCoinbaseTX(req.From, "")
		txs := []*Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		for _, node := range knownNodes {
			if node != req.NodeId {
				AddBlockRequest(txs, node)
			}
		}
		//newBlock을 전파하자 모든 노드에게

		UTXOSet.Update(newBlock)
		//여기서 모든 블럭이 생겼다고 알려줘야 한다. UTXO는 따로 DB가 존재
	} else { //첫 번째 실패
		//마이낭하고 있는 사람이 없다면 모든 노드에게 TX를 보내서 일관성을 유지하자.
		for node := 0; node < len(knownNodes); node++ { //마이닝을 찾았다면 종료해야한다.
			//굳이 클라이언트에 반환해서 하지 않는 이유는 일관성 유지, 병목현상 제거이다.
			log.Println("node : ", knownNodes[node])
			bc.db.Close()
			if sendTxToNode(knownNodes[node], tx) == "error" {
				break
			}
		}

	}

	defer bc.db.Close()

	fmt.Println("Success!")
	return &proto.SendResponse{Response: "Success"}, nil
}
func (s *server) AddBlock(ctx context.Context, req *proto.AddBlockRequest) (*proto.AddBlockResponse, error) {
	bc, err := GetBlockchain(req.NodeId)
	var lastHash []byte
	var lastHeight int64

	err = bc.db.View(func(tx *bolt.Tx) error {
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

	var transactions []*Transaction

	for _, tx := range req.Transactions {
		var vin []TXInput
		for _, input := range tx.Vin {
			vin = append(vin, TXInput{
				Txid:      input.Txid,
				Vout:      input.Vout,
				Signature: input.Signature,
				PubKey:    input.PubKey,
			})
		}

		var vout []TXOutput
		for _, output := range tx.Vout {
			vout = append(vout, TXOutput{
				Value:      output.Value,
				PubKeyHash: output.PubKeyHash,
			})
		}

		transactions = append(transactions, &Transaction{
			ID:   tx.Id,
			Vin:  vin,
			Vout: vout,
		})
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

	return &proto.AddBlockResponse{}, nil

}
func (s *server) SendTx(ctx context.Context, req *proto.SendTxRequest) (*proto.SendTxResponse, error) {
	log.Println("SendTx_gRPC_Server")
	var buff bytes.Buffer
	var payload tx

	buff.Write(req.Payload[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Payload를 읽었습니다 ")
	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx // 서버의 mempool에 기입

	bc, err := GetBlockchain(req.Address) //내 블록체인
	if err != nil {
		log.Println("블록체인 가져오기 실패")
	}
	log.Println("블록체인을 읽었습니다")
	//모든 노드에게 요청을 보내고 있기 때문에 나의 노드에만 더하면 된다.
	log.Println("mempool을 출력해봅니다.", mempool)
	//어라? 마이닝하고 있는 사람이 있다면
	if len(mempool) >= 2 && len(miningAddress) > 0 {
	MineTransactions:
		var txs []*Transaction

		for id := range mempool { //mempool에 있는 것 txs에 다 넣음.
			tx := mempool[id]
			if bc.VerifyTransaction(&tx) {
				txs = append(txs, &tx)
			}
		}
		log.Println("mempool을 출력해봅니다.", mempool)
		if len(txs) == 0 {
			fmt.Println("All transactions are invalid! Waiting for new ones...")

		} else {
			cbTx := NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs) //마이닝 블럭 추가 동기화
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()
			fmt.Println(newBlock)
			fmt.Println("New block is mined!")
			//블록 생성됨

			for _, tx := range txs { //그동안 했던 것 지우고 -> 이것도 gRPC 호출 해야 함 서버마다
				log.Println("그동안 추가했던 mempool을 지웁니다.")
				txID := hex.EncodeToString(tx.ID)
				for _, no := range knownNodes {
					conn, err := grpc.Dial("localhost:"+no, grpc.WithInsecure())
					if err != nil {
						log.Fatalf("Failed to dial node %s: %v", "localhost:"+no, err)
					}
					defer conn.Close()
					client := proto.NewBlockchainServiceClient(conn)
					req := &blockchain.DeleteTxInMempoolRequest{
						Address: no,
						TxId:    txID, // 직렬화된 데이터를 Payload 필드에 직접 할당
					}
					_, err = client.DeleteTxInMempool(context.Background(), req)
					if err != nil {
						log.Fatalf("Failed to DELETE TX IN MEMPOOL %s: %v", no, err)
					}
				}
			}
			return &proto.SendTxResponse{}, nil

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
	return &proto.SendTxResponse{Response: "Pass"}, nil
}

func (s *server) Addr(ctx context.Context, req *proto.AddrRequest) (*proto.AddrResponse, error) {
	return &proto.AddrResponse{}, nil
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
