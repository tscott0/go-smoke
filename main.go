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

	for _, endpoint := range config.Endpoints {
		//fmt.Printf("Server: %s (%s, %s, %d)\n", name, endpoint.Name, endpoint.URL, endpoint.Timeout)
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
		color.Cyan(outputString)
	case strings.HasPrefix(s, "2"): // 2XX Success
		color.Green(outputString)
	case strings.HasPrefix(s, "3"): // 3XX Redirect
		color.Magenta(outputString)
	case strings.HasPrefix(s, "4"): // 4XX Client Erorr
		color.Red(outputString)
	case strings.HasPrefix(s, "5"): // 5XX Server Error
		color.Red(outputString)
	default:
		color.White(outputString)
	}

}
