package main

import (
	"bufio"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v3"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	//========handle web page=========
	router.HandleFunc("/css/style.css", css).Methods("GET")
	router.HandleFunc("/js/index.js", js).Methods("GET")
	router.HandleFunc("/bitnuke.png", img).Methods("GET")
	router.HandleFunc("/js/dropzone.js", dropjs).Methods("GET")
	router.HandleFunc("/css/dropzone.css", dropcss).Methods("GET")
	//========/handle web page========
	router.HandleFunc("/{fdata}", handlerdynamic).Methods("GET")
	router.HandleFunc("/upload", upload)
	log.Fatal(http.ListenAndServe(":8803", router))
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

func dropjs(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./upload/js/dropzone.js")
}

func dropcss(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./upload/css/dropzone.css")
}

func handlerdynamic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fdata := vars["fdata"]
	// init redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// hash the token that is passed
	hash := sha3.Sum512([]byte(fdata))
	hashstr := fmt.Sprintf("%x", hash)

	val, err := client.Get(hashstr).Result()
	if err != nil {
		log.Printf("data does not exist")
		fmt.Fprintf(w, "token not found")
	} else {
		//log.Printf("data exists")
		ip := strings.Split(r.RemoteAddr, ":")[0]
		log.Printf("Responsing to %s :: from: %s", fdata, ip)

		decodeVal, _ := base64.StdEncoding.DecodeString(val)

		file, _ := os.Create("tmpfile")
		io.WriteString(file, string(decodeVal))
		file.Close()

		http.ServeFile(w, r, "tmpfile")
		os.Remove("tmpfile")
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	// get file POST from index
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
			token := randStr(8)
			w.Header().Set("token", token)
			fmt.Fprintf(w, "https://bitnuke.io/%s", token)
			fmt.Println("deez")
			fmt.Println(err)
			return
		}
		defer file.Close()

		// generate token and hash to save
		token := randStr(8)
		w.Header().Set("token", token)
		fmt.Fprintf(w, "https://bitnuke.io/%s", token)

		// done with client, rest is server side
		hash := sha3.Sum512([]byte(token))
		hashstr := fmt.Sprintf("%x", hash)
		fmt.Println("uploading:", token)

		// write file temporarily to get filesize
		f, _ := os.OpenFile("tmpfile", os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		io.Copy(f, file)

		tmpFile, _ := os.Open("tmpfile")
		defer tmpFile.Close()

		client := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})

		fInfo, _ := tmpFile.Stat()
		var size int64 = fInfo.Size()
		buf := make([]byte, size)

		// read file content into buffer
		fReader := bufio.NewReader(tmpFile)
		fReader.Read(buf)

		fileBase64Str := base64.StdEncoding.EncodeToString(buf)

		//println("uploading ", "file")
		client.Set(hashstr, fileBase64Str, 0).Err()
		client.Expire(hashstr, (12 * time.Hour)).Err()
		os.Remove("tmpfile")
	}
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
