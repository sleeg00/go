package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

// 지값에 키들을 저장하고 다님
type Wallet struct {
	PrivateKey ecdsa.PrivateKey //비밀 키
	PublicKey  []byte           //공개 키

}

// 새로운 지갑에 키쌍을 생성
func NewWallet() *Wallet {
	private, public := newKeyPair()   //키 쌍 생성
	wallet := Wallet{private, public} //저장

	return &wallet
}

// 지갑의 주소를 반환한다. (엄청난 해싱을 거쳐서)
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey) //공개 키를 암호화(해싱)이다 => 해시가 되었다.

	versionedPayload := append([]byte{version}, pubKeyHash...) //주소 생성 알고리즘(암호하 했던 것 Hash)에다가 버전을 추가한다.
	checksum := checksum(versionedPayload)                     //체크섬은 = Version + 공개 키 조합의 접미에 추가한다.

	fullPayload := append(versionedPayload, checksum...) //다 더한다.
	address := Base58Encode(fullPayload)                 //또 Base58로 해싱

	return address
}

// 공개키를 SHA256해싱 알고리즘으로 두번 해싱한다
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// 주소가 유효한 것인지 확인합니다.
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// 아무튼 합해서 해시하고 또해서 암호화
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

// 타원곡선 알고리즘을 이용해  비밀키를 생성해낸다 X,Y좌표의 조합 이 좌표들은 하나로 연결되어 공개 키를 형성한다.
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}
