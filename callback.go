package main

import (
	"errors"
	"log"
	"net/http"
)

const (
	HTTP = 0
)

type Callback struct {
	protocol int
	url      string
}

func (c *Callback) Call() error {
	switch c.protocol {
	case HTTP:
		log.Println("http req url", c.url)
		resp, err := http.Get(c.url)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 300 {
			return errors.New("Bad status:" + resp.Status)
		}
	}
	return nil
}

func GetCallbacks(i interface{}) ([]Callback, error) {
	callbacks := make([]Callback, 0)
	switch v := i.(type) {
	case string:
		cb := Callback{protocol: HTTP, url: v}
		callbacks = append(callbacks, cb)
	default:
		log.Println("could not decode callback")
	}
	return callbacks, nil
}

func ExecuteCallbacks(cbks []Callback) (err error) {
	for _, cb := range cbks {
		err = cb.Call()
		if err != nil {
			return
		}

	}
	return
}
