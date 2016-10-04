package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
)

type tomlConfig struct {
	Endpoints map[string]endpoint
}

type endpoint struct {
	Name    string
	URL     string
	Timeout int
}

func main() {

	var config tomlConfig
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println(err)
		return
	}

	var wg sync.WaitGroup
	var queue = make(chan string, 1)

	for _, endpoint := range config.Endpoints {
		//fmt.Printf("Server: %s (%s, %s, %d)\n", name, endpoint.Name, endpoint.URL, endpoint.Timeout)
		// Increment the WaitGroup counter.
		wg.Add(1)
		// req will be overwritten. take a copy for each iteration
		e := endpoint
		// Launch a goroutine to fetch the URL.
		go e.hitURL(&wg, queue)
	}

	go func() {
		wg.Wait()
		close(queue)
	}()

	// Range over queue channel to drain and print the output to screen
	for s := range queue {
		fmt.Println(s)
	}

}

func (e *endpoint) hitURL(wg *sync.WaitGroup, q chan string) {
	// Decrement the counter when the goroutine completes.
	// Defer to allow the goroutine to fail
	defer wg.Done()

	timeout := time.Duration(time.Duration(e.Timeout) * time.Millisecond)
	client := http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	t0 := time.Now()
	resp, err := client.Get(e.URL)
	duration := time.Now().Sub(t0)

	if err != nil {
		color.Red(e.Name + " (" + e.URL + "): " + err.Error())
		return
	}

	s := resp.Status

	outputString := fmt.Sprintf("%s | %s | %s | %v", resp.Status, e.Name, e.URL, duration)

	switch {
	case strings.HasPrefix(s, "1"): // 1XX Info
		q <- color.CyanString(outputString)
	case strings.HasPrefix(s, "2"): // 2XX Success
		q <- color.GreenString(outputString)
	case strings.HasPrefix(s, "3"): // 3XX Redirect
		q <- color.MagentaString(outputString)
	case strings.HasPrefix(s, "4"): // 4XX Client Erorr
		q <- color.RedString(outputString)
	case strings.HasPrefix(s, "5"): // 5XX Server Error
		q <- color.RedString(outputString)
	default:
		q <- color.WhiteString(outputString)
	}

}
