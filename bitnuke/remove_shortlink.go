package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v5"
)

func removeShortLink(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	r.ParseForm()
	shortToken := strings.TrimSpace(r.FormValue("short_token"))
	deleteToken := strings.TrimSpace(r.FormValue("delete_token"))

	if len(shortToken) == 0 || len(deleteToken) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Hash the token as in linkhandler
	hash := sha3.Sum512([]byte(shortToken))
	hashstr := fmt.Sprintf("%x", hash)
	redisKey := fmt.Sprintf("link:%s", hashstr)
	deleteTokenKey := fmt.Sprintf("link_delete_token:%s", hashstr)

	// Check the delete token in Redis
	storedDeleteToken, err := redisClient.Get(deleteTokenKey).Result()
	if err != nil {
		glogger.Debug.Println("delete token does not exist for short link")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if deleteToken != storedDeleteToken {
		glogger.Debug.Println("delete token does not match for short link :: forbidden")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Delete the short link and its delete token from Redis
	err = redisClient.Del(redisKey).Err()
	err2 := redisClient.Del(deleteTokenKey).Err()
	if err != nil || err2 != nil {
		glogger.Debug.Println("error removing short link or delete token from redis")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	glogger.Debug.Printf("removed short link %s and delete token from redis\n", redisKey)
	w.WriteHeader(http.StatusOK)
}