package client

import (
	"io/ioutil"
	"log"
	"net/http"
)

func Client() {
	res, err := http.Get("http://localhost:9000/api/up/v1")
	if err != nil {
		log.Println(err)
		return
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("%d: %s", res.StatusCode, string(data))
}
