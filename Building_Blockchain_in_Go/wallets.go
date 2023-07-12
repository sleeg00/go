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
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile() //지갑정보를 가져온다

	return &wallets, err
}

// 지갑 주소 반환 (공개 키 해싱 값 )
func (ws *Wallets) CreateWallet() string {
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

// 지갑 정보를 찾고 디코딩해서 반환
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) { // 지갑 없으면 에러
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile) //지갑 읽기
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent)) //디코딩
	err = decoder.Decode(&wallets)                          //디코딩
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// 지갑을 디코딩해서 저장한다
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
