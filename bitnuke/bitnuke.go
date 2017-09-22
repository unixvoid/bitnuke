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
	"gopkg.in/redis.v5"
)

type Config struct {
	Bitnuke struct {
		Loglevel            string
		Port                int
		FileStorePath       string
		JanitorSleepSeconds time.Duration
		TokenSize           int
		SecTokenSize        int
		DelTokenSize        int
		LinkTokenSize       int
		TTL                 time.Duration
		TokenDictionary     string
		SecDictionary       string
		BootstrapDelay      time.Duration
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
	// allow the bootstrap delay time if needed
	// this allows redis to start before the app connects
	// valuable when deploying in a container

	redisClient, redisErr := initRedisConnection()
	if redisErr != nil {
		glogger.Debug.Printf("redis connection cannot be made, trying again in %s second(s)\n", config.Bitnuke.BootstrapDelay*time.Second)
		time.Sleep(config.Bitnuke.BootstrapDelay * time.Second)
		redisClient, redisErr = initRedisConnection()
		if redisErr != nil {
			glogger.Error.Println("redis connection cannot be made, exiting.")
			panic(redisErr)
		}
	} else {
		glogger.Debug.Println("connection to redis succeeded.")
	}

	// create the filestore dir if it does not exist
	_, err := os.Stat(config.Bitnuke.FileStorePath)
	if err != nil {
		// dir does not exist, create it
		glogger.Debug.Printf("creating directory: %s\n", config.Bitnuke.FileStorePath)
		os.Mkdir(config.Bitnuke.FileStorePath, os.ModePerm)
	}

	// fire up the janitor to monitor expire times
	go janitor(redisClient)

	// all handlers. lookin funcy casue i have to pass redis handler
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		upload(w, r, redisClient, "tmp")
	})
	router.HandleFunc("/remove", func(w http.ResponseWriter, r *http.Request) {
		remove(w, r, redisClient)
	})
	router.HandleFunc("/compress", func(w http.ResponseWriter, r *http.Request) {
		linkcompressor(w, r, redisClient)
	})
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// client wants favicon, send back a does not exist
		w.WriteHeader(http.StatusNotFound)
	}).Methods("GET")
	router.HandleFunc("/{dataId}/{secureKey}", func(w http.ResponseWriter, r *http.Request) {
		handlerdynamic(w, r, redisClient)
	}).Methods("GET")
	router.HandleFunc("/{dataId}", func(w http.ResponseWriter, r *http.Request) {
		linkhandler(w, r, redisClient)
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

func janitor(redisClient *redis.Client) {
	// run janitor loop forever, sleep for a time between checks
	for {
		// diff keys in filesystem and keys in redis

		// get the list of keys in redis
		rl := redisClient.Keys("*")

		// get a list of all keys on filesystem
		fl, _ := ioutil.ReadDir(config.Bitnuke.FileStorePath)
		fs := make([]string, 0)
		for _, f := range fl {
			fs = append(fs, f.Name())
		}

		// get the diff of these two slices
		// since these are unsorted we will have to use maps
		m1 := map[string]bool{}
		for _, x := range rl.Val() {
			m1[x] = true
		}
		m2 := []string{}
		for _, x := range fs {
			if _, ok := m1[x]; !ok {
				m2 = append(m2, x)
			}
		}

		// remove file on filesystem that is not in redis
		for _, cf := range m2 {
			err := os.Remove(fmt.Sprintf("%s/%s", config.Bitnuke.FileStorePath, cf))
			if err != nil {
				glogger.Debug.Println("error removing file from filesystem")
			} else {
				glogger.Debug.Printf("removed expired file %s\n", cf)
			}
		}

		// sleep for a time between checks
		time.Sleep(config.Bitnuke.JanitorSleepSeconds * time.Second)
	}
}
