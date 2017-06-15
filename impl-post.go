package main

import (
	"io/ioutil"
	"net/http"

	"github.com/gronnbeck/ndc2017/domain"
	"github.com/mediocregopher/radix.v2/pool"
)

func post(client *pool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		byt, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		req := domain.RequestFromBytes(byt)
		setValue(client, "current-value", req)
	}
}
