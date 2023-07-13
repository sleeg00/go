package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// 지갑에 대한 배열 key - Struct
type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallets() (*Wallets, error) {
	wallets := Wallets{}                       //Walltes 구조체 생성
	wallets.Wallets = make(map[string]*Wallet) //Wallets필드에 make구조체 생성

	err := wallets.LoadFromFile() //지갑정보를 가져온다

	return &wallets, err
}

// 지갑 주소 반환 (공개 키 해싱 값 )
func (ws *Wallets) CreateWallet() string {
	if ws.Wallets == nil {
		ws.Wallets = make(map[string]*Wallet) // 맵 초기화
	}
	wallet := NewWallet()                             //지갑을 만들고 map[String][공개 키 - 비킬 키]로 이루어진
	address := fmt.Sprintf("%s", wallet.GetAddress()) //주소 출력

	ws.Wallets[address] = wallet //지갑 여러개 [주소] = 지갑 하나

	return address
}

// 지갑들의 주소를 address에다가 {1, 3423, 234132} 저장한다.
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a Wallet by its address
func (ws Wallets) GetWallet(address string) Wallet {

	return *ws.Wallets[address]
}

func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	if len(fileContent) == 0 {
		// 파일이 비어있는 경우, Wallets를 초기화합니다.
		ws.Wallets = make(map[string]*Wallet)
		return nil
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

//1FPBR8iLSM3wm1mwfZsMuoMY8D5myJRVsZ
//15sdgcZXS8EKER1TJT5FUjT51bwdvEJ5vN
