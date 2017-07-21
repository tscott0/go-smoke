package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
)

type endpoint struct {
	Name    string
	URL     string
	Timeout int
}

type endpoints struct {
	Endpoint []endpoint
}

func main() {

	color.Cyan("Prints text in cyan.")

	var config endpoints
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println(err)
		return
	}

	var wg sync.WaitGroup

	for _, endpoint := range config.Endpoint {
		fmt.Printf("Server: (%s, %s, %d)\n", endpoint.Name, endpoint.URL, endpoint.Timeout)
		// Increment the WaitGroup counter.
		wg.Add(1)
		// req will be overwritten. take a copy for each iteration
		e := endpoint
		// Launch a goroutine to fetch the URL.
		go e.hitURL(&wg)
	}

	wg.Wait()

}

func (e *endpoint) hitURL(wg *sync.WaitGroup) {
	// Decrement the counter when the goroutine completes.
	defer wg.Done()

	timeout := time.Duration(time.Duration(e.Timeout) * time.Millisecond)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(e.URL)

	if err != nil {
		fmt.Println("Error while getting \"" + e.URL + "\": " + err.Error())
		return
	}

	fmt.Println(resp.Status + " from " + e.URL)
}
