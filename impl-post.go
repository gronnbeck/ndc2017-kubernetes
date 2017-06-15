package main

import (
	"net/http"

	"github.com/mediocregopher/radix.v2/pool"
)

func post(client *pool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
