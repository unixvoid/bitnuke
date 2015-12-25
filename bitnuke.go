package main

import (
	"crypto/md5"
	"fmt"
	"github.com/gorilla/mux"
	//"io/ioutil"
	"crypto/rand"
	//"encoding/base64"
	"log"
	"net/http"
	"os"
	//"strconv"
	//"time"
)

/*
//=====================================
// general strategy:
// we take in a file, the filename is a hashed random string.
// the file is stored with its filename as the hased string.
// the random string (token) is returned back to the user.
//
// now when the user wants to retrive the file, he puts in the
// token (random string from earlier). his request is hashed and
// the stored has is returned. ez
//=====================================
*/

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/{fdata}", handlerdynamic).Methods("GET")
	log.Fatal(http.ListenAndServe(":8888", router))
}

func handlerdynamic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fdata := vars["fdata"]

	// hash the token that is passed
	hash := md5.Sum([]byte(fdata))
	hashstr := fmt.Sprintf("%x", hash)

	token := randStr(8)
	//log.Printf(token)

	// serve up the hashed token if it exists
	path := fmt.Sprintf("./tmpnuke/%s", hashstr)
	// if data exists
	if _, err := os.Stat(path); err == nil {
		log.Printf("data exists")
		tokenserv(w, r, hashstr)
	}
	// if data does not exist
	if _, err := os.Stat(path); err != nil {
		fmt.Fprintf(w, "token not found")
	}
}

func tokenserv(w http.ResponseWriter, r *http.Request, data string) {
	log.Printf("Responsing to", data)
	item := fmt.Sprintf("./tmpnuke/%s", data)
	http.ServeFile(w, r, item)
}

func randStr(strSize int) string {
	dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}
