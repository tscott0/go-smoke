package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type GetRequest struct {
	url     string
	timeout int // in milliseconds
}

func main() {
	var wg sync.WaitGroup

	var gets = []GetRequest{
		{url: "http://www.golang.org/", timeout: 1000},
		{url: "http://www.google.com/", timeout: 1000},
		{url: "http://www.somestupidname.com/", timeout: 1000},
		{url: "http://www.openbet.com/", timeout: 1000},
		{url: "http://www.ladbrokes.com/", timeout: 1000},
		{url: "https://github.com/tscott0/saijgs", timeout: 1000},
	}

	for _, req := range gets {
		// Increment the WaitGroup counter.
		wg.Add(1)

		// req will be overwritten. take a copy for each iteration
		r := req

		// Launch a goroutine to fetch the URL.
		go r.hitURL(&wg)
	}
	// Wait for all HTTP fetches to complete.
	wg.Wait()
}

func (r *GetRequest) hitURL(wg *sync.WaitGroup) {
	// Decrement the counter when the goroutine completes.
	defer wg.Done()

	timeout := time.Duration(time.Duration(r.timeout) * time.Millisecond)
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
