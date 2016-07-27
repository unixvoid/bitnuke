package main

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v3"
)

func remove(w http.ResponseWriter, r *http.Request, client *redis.Client) {
	// parse form and values
	r.ParseForm()
	clientToken := strings.TrimSpace(r.FormValue("token"))
	clientSec := strings.TrimSpace(r.FormValue("sec"))
	//hash := sha3.Sum512([]byte(clientToken))
	clientHash := fmt.Sprintf("%x", sha3.Sum512([]byte(clientToken)))
	clientSecHash := fmt.Sprintf("%x", sha3.Sum512([]byte(clientSec)))
	// verify all params sent
	if len(clientToken) == 0 || len(clientSec) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// make sure token exists
	_, err := client.Get(clientHash).Result()
	if err == redis.Nil {
		// id does not exist
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		// id exists, verify that auth matches
		storedSec, _ := client.Get(fmt.Sprintf("sec:%s", clientHash)).Result()
		if storedSec == clientSecHash {
			// authed. remove token
			client.Del(clientHash)
			client.Del(fmt.Sprintf("sec:%s", clientHash))
		} else {
			// forbidden
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
}
