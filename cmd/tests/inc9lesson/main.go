package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	password = "x35k9f"
	msg      = `0ba7cd8c624345451df4710b81d1a349ce401e61bc7eb704ca` +
		`a84a8cde9f9959699f75d0d1075d676f1fe2eb475cf81f62ef` +
		`f701fee6a433cfd289d231440cf549e40b6c13d8843197a95f` +
		`8639911b7ed39a3aec4dfa9d286095c705e1a825b10a9104c6` +
		`be55d1079e6c6167118ac91318fe`
)

// func generateRandom(size int) ([]byte, error) {
// 	b := make([]byte, size)
// 	_, err := rand.Read(b)
// 	if err != nil {
// 		return nil, err
// 	}

//		return b, nil
//	}
func main() {

	src := []byte(password)

	key := sha256.Sum256(src)

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	aesgcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	// создаём вектор инициализации

	nonceSize := aesgcm.NonceSize()
	nonce := key[len(key)-nonceSize:]
	msgBytes, err := hex.DecodeString(msg)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	msgOpen, err := aesgcm.Open(nil, nonce, msgBytes, nil)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("decrypted: %s\n", msgOpen)
}
