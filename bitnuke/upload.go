package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v5"
)

// the `json:""` is so we can have fields without capital letters
type CValue struct {
	File_id     string `json:"file_id"`
	File_name   string `json:"file_name"`
	Sec_key     string `json:"sec_key"`
	Removal_key string `json:"removal_key"`
}

func upload(w http.ResponseWriter, r *http.Request, redisClient *redis.Client, state string) {
	// get file POST from index
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))
		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)

		// set default filename
		filename := "unnamed_file"

		file, multipartFileHeader, err := r.FormFile("file")
		if err != nil {
			glogger.Error.Println(err)
			return
		} else {
			// overwrite default filename with parsed filename
			filename = fmt.Sprintf("%v", multipartFileHeader.Filename)
		}
		defer file.Close()

		// generate all tokens/keys
		fileId := tokenGen(config.Bitnuke.TokenSize, redisClient)
		secToken := tokenGen(config.Bitnuke.SecTokenSize, redisClient)
		delToken := tokenGen(config.Bitnuke.DelTokenSize, redisClient)

		// set client headers
		w.Header().Set("file_id", fileId)
		w.Header().Set("sec_key", secToken)
		w.Header().Set("removal_key", delToken)

		// generate json for cookie
		cVal := &CValue{
			File_id:     fileId,
			File_name:   filename,
			Sec_key:     secToken,
			Removal_key: delToken,
		}
		b, _ := json.Marshal(cVal)
		base64JsonC := base64.StdEncoding.EncodeToString(b)

		// set cookie expiration
		expiration := time.Now().Add(24 * time.Hour)
		cookie := http.Cookie{Name: fileId, Value: base64JsonC, Expires: expiration}
		http.SetCookie(w, &cookie)

		// return file_id and sec_key to client
		fmt.Fprintf(w, "%s/%s", fileId, secToken)

		//glogger.Debug.Println("file id:       ", fileId)
		//glogger.Debug.Println("secret key:    ", secToken)
		//glogger.Debug.Println("delete token:  ", delToken)

		// encrypt fileId
		encryptedFilename, err := encrypt([]byte(secToken), []byte(filename))
		if err != nil {
			glogger.Debug.Println("error encrypting filename")
			panic(err.Error())
		}
		// encrypt delToken
		encryptedDelToken, err := encrypt([]byte(secToken), []byte(delToken))
		if err != nil {
			glogger.Debug.Println("error encrypting delete token")
			panic(err.Error())
		}

		// get sha3:512 of fileId
		fileIdHash := sha3.Sum512([]byte(fileId))
		longFileId := fmt.Sprintf("%x", fileIdHash)

		// set hash metadata in redis
		_, err = redisClient.HMSet(longFileId, map[string]string{
			"filename":    string(encryptedFilename),
			"deleteToken": string(encryptedDelToken),
		}).Result()
		if err != nil {
			glogger.Error.Println("error setting meta hash key in redis")
		}

		// set <hash> (file) in redis
		glogger.Debug.Println("uploading:", fileId)
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

		// encrypt file
		encryptedFile, err := encrypt([]byte(secToken), []byte(fileBase64Str))
		if err != nil {
			glogger.Debug.Println("error encrypting file")
			panic(err.Error())
		}

		// store on disk instead of writing to redis
		//redisClient.Set(fmt.Sprintf("%s", longFileId), encryptedFile, 0).Err()

		// write contents to file
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s", config.Bitnuke.FileStorePath, longFileId), encryptedFile, 0644)
		if err != nil {
			glogger.Error.Println("error creating file on filesystem")
			panic(err.Error())
		}

		// expire data
		redisClient.Expire(fmt.Sprintf("%s", longFileId), (config.Bitnuke.TTL * time.Hour)).Err()
		redisClient.Expire(longFileId, (config.Bitnuke.TTL * time.Hour)).Err()
		glogger.Debug.Println("expire link generated")
		os.Remove("tmpfile")
	}
}
