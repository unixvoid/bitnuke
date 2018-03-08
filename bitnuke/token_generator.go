package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v5"
)

func tokenGen(strSize int, client *redis.Client) string {
	// generate new token
	token := randStr(strSize, &config.Bitnuke.TokenDictionary)

	// hash token
	hash := sha3.Sum512([]byte(token))
	hashstr := fmt.Sprintf("%x", hash)
	_, err := client.Get(hashstr).Result()
	for err != redis.Nil {
		glogger.Debug.Println("DEBUG :: COLLISION")
		token = randStr(strSize, &config.Bitnuke.TokenDictionary)
		hash := sha3.Sum512([]byte(token))
		hashstr := fmt.Sprintf("%x", hash)
		_, err = client.Get(hashstr).Result()
		// sleep inbetween retries
		time.Sleep(time.Second * 1)
	}
	return token
}

func randStr(strSize int, dictionary *string) string {
	uberDictionary := *dictionary
	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = uberDictionary[v%byte(len(uberDictionary))]
	}
	return string(bytes)
}

func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}
