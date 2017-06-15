package domain

import "encoding/json"

type Request struct {
	Value float64 `json:"value"`
}

func RequestFromBytes(byt []byte) Request {
	var r Request
	if err := json.Unmarshal(byt, &r); err != nil {
		panic(err)
	}
	return r
}

func (r Request) JSON() []byte {
	byt, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return byt
}
