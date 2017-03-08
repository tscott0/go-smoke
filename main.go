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

type endpoint struct {
	Name    string
	URL     string
	Timeout int
}

type tomlConfig struct {
	Endpoint []endpoint
}

type result struct {
	endpoint
	resp     *http.Response
	respTime time.Duration
}

func main() {

	var config tomlConfig
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println(err)
		return
	}

	var results []result
	results = make([]result, 0)

	var wg sync.WaitGroup
	var queue = make(chan string, 1)

	for _, endpoint := range config.Endpoint {
		//fmt.Printf("Server: %s (%s, %s, %d)\n", name, endpoint.Name, endpoint.URL, endpoint.Timeout)
		// Increment the WaitGroup counter.
		wg.Add(1)

		// initialise a new result struct to wrap the endpoint
		r := result{endpoint, nil, 0}
		results = append(results, r)

		// Launch a goroutine to fetch the URL.
		go r.hitURL(&wg, queue)
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

func (r *result) hitURL(wg *sync.WaitGroup, q chan string) {
	// Decrement the counter when the goroutine completes.
	// Defer to allow the goroutine to fail
	defer wg.Done()

	timeout := time.Duration(time.Duration(r.Timeout) * time.Millisecond)
	client := http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	t0 := time.Now()

	// TODO: Add support for http methods here
	resp, err := client.Get(r.URL)
	duration := time.Now().Sub(t0)

	if err != nil {
		color.Red(r.Name + " (" + r.URL + "): " + err.Error())
		return
	}

	statusLine := format(resp, r, duration)
	q <- statusLine

}

func format(resp *http.Response, r *result, dur time.Duration) string {
	var status string

	ep := r.endpoint

	switch {
	case strings.HasPrefix(resp.Status, "1"): // 1XX Info
		status = color.CyanString(resp.Status)
	case strings.HasPrefix(resp.Status, "2"): // 2XX Success
		status = color.GreenString(resp.Status)
	case strings.HasPrefix(resp.Status, "3"): // 3XX Redirect
		status = color.MagentaString(resp.Status)
	case strings.HasPrefix(resp.Status, "4"): // 4XX Client Erorr
		status = color.RedString(resp.Status)
	case strings.HasPrefix(resp.Status, "5"): // 5XX Server Error
		status = color.RedString(resp.Status)
	default:
		status = color.WhiteString(resp.Status)
	}

	return fmt.Sprintf("%s | %s | %s | %v", status, ep.Name, ep.URL, dur)

}
