package main

import (
	"net/http"

	"github.com/gronnbeck/ndc2017/domain"
	"github.com/mediocregopher/radix.v2/pool"
)

func get(client *pool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		has := hasValue(client, "current-value")

		if !has {
			w.Write(domain.NewErrorResponse("no value has been intialized").JSON())
		} else {
			value := getValue(client, "current-value")
			w.Write(domain.NewResponse(value).JSON())
		}
	}
}
