package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mediocregopher/radix.v2/pool"
)

var (
	redisAddr = flag.String("redis-addr", "localhost:6379", "tcp addr to connect redis tor")
	redisPass = flag.String("redis-pass", "", "password to redis server")
)

func init() {
	flag.Parse()

	if redisAddr == nil || *redisAddr == "" {
		panic("Reids addr cannot be empty")
	}
}

func main() {
	var client *pool.Pool
	getHandleFunc := get(client)
	postHandleFunc := post(client)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		done := make(chan bool)
		go func() {
			if strings.ToLower(r.Method) == "get" {
				getHandleFunc(w, r)
			} else if strings.ToLower(r.Method) == "post" {
				postHandleFunc(w, r)
			} else {
				w.WriteHeader(404)
			}
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
