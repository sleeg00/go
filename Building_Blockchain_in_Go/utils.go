package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)                        //새로운 바이트 생성
	err := binary.Write(buff, binary.BigEndian, num) //바이트리 인코딩을 저장
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes() //반환
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
