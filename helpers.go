package main

import (
	"github.com/gronnbeck/ndc2017/domain"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/pkg/errors"
)

func hasValue(client *pool.Pool, key string) bool {
	hasCmd := client.Cmd("HEXISTS", "currency", key)

	if hasCmd.Err != nil {
		panic(errors.Wrap(hasCmd.Err, "Has command failed"))
	}

	res, err := hasCmd.Int()
	if err != nil {
		panic(errors.Wrap(err, "Expected bool but was something else"))
	}

	return res != 0
}

func getValue(client *pool.Pool, key string) float64 {
	getCmd := client.Cmd("HGET", "currency", key)

	if getCmd.Err != nil {
		panic(errors.Wrap(getCmd.Err, "Could not get currency/current-value"))
	}

	res, err := getCmd.Float64()
	if err != nil {
		panic(errors.Wrap(err, "Unexpected value stored in currency/current-value"))
	}

	return res
}

func setValue(client *pool.Pool, key string, req domain.Request) {
	res := client.Cmd("HSET", "currency", key, req.Value)
	if res.Err != nil {
		panic(res.Err)
	}
}
