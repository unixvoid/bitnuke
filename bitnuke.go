package main

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/sha3"
	"html/template"
	"io"
	//"encoding/base64"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
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
	router.HandleFunc("/", landingpage).Methods("GET")
	router.HandleFunc("/css/style.css", css).Methods("GET")
	router.HandleFunc("/js/index.js", js).Methods("GET")
	router.HandleFunc("/bitnuke.png", img).Methods("GET")
	router.HandleFunc("/{fdata}", handlerdynamic).Methods("GET")
	router.HandleFunc("/upload", upload)
	log.Fatal(http.ListenAndServe(":8802", router))
}

func landingpage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./upload/index.html")
}

func css(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./upload/css/style.css")
}

func js(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./upload/js/index.js")
}

func img(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./upload/bitnuke.png")
}

func handlerdynamic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fdata := vars["fdata"]

	// hash the token that is passed
	hash := sha3.Sum512([]byte(fdata))
	hashstr := fmt.Sprintf("%x", hash)

	//token := randStr(8)
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

func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		//fmt.Fprintf(w, "%v", handler.Header)

		// generate token and hash to save
		token := randStr(8)
		fmt.Fprintf(w, "<html>"+
			"<style> "+
			//"body {background-color: #d3d3d3;"+
			//"font-family: Lato, Arial;"+
			//"color: #fff;}"+
			"a:link{color: black;"+
			"text-decoration: none;"+
			"font-weight: normal;}"+
			"a:visited{color: black;"+
			"text-decoration: none;"+
			"font-weight: normal;}"+
			"</style>"+
			"<p><a href=\"https://bitnuke.io/%v\">bitnuke.io/%v</a></p>"+
			"</html>", token, token)
		hash := sha3.Sum512([]byte(token))
		hashstr := fmt.Sprintf("%x", hash)
		fmt.Println(token)

		f, err := os.OpenFile("./tmpnuke/"+hashstr, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
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
