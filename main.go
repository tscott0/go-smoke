package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	ui "github.com/gizak/termui"
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
	status   string
	respTime time.Duration
}

func main() {

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	table1 := ui.NewTable()
	table1.FgColor = ui.ColorWhite
	table1.BgColor = ui.ColorDefault
	table1.Y = 0
	table1.X = 0
	table1.Width = 150
	table1.Height = 4

	var config tomlConfig
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		//fmt.Println(err)
		return
	}

	var results []*result
	results = make([]*result, 0)

	var wg sync.WaitGroup
	var queue = make(chan string, 1)

	for _, endpoint := range config.Endpoint {
		//fmt.Printf("Server: %s (%s, %s, %d)\n", name, endpoint.Name, endpoint.URL, endpoint.Timeout)
		// Increment the WaitGroup counter.
		wg.Add(1)

		// initialise a new result struct to wrap the endpoint
		r := result{endpoint, "-", 0}
		results = append(results, &r)

		// Launch a goroutine to fetch the URL.
		go r.hitURL(&wg, queue)
	}

	go func() {
		wg.Wait()
		close(queue)
	}()

	// build layout
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, table1)))

	// calculate layout
	ui.Body.Align()

	rows1 := make([][]string, 1+len(config.Endpoint))
	rows1[0] = []string{"Name", "URL", "Timeout", "Response", "Duration"}

	draw := func(t int) {
		table1.Height = 3 + (2 * len(results))
		i := 1
		for _, r := range results {

			rows1[i] = []string{r.Name, r.URL, strconv.Itoa(r.Timeout), r.status, r.respTime.String()}
			i++
		}
		table1.Rows = rows1
		ui.Render(ui.Body)
	}

	// TODO: Bit of a hack. Not sure why it doesn't draw immediately
	draw(0)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Handle("/timer/1s", func(e ui.Event) {
		t := e.Data.(ui.EvtTimer)
		draw(int(t.Count))
	})

	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Clear()
		ui.Render(ui.Body)
	})

	ui.Loop()

	// Range over queue channel to drain and print the output to screen
	//for s := range queue {
	//fmt.Println(s)
	//_ = s
	//}

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
		//color.Red(r.Name + " (" + r.URL + "): " + err.Error())
		return
	}

	statusLine := format(resp, r, duration)
	q <- statusLine

}

func format(resp *http.Response, r *result, dur time.Duration) string {
	var status string

	ep := r.endpoint

	r.status = resp.Status
	r.respTime = dur

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
