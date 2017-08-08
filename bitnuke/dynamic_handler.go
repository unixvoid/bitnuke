package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v3"
)

func handlerdynamic(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	vars := mux.Vars(r)
	fdata := vars["fdata"]

	// hash the token that is passed
	hash := sha3.Sum512([]byte(fdata))
	hashstr := fmt.Sprintf("%x", hash)

	val, err := redisClient.Get(hashstr).Result()
	filename, err := redisClient.Get(fmt.Sprintf("fname:%s", hashstr)).Result()
	if err != nil {
		glogger.Debug.Printf("data does not exist")
		fmt.Fprintf(w, "token not found")
	} else {
		glogger.Debug.Printf("data exists")
		ip := strings.Split(r.RemoteAddr, ":")[0]
		if ip == "[" {
			// set client as localhost if it comes from localhost
			ip = "localhost"
		}
		glogger.Debug.Printf("Responsing to %s :: from: %s", fdata, ip)

		decodeVal, _ := base64.StdEncoding.DecodeString(val)

		file, _ := os.Create("tmpfile")
		io.WriteString(file, string(decodeVal))
		file.Close()

		if filename != "bitnuke:link" {
			finalFname := fmt.Sprintf("attachment; filename=%s", filename)
			w.Header().Set("Content-Disposition", finalFname)
		}
		http.ServeFile(w, r, "tmpfile")
		os.Remove("tmpfile")
	}
}
