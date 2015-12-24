package main

import (
	"crypto/md5"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
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

	//hash the token that is passed
	hash := md5.Sum([]byte(fdata))
	hashstr := fmt.Sprintf("%x", hash)

	//serve up the hashed token if it exists
	token(w, r, hashstr)
}

func token(w http.ResponseWriter, r *http.Request, data string) {
	log.Printf("Responsing to", data)
	item := fmt.Sprintf("./tmpnuke/%s", data)
	http.ServeFile(w, r, item)
}
