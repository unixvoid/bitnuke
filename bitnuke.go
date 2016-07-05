package main

import (
	"bufio"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/gcfg.v1"
	"gopkg.in/redis.v3"
)

/*
//================================================================
// general strategy:
// we take in a file, the filename is a hashed random string.
// the file is stored with its filename as the hased string.
// the random string (token) is returned back to the user.
//
// now when the user wants to retrive the file, he puts in the
// token (random string from earlier). his request is hashed and
// the stored hash is returned. ez
//================================================================
*/

type Config struct {
	Bitnuke struct {
		Port            int
		TokenSize       int
		LinkTokenSize   int
		TokenDictionary string
		TTL             time.Duration
	}
	Redis struct {
		Host string
		Port int
	}
	Server struct {
		Loglevel string
	}
}

var (
	config = Config{}
)

func main() {
	err := gcfg.ReadFileInto(&config, "config.gcfg")
	if err != nil {
		fmt.Printf("Could not load config.gcfg, error: %s\n", err)
		return
	}

	// init logger
	if config.Server.Loglevel == "debug" {
		glogger.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else if config.Server.Loglevel == "cluster" {
		glogger.LogInit(os.Stdout, os.Stdout, ioutil.Discard, os.Stderr)
	} else if config.Server.Loglevel == "info" {
		glogger.LogInit(os.Stdout, ioutil.Discard, ioutil.Discard, os.Stderr)
	} else {
		glogger.LogInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	}

	redisaddr := fmt.Sprint(config.Redis.Host, ":", config.Redis.Port)
	bitport := fmt.Sprint(":", config.Bitnuke.Port)
	glogger.Info.Println("bitnuke running on", config.Bitnuke.Port)
	glogger.Info.Println("link to redis on", redisaddr)
	// initialize redis connection
	client := redis.NewClient(&redis.Options{
		Addr:     redisaddr,
		Password: "",
		DB:       0,
	})

	// all handlers. lookin funcy casue i have to pass redis handler
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		upload(w, r, client, "tmp")
	})
	router.HandleFunc("/supload", func(w http.ResponseWriter, r *http.Request) {
		upload(w, r, client, "persist")
	})
	router.HandleFunc("/compress", func(w http.ResponseWriter, r *http.Request) {
		linkcompressor(w, r, client)
	})
	router.HandleFunc("/{fdata}", func(w http.ResponseWriter, r *http.Request) {
		handlerdynamic(w, r, client)
	}).Methods("GET")
	glogger.Error.Println(http.ListenAndServe(bitport, router))
}

func handlerdynamic(w http.ResponseWriter, r *http.Request, client *redis.Client) {
	vars := mux.Vars(r)
	fdata := vars["fdata"]

	// hash the token that is passed
	hash := sha3.Sum512([]byte(fdata))
	hashstr := fmt.Sprintf("%x", hash)

	val, err := client.Get(hashstr).Result()
	if err != nil {
		glogger.Debug.Println("data does not exist")
		fmt.Fprintf(w, "token not found")
	} else {
		//log.Printf("data exists")
		ip := strings.Split(r.RemoteAddr, ":")[0]
		glogger.Debug.Printf("Responsing to %s :: from: %s", fdata, ip)

		decodeVal, _ := base64.StdEncoding.DecodeString(val)

		file, _ := os.Create("tmpfile")
		io.WriteString(file, string(decodeVal))
		file.Close()

		http.ServeFile(w, r, "tmpfile")
		os.Remove("tmpfile")
	}
}

func upload(w http.ResponseWriter, r *http.Request, client *redis.Client, state string) {
	// get file POST from index
	//fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("file")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		// generate token and hash to save
		token := tokenGen(config.Bitnuke.TokenSize, client)
		w.Header().Set("token", token)
		fmt.Fprintf(w, "%s", token)

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

		fInfo, _ := tmpFile.Stat()
		var size int64 = fInfo.Size()
		buf := make([]byte, size)

		// read file content into buffer
		fReader := bufio.NewReader(tmpFile)
		fReader.Read(buf)

		fileBase64Str := base64.StdEncoding.EncodeToString(buf)

		//println("uploading ", "file")
		client.Set(hashstr, fileBase64Str, 0).Err()
		if strings.EqualFold(state, "tmp") {
			client.Expire(hashstr, (config.Bitnuke.TTL * time.Hour)).Err()
			glogger.Debug.Println("expire link generated")
		} else {
			client.Persist(hashstr).Err()
			glogger.Debug.Println("persistent link generated")
		}
		os.Remove("tmpfile")
	}
}

func linkcompressor(w http.ResponseWriter, r *http.Request, client *redis.Client) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println("error during form parse")
	}
	content := r.PostFormValue("link")
	page := fmt.Sprintf("<html><head><meta http-equiv=\"refresh\" content=\"0;URL=%s\"></head></html>", content)
	glogger.Debug.Println(page)
	content64Str := base64.StdEncoding.EncodeToString([]byte(page))
	// generate token and hash it to store in db
	token := tokenGen(config.Bitnuke.LinkTokenSize, client)
	hash := sha3.Sum512([]byte(token))
	hashstr := fmt.Sprintf("%x", hash)

	// throw it in the db
	client.Set(hashstr, content64Str, 0).Err()
	//client.Expire(hashstr, (config.Bitnuke.TTL * time.Hour)).Err()
	// return token to client
	w.Header().Set("compressor", token)
	fmt.Fprintf(w, "%s", token)
}

func tokenGen(strSize int, client *redis.Client) string {
	// generate new token
	token := randStr(strSize)
	println(token)
	// hash token
	hash := sha3.Sum512([]byte(token))
	hashstr := fmt.Sprintf("%x", hash)
	println(hashstr)

	test, err := client.Get(hashstr).Result()
	println(test)

	for err != redis.Nil {
		glogger.Error.Println("DEBUG :: COLLISION")
		token = randStr(strSize)
		hash := sha3.Sum512([]byte(token))
		hashstr := fmt.Sprintf("%x", hash)

		_, err = client.Get(hashstr).Result()
		// do not ddos box if db is full
		time.Sleep(time.Second * 1)
	}
	return token
}

func randStr(strSize int) string {
	//dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	dictionary := config.Bitnuke.TokenDictionary

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(bytes)
}
