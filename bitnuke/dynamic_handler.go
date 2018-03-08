package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/gorilla/mux"
	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v5"
)

func handlerdynamic(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	vars := mux.Vars(r)
	dataId := vars["dataId"]
	secureKey := vars["secureKey"]

	// hash the token that is passed
	fileIdHash := sha3.Sum512([]byte(dataId))
	longFileId := fmt.Sprintf("%x", fileIdHash)

	// get the client's ip
	ip := strings.Split(r.RemoteAddr, ":")[0]

	// set client as localhost if it comes from localhost
	if ip == "[" {
		ip = "localhost"
	}

	// pull the client's real header if proxied. (if X-Forwarded-For is set)
	realIp := r.Header.Get("X-Forwarded-For")
	if realIp != "" {
		ip = realIp
	}

	// try and pull the data from redis
	encryptedFilename, err := redisClient.HGet(longFileId, "filename").Result()
	if err != nil {
		// handle the error if the token does not exist
		glogger.Debug.Printf("data does not exist %s :: from: %s\n", dataId, ip)
		fmt.Fprintf(w, "token not found")
	} else {

		glogger.Debug.Printf("Responsing to %s :: from: %s\n", dataId, ip)

		// token exists, try and decrypt
		val, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", config.Bitnuke.FileStorePath, longFileId))
		if err != nil {
			// the file is either unreadable or not found
			glogger.Debug.Println("error reading file from filesystem")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// decrypt
		plainFile, err := decrypt([]byte(secureKey), []byte(val))
		if err != nil {
			glogger.Debug.Println("unauthorized access.")
			// error decrypting file, looks like the wrong key
			w.WriteHeader(http.StatusForbidden)
			return
		}
		decodeVal, _ := base64.StdEncoding.DecodeString(string(plainFile))

		file, _ := os.Create("tmpfile")
		io.WriteString(file, string(decodeVal))
		file.Close()

		// unencrypt filename
		filename, err := decrypt([]byte(secureKey), []byte(encryptedFilename))
		if err != nil {
			glogger.Debug.Println("error decrypting filename")
			panic(err.Error())
		}

		// dont add the filename header to links
		if string(filename) != "bitnuke:link" {
			finalFname := fmt.Sprintf("INLINE; filename=%s", filename)
			w.Header().Set("Content-Disposition", finalFname)
		}
		http.ServeFile(w, r, "tmpfile")
		os.Remove("tmpfile")
	}

	// force garbage collection
	runtime.GC()
}
