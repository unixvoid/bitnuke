package main

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v3"
)

//func tokenGen(strSize int, redisClient *redis.Client) string {
//	// generate new token
//	token := randStr(strSize)
//	// hash token
//	hash := sha3.Sum512([]byte(token))
//	hashstr := fmt.Sprintf("%x", hash)
//
//	_, err := redisClient.Get(hashstr).Result()
//
//	for err != redis.Nil {
//		glogger.Debug.Println("COLLISION")
//		token = randStr(strSize)
//		hash := sha3.Sum512([]byte(token))
//		hashstr := fmt.Sprintf("%x", hash)
//
//		_, err = redisClient.Get(hashstr).Result()
//		// do not ddos box if db is full
//		time.Sleep(time.Second * 1)
//	}
//	return token
//}
//
//func randStr(strSize int) string {
//	//dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
//	dictionary := config.Bitnuke.TokenDictionary
//
//	var bytes = make([]byte, strSize)
//	rand.Read(bytes)
//	for k, v := range bytes {
//		bytes[k] = dictionary[v%byte(len(dictionary))]
//	}
//
//	return string(bytes)
//}

func secTokenGen(tokenHash string, client *redis.Client) string {
	// TODO expire this
	// generate sec token (25 char string) return to client
	// hash this (sha3:512) and store the hash under 'sec:<id> <hashed_sec>'
	// first gen the new string to spec
	secToken := randStr(config.Bitnuke.SecTokenSize, &config.Bitnuke.SecDictionary)
	secTokenHash := sha3.Sum512([]byte(secToken))
	// throw hash into redis
	client.Set(fmt.Sprintf("sec:%s", tokenHash), fmt.Sprintf("%x", secTokenHash), 0).Err()
	// return unhashed string
	return secToken
}
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
		// do not ddos box if db is full
		time.Sleep(time.Second * 1)
	}
	return token
}
func randStr(strSize int, dictionary *string) string {
	//dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	//dictionary := config.Bitnuke.TokenDictionary
	uberDictionary := *dictionary
	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = uberDictionary[v%byte(len(uberDictionary))]
	}
	return string(bytes)
}
