package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gronnbeck/ndc2017/setupredis"
)

var (
	readRedisAddr = flag.String("read-redis-addr", "localhost:6379", "tcp addr to connect redis tor")
	readRedisPass = flag.String("read-redis-pass", "", "password to redis server")

	writeRedisAddr = flag.String("write-redis-addr", "localhost:6379", "tcp addr to connect redis tor")
	writeRedisPass = flag.String("write-redis-pass", "", "password to redis server")
)

func init() {
	flag.Parse()

	if readRedisAddr == nil || *readRedisAddr == "" {
		panic("Read Reids addr cannot be empty")
	}

	if writeRedisAddr == nil || *writeRedisAddr == "" {
		panic("write Reids addr cannot be empty")
	}
}

func main() {
	readClient, err := setupredis.NewWait(*readRedisAddr, *readRedisPass)
	if err != nil {
		panic(err)
	}
	getHandleFunc := get(readClient)

	writeClient, err := setupredis.NewWait(*writeRedisAddr, *writeRedisPass)
	if err != nil {
		panic(err)
	}
	postHandleFunc := post(writeClient)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		done := make(chan bool)
		go func() {
			if strings.ToLower(r.Method) == "get" {
				getHandleFunc(w, r)
			} else if strings.ToLower(r.Method) == "post" {
				postHandleFunc(w, r)
			} else {
				w.WriteHeader(404)
			}

			done <- true
		}()

		select {
		case <-done:
			return
		case <-time.After(400 * time.Millisecond):
			w.WriteHeader(418)
		}
	})

	log.Println("Running API on :8080")
	http.ListenAndServe(":8080", nil)
}
