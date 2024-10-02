package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v5"
)

func linkhandler(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	vars := mux.Vars(r)
	dataId := vars["dataId"]

	// dont commit to work unless the link is properly sized
	if len(dataId) != config.Bitnuke.LinkTokenSize {
		return
	}

	hash := sha3.Sum512([]byte(dataId))
	hashstr := fmt.Sprintf("%x", hash)

	// check if the data is a link or not
	// base64 encode link request
	link, err := redisClient.Get(fmt.Sprintf("link:%s", hashstr)).Result()
	if err != nil {
		// data does not exist, 404
		w.WriteHeader(http.StatusNotFound)
	} else {
		// serve up redirect
		decodeVal, _ := base64.StdEncoding.DecodeString(link)
		http.Redirect(w, r, string(decodeVal), 302)
	}
}

func linkcompressor(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	err := r.ParseForm()
	if err != nil {
		glogger.Error.Println("error during form parse")
	}

	// generate redirect html
	//page := fmt.Sprintf("<html><head><meta http-equiv=\"refresh\" content=\"0;URL=%s\"></head></html>", r.PostFormValue("link"))
	glogger.Debug.Println("creating link")
	content64Str := base64.StdEncoding.EncodeToString([]byte(r.PostFormValue("link")))

	// generate token and hash it to store in db
	token := tokenGen(config.Bitnuke.LinkTokenSize, redisClient)
	hash := sha3.Sum512([]byte(token))
	hashstr := fmt.Sprintf("%x", hash)

	// throw it in the db
	redisClient.Set(fmt.Sprintf("link:%s", hashstr), content64Str, 0).Err()
	redisClient.Expire(fmt.Sprintf("link:%s", hashstr), (config.Bitnuke.TTL * time.Hour)).Err()

	// return token to redisClient
	w.Header().Set("compressor", token)
	fmt.Fprintf(w, "%s", token)
}
