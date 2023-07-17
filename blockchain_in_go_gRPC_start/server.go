package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/sleeg00/blockchain_go/proto"
	blockchain "github.com/sleeg00/blockchain_go/proto"
	"google.golang.org/grpc"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
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
	BestHeight int
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

func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}

func sendAddr(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	sendData(address, request)
}

func sendBlock(addr string, b *Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)
}

func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(address, request)
}

func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

func sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})

	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
}

func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	requestBlocks()
}

func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	fmt.Println("Recevied a new block!")
	bc.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

func handleInv(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func handleGetBlocks(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

func handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]
		log.Println(tx)
		//sendTx(payload.AddrFrom, &tx)
		// delete(mempool, txID)
	}
}

func handleTx(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction

			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbTx := NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()

			fmt.Println("New block is mined!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func handleVersion(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	sendAddr(payload.AddrFrom)
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

func handleConnection(conn net.Conn, bc *Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

// StartServer starts a node

func StartServer2(nodeID, minerAddress string) {
	srv := grpc.NewServer()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", nodeID))
	if err != nil {
		log.Panic(err)
	}
	bc := NewBlockchain(nodeID) // 풀 노드 가져옴
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}
	blockchainService := &server{}
	blockchain.RegisterBlockchainServiceServer(srv, blockchainService)

	log.Println(fmt.Sprintf("Server listening on localhost:%s", nodeID))

	srv.Serve(lis)
}

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
	wallets, _ := NewWallets(req.NodeId)
	address := wallets.CreateWallet()
	wallets.SaveToFile(req.NodeId)

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
	wallets, err := NewWallets(req.NodeId)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(req.From)

	//NodeID의 블록체인을 가져온다.
	bc := NewBlockchain(req.NodeId)

	//블록체인 동기화 한 번 해주자 -> 할 필요 없다 Mempool에서 TX가 다 차고 블럭이 생성되는 순간 어자피 동기화를 하기 때문이고,
	//Mempool이 다 안차있는데 블럭이 새로 생기는 경우가 없기 때문이다.
	UTXOSet := UTXOSet{bc}
	//UTXO를 찾아 있으면 TX생성해서 Output만듬 -> 여기서 서명도 했음
	tx := NewUTXOTransaction(&wallet, req.To, int(req.Amount), &UTXOSet)
	//순서대로 from, to, amount, UTXOset이다.

	//마이닝하고 있는 사람이 없으므로 바로 TX 전송과 동시에 마이닝한다.
	//Genesis Block에서 마이닝한 사람들이기 때문에 Input, OUTPUT의 TX는 임의로 정해둔다
	if req.MineNow {
		cbTx := NewCoinbaseTX(req.From, "")
		txs := []*Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		//마이낭하고 있는 사람이 없다면 모든 노드에게 TX를 보내서 일관성을 유지하자.
		for node := 0; node < len(knownNodes); node++ { //마이닝을 찾았다면 종료해야한다.
			//굳이 클라이언트에 반환해서 하지 않는 이유는 일관성 유지, 병목현상 제거이다.

			if sendTxToNode(knownNodes[node], tx) == "error" {
				break
			}
		}

	}

	defer bc.db.Close()

	fmt.Println("Success!")
	return &proto.SendResponse{}, nil
}
func (s *server) SendTx(ctx context.Context, req *proto.SendTxRequest) (*proto.SendTxResponse, error) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(req.Payload[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx // 서버의 mempool에 기입

	bc := NewBlockchain(req.Address) //내 블록체인

	//모든 노드에게 요청을 보내고 있기 때문에 나의 노드에만 더하면 된다.

	//어라? 마이닝하고 있는 사람이 있다면
	if len(mempool) >= 2 && len(miningAddress) > 0 {
	MineTransactions:
		var txs []*Transaction

		for id := range mempool {
			tx := mempool[id]
			if bc.VerifyTransaction(&tx) {
				txs = append(txs, &tx)
			}
		}

		if len(txs) == 0 {
			fmt.Println("All transactions are invalid! Waiting for new ones...")

		} else {
			cbTx := NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs) //마이닝
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()
			fmt.Println(newBlock)
			fmt.Println("New block is mined!")

			for _, tx := range txs { //그동안 했던 것 지우고 -> 이것도 gRPC 호출 해야 함 서버마다
				txID := hex.EncodeToString(tx.ID)
				for _, no := range knownNodes {
					conn, err := grpc.Dial(no, grpc.WithInsecure())
					if err != nil {
						log.Fatalf("Failed to dial node %s: %v", no, err)
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
			/*
				//Mempool에서 TX를 다 지우고 블럭을 추가한다.
				for _, node := range knownNodes {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			*/

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
	return &proto.SendTxResponse{Response: "Pass"}, nil
}

func (s *server) Version(ctx context.Context, req *proto.VersionRequest) (*proto.VersionResponse, error) {

	return &proto.VersionResponse{}, nil
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
