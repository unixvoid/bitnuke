package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/unixvoid/glogger"
	"golang.org/x/crypto/sha3"
	"gopkg.in/redis.v3"
)

func upload(w http.ResponseWriter, r *http.Request, client *redis.Client, state string) {
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
			rawFilename := fmt.Sprintf("%v", multipartFileHeader.Filename)
			filename = rawFilename
		}
		defer file.Close()

		// generate token and hash to save
		token := tokenGen(config.Bitnuke.TokenSize, client)
		hash := sha3.Sum512([]byte(token))
		hashstr := fmt.Sprintf("%x", hash)

		// generate security token for removing content
		secToken := secTokenGen(hashstr, client)
		w.Header().Set("token", token)
		w.Header().Set("sec", secToken)
		fmt.Fprintf(w, "%s", token)

		// done with client, rest is server side
		glogger.Debug.Println("uploading:", token)
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

		client.Set(hashstr, fileBase64Str, 0).Err()
		client.Set(fmt.Sprintf("fname:%s", hashstr), filename, 0).Err()
		if strings.EqualFold(state, "tmp") {
			// expire if coming from /supload
			client.Expire(hashstr, (config.Bitnuke.TTL * time.Hour)).Err()
			client.Expire(fmt.Sprintf("sec:%s", hashstr), (config.Bitnuke.TTL * time.Hour)).Err()
			client.Expire(fmt.Sprintf("filename:%s", hashstr), (config.Bitnuke.TTL * time.Hour)).Err()
			glogger.Debug.Println("expire link generated")
		}
		os.Remove("tmpfile")
	}
}
