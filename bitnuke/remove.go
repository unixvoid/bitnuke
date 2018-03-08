package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v5"
)

func remove(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	// parse form and values
	r.ParseForm()
	fileId := strings.TrimSpace(r.FormValue("file_id"))
	secureKey := strings.TrimSpace(r.FormValue("sec_key"))
	deleteToken := strings.TrimSpace(r.FormValue("removal_key"))

	// verify all params sent
	if len(fileId) == 0 || len(deleteToken) == 0 || len(secureKey) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// sha3:512 hash the file id to get the long id
	longFileId := fmt.Sprintf("%x", sha3.Sum512([]byte(fileId)))

	// make sure token exists
	encryptedDeleteToken, err := redisClient.HGet(longFileId, "delete_token").Result()
	if err != nil {
		glogger.Debug.Println("token does not exist")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// try and decrypt the delete token
	unencryptedDeleteToken, err := decrypt([]byte(secureKey), []byte(encryptedDeleteToken))
	if err != nil {
		glogger.Debug.Println("error decrypting delete token :: forbidden")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// see if the delete token matches the client provided one
	if deleteToken == string(unencryptedDeleteToken) {
		// expire the cookie client-side
		cookie := http.Cookie{Name: fileId, Expires: time.Now().Add(-100 * time.Hour)}
		http.SetCookie(w, &cookie)

		// client is authed to remove data
		err := os.Remove(fmt.Sprintf("%s/%s", config.Bitnuke.FileStorePath, longFileId))
		if err != nil {
			glogger.Debug.Println("error removing file from filesystem")
		} else {
			glogger.Debug.Printf("removing %s filesystem\n", longFileId)
		}
		redisClient.Del(longFileId)
	} else {
		// client is not authed to remove data
		glogger.Debug.Println("delete tokens do not match :: forbidden")
		w.WriteHeader(http.StatusForbidden)
		return
	}
}
