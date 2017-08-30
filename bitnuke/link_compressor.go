package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v5"
)

func linkcompressor(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	err := r.ParseForm()
	if err != nil {
		glogger.Error.Println("error during form parse")
	}

	// generate redirect html
	page := fmt.Sprintf("<html><head><meta http-equiv=\"refresh\" content=\"0;URL=%s\"></head></html>", r.PostFormValue("link"))
	glogger.Debug.Println("link created")
	content64Str := base64.StdEncoding.EncodeToString([]byte(page))

	// generate token and hash it to store in db
	token := tokenGen(config.Bitnuke.LinkTokenSize, redisClient)
	hash := sha3.Sum512([]byte(token))
	hashstr := fmt.Sprintf("%x", hash)

	// throw it in the db
	redisClient.Set(hashstr, content64Str, 0).Err()
	redisClient.Set(fmt.Sprintf("fname:%s", hashstr), "bitnuke:link", 0).Err()
	redisClient.Expire(hashstr, (config.Bitnuke.TTL * time.Hour)).Err()

	// return token to redisClient
	w.Header().Set("compressor", token)
	fmt.Fprintf(w, "%s", token)
}
