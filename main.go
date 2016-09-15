package main

import (
	"fmt"
	"net/http"
	"time"
)

type GetRequest struct {
	url     string
	timeout int // in milliseconds
}

func main() {
	test := GetRequest{url: "http://www.google.co.uk", timeout: 1000}

	fmt.Println("TESTING")
	go test.hitURL()
}

func (r *GetRequest) hitURL() {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(r.url)

	if err != nil {
		fmt.Println("Error while getting \"" + r.url + "\": " + err.Error())
		return
	}

	fmt.Println(resp.Status + " from " + r.url)
}
