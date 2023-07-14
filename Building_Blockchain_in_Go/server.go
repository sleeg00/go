package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type verzion struct {
	Version    int //버전을 받은 노드는 자신의 version 메시지로 응답 핸드셰이크 개념
	BestHeight int //노드가 version 메시지를 받으면 자신이 가진 블록체인 BestHeight보다 더 긴지 확인 짧을 경우 누락된 블록을 요청해 다운받는다
	AddrFrom   string
}
type addr struct {
	AddrList []string //주소리스트
}

type block struct {
	AddrFrom string //어떤주소에게
	Block    []byte //블럭
}

type getblocks struct {
	AddrFrom string //어떤주소에게
}

type getdata struct {
	AddrFrom string //어떤주소에게
	Type     string //타입은
	ID       []byte //ID는
}

type inv struct {
	AddrFrom string   //어떤주소에게
	Type     string   //타입은
	Items    [][]byte //어떤 것을?
}

type tx struct {
	AddFrom     string //어떤주소에게
	Transaction []byte //TX
}

// 서버 시작 노드 주소 설정 & 응답이 올때 까지 기다립니다 + 내 노드를 최신화 시킵니다 -> sendVersion
func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID) //노드주소는 localhost와 전달받은 포트로 시작
	miningAddress = minerAddress                      //채굴 보상을 받을 주소
	ln, err := net.Listen(protocol, nodeAddress)      //TCP 서버 가동
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc := NewBlockchain(nodeID) //nodeID로 블록체인 생성

	if nodeAddress != knownNodes[0] { //중앙 노드가 아니면
		sendVersion(knownNodes[0], bc) //블록이 최신화가 되어 있는지 확인하기 위해 중앙 노드에 version 메시지 전송
	}

	for {
		conn, err := ln.Accept() //서버가 요청을 받으면
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc) //새로운 연결을 처리하기 위해 ㄹ고루틴 실행
	}
}

// 중앙노드에게 나의 BC를 보냅니다. addr == 중앙노드
func sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight() //나의 노드 BC의 최고길이를 가져온다
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
	//1과 최고길이, 중앙노드를, 나의 주소를 보냅니다
	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
	//중앙노드에게 보낸다
	//var blocksInTransit = [][]byte{} 해당 변수가 중앙노드의 블록으로 동기화됩니다.
}
func commandToBytes(command string) []byte { //12 바이트 버퍼를 만들어 커맨드명을 채워넣은 뒤 나미저 바이트는 빈 상태 그대로 둔다
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}
func bytesToCommand(bytes []byte) string { //바이트 시퀀스를 커맨드로 변환하는 함수,
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}
func handleConnection(conn net.Conn, bc *Blockchain) { // 노드가 컨맨드를 수신하면 커맨드명을 가져와 핸들러로 조절함
	request, err := ioutil.ReadAll(conn) //모든 요청을 읽고
	if err != nil {
		log.Panic(err)
	}

	command := bytesToCommand(request[:commandLength]) //커멘드를 읽어온다
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

// 주소를 추가하는  핸들러
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

// 노드들에게 getBloks, 노드 주소를 보냅니다.
func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}

// 중앙노드에게 getBloks를 요청
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})               //payload안에 든 노드 주소를 인코딩
	request := append(commandToBytes("getblocks"), payload...) //요청한다 getBlocks payload 정보 전달

	sendData(address, request) //내 노드에게 getBlocks를 보냅니다?
}

// 블럭을 받습니다
func sendBlock(addr string, b *Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request) //데이터를 보냅니다
}

// Version입력이 오면 해당 핸들러 실행
func handleVersion(request []byte, bc *Blockchain) {
	var buff bytes.Buffer //임시 데이터 저장소 request 바이트 슬라이스 일부를 저장함
	var payload verzion

	buff.Write(request[commandLength:]) //request를  읽어서 저장 12[]까지
	dec := gob.NewDecoder(&buff)        //request를 디코딩
	err := dec.Decode(&payload)         //buff에서 디코딩된 것을 payload에 저장
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()        //나의 최대길이와
	foreignerBestHeight := payload.BestHeight //다른 노드에 최대길이를 구함

	if myBestHeight < foreignerBestHeight { //중앙 노드 최대길이가 더 길다면
		sendGetBlocks(payload.AddrFrom) //중앙노드에게 요청
	} else if myBestHeight > foreignerBestHeight { //중앙 노드보다 내 최대 길이가 더 길다면
		sendVersion(payload.AddrFrom, bc) //중앙노드에게 내 BC를 보냄
	}

	// sendAddr(payload.AddrFrom)
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

// addr == 중앙노드  데이터를 보냅니다
func sendData(addr string, data []byte) { //주소와 getBlocks를 보낸다
	conn, err := net.Dial(protocol, addr) //서버 열려있는지 확인
	if err != nil {
		fmt.Printf("%s is not available\n", addr) //주소가 올바르지 않음
		var updatedNodes []string

		for _, node := range knownNodes { //서버업데이트
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data)) //중앙노드에게 data 즉 payload를 conn -> 노드 주소에게 요청받은 Reuqest를 보냅니다 Payload죠
	if err != nil {
		log.Panic(err)
	}
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

// 중앙노드에게 블록 Hash 요청
func handleGetBlocks(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:]) //커멘드를 읽습니다.
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload) //페이로드에 페이로드를 저장합니다.
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()              //모든 블록의 Hash를 List화해서 가져옵니다.
	sendInv(payload.AddrFrom, "block", blocks) //중앙 노드에게 block과 List를 보냄
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

// 중앙노드에게 inv Block Hash List 전달
func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}           //노드 주소, Block, HashList
	payload := gobEncode(inventory)                      //페이로드에 저장
	request := append(commandToBytes("inv"), payload...) //inv라는 커멘드로 저장

	sendData(address, request) //해당 노드 주소에게 Request를 보냅니다 //blocks
}

// 중앙노드가 Block Hahs List를 전달받음 메시지 : Block
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
		blocksInTransit = payload.Items //Block Hash List 저장

		blockHash := payload.Items[0]                     //마지막 Hash 저장?
		sendGetData(payload.AddrFrom, "block", blockHash) //중앙노드에게 메시지 : block, 첫번째 Hash

		newInTransit := [][]byte{} //새로운 블록HashList

		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b) //첫 번째 해시까지 복사함 아니면 마지막 Hash거나.. 아무튼 중앙노드 블록 체인 다 복사
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0] //마지막 Hash전달"?

		if mempool[hex.EncodeToString(txID)].ID == nil { //mempool에 TXID가 없으면? 그러니까 블럭이 있다는 거겠쬬
			sendGetData(payload.AddrFrom, "tx", txID) //주소에게 txID라는 handler에게 전달
		}
	}
}
func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

// 요청받은 것을 바탕으로 메소들를 실행합니다
func handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:]) //어떤 내용의 메시지인지 읽습니다
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID)) //paylaod.ID에 해당하는 블록의 정보를 가져옵니다
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" { //TX
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID] //mempool안에 존재하는 TX를 가져옵니다

		sendTx(payload.AddrFrom, &tx) //해당 주소에게 mempool에 존재하는 TX를 전달합니다.
		// delete(mempool, txID)
	}
}

// 주소와 TX를 받습니다.
func sendTx(addr string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(addr, request) //데이터를 보냅니다.
}

// 새로운 블럭이 생성되면 추가하는 핸들러
func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block //해시가 담겨 있겠죠
	block := DeserializeBlock(blockData)

	fmt.Println("Recevied a new block!")
	bc.AddBlock(block) //블록체인에 block을 더합니다

	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]                   //마지막 해쉬 바꿈
		sendGetData(payload.AddrFrom, "block", blockHash) //

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

// TX와 요청을 전달받음
func handleTx(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	//주소 "TX" TXID
	txData := payload.Transaction
	tx := DeserializeTransaction(txData)    //TXID를 역직렬화
	mempool[hex.EncodeToString(tx.ID)] = tx //mempool의 TXID의 TX를 저장합니다

	if nodeAddress == knownNodes[0] { //중앙서버라면
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID}) //중앙서버가 아닌 것들에게 TXID를 전부 보냅니다
			}
		}
	} else { //중앙 서버가 아니라면
		if len(mempool) >= 2 && len(miningAddress) > 0 { //mempool이 >=2	채굴자가 1명이상이면
		MineTransactions:
			var txs []*Transaction

			for id := range mempool {
				tx := mempool[id]              //mempool에 있는 tx들을 가져옴
				if bc.VerifyTransaction(&tx) { //검증
					txs = append(txs, &tx) //검증된 것을 더함
				}
			}

			if len(txs) == 0 { //TX가 없다면
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbTx := NewCoinbaseTX(miningAddress, "") //첫 트랜잭션 생성
			txs = append(txs, cbTx)                  //TXS에 첫 TX를 더함

			newBlock := bc.MineBlock(txs) //채굴자 블럭 생성
			UTXOSet := UTXOSet{bc}        //UTXOSet생성
			UTXOSet.Reindex()             //초기 진행

			fmt.Println("New block is mined!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID) //블럭 생성 했으니 mempool TX 삭제
			}

			for _, node := range knownNodes {
				if node != nodeAddress { //현재 노드가 아닌 노드들에게 block을 더해줍니다.
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(mempool) > 0 { //TX가 하나라도 있으면 Mining합니다
				goto MineTransactions
			}
		}
	}
}
