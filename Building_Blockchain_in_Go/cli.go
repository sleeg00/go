package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) createBlockchain(address string) { //CreateBlockChain
	bc := CreateBlockchain(address)
	//// 말그대로 블록체인을 생성한다 Genesis 블록을 생성하고 Output 생성 (Genesis, Output, LastHash-> 블록 연결점 생성)
	//주소는 어디다가 쓰는거지?
	bc.db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockchain(address)
	defer bc.db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) printUsage() { //사용법
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) validateArgs() { //예외처리 두글자 이상인가?
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printChain() { //Print BlockChain
	// TODO: Fix this
	bc := NewBlockchain("")
	defer bc.db.Close()

	bci := bc.Iterator() //최근 거래내역 부터

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)         //이전 블록
		fmt.Printf("Hash: %x\n", block.Hash)                        //블록 해쉬
		pow := NewProofOfWork(block)                                //작업증명 구조체 (블록, 타겟값)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate())) //이것이 존재하는 블록 즉 유효한 블럭인지 판별
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) send(from, to string, amount int) { //메시지 보내기  (Output작성)
	bc := NewBlockchain(from) //블록체인에서  DB를 뒤져 마지막 블록 Hash를 가져옵니다
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc) //TX를 기록하고 그 기록을 가져옵니다.
	bc.MineBlock([]*Transaction{tx})               //코인은 전송한다는 건 TX를 만들고 블록 채굴을 통해 이를 블록체인에 추가한다는 것.
	//Tansaction이 생기고 비트코인을 채굴하면 그 Transaction을 기록한다. 평균 10분간격이지만 보상이 지금은 낮다.
	fmt.Println("Success!") //성공 돈을 보냈고 돈은 받은놈이 채굴을 했다.
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
