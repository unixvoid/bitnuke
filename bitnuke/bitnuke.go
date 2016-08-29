package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/unixvoid/glogger"
	"gopkg.in/gcfg.v1"
	"gopkg.in/redis.v3"
)

type Config struct {
	Bitnuke struct {
		Loglevel        string
		Port            int
		TokenSize       int
		SecTokenSize    int
		LinkTokenSize   int
		TTL             time.Duration
		TokenDictionary string
		SecDictionary   string
	}
	Redis struct {
		Host     string
		Password string
	}
}

var (
	config = Config{}
)

func main() {
	// init config file and logger
	readConf()
	initLogger()

	// init redis connection
	// small sleep to allow redis to load data into memory
	// this is considered a hack for now. Redis will fail the ping/pong if
	// the database is not ready.  It is safe to use redis before it is ready
	// as it will store the new data in memory until it can access the databse
	time.Sleep(time.Second * 1)
	redisClient, redisErr := initRedisConnection()
	if redisErr != nil {
		glogger.Error.Println("redis connection cannot be made, exiting.")
		panic(redisErr)
	} else {
		glogger.Debug.Println("connection to redis succeeded.")
	}

	// all handlers. lookin funcy casue i have to pass redis handler
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		upload(w, r, redisClient, "tmp")
	})
	router.HandleFunc("/remove", func(w http.ResponseWriter, r *http.Request) {
		remove(w, r, redisClient)
	})
	router.HandleFunc("/supload", func(w http.ResponseWriter, r *http.Request) {
		upload(w, r, redisClient, "persist")
	})
	router.HandleFunc("/compress", func(w http.ResponseWriter, r *http.Request) {
		linkcompressor(w, r, redisClient)
	})
	router.HandleFunc("/{fdata}", func(w http.ResponseWriter, r *http.Request) {
		handlerdynamic(w, r, redisClient)
	}).Methods("GET")
	//log.Fatal(http.ListenAndServe(bitport, router))

	glogger.Info.Println("started server on", config.Bitnuke.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Bitnuke.Port), router))
}

func readConf() {
	// init config file
	err := gcfg.ReadFileInto(&config, "config.gcfg")
	if err != nil {
		panic(fmt.Sprintf("Could not load config.gcfg, error: %s\n", err))
	}
}

func initLogger() {
	// init logger
	if config.Bitnuke.Loglevel == "debug" {
		glogger.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else if config.Bitnuke.Loglevel == "cluster" {
		glogger.LogInit(os.Stdout, os.Stdout, ioutil.Discard, os.Stderr)
	} else if config.Bitnuke.Loglevel == "info" {
		glogger.LogInit(os.Stdout, ioutil.Discard, ioutil.Discard, os.Stderr)
	} else {
		glogger.LogInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	}
}
func initRedisConnection() (*redis.Client, error) {
	// init redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host,
		Password: config.Redis.Password,
		DB:       0,
	})

	_, redisErr := redisClient.Ping().Result()
	return redisClient, redisErr
}
