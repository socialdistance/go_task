package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Result struct {
	Response *http.Response
	Error    error
	Status   string
	// add retry field ?
}

// Ретраи перекладывать в другую горутину и делать запросы в другой горутине ?
// Писать 200 ОК в один канал, при попытках ретарая писать результат в другой канал, после слить их в один и вернуть
func checkStatus(done <-chan struct{}, urls ...string) <-chan Result {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resultsSuccesfull := make(chan Result)

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 60 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 60 * time.Second,
	}

	client := &http.Client{
		Timeout:   time.Second * 60,
		Transport: netTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	go func() {
		defer close(resultsSuccesfull)
		for _, url := range urls {
			var result Result
			// resp, err := http.Get(url)
			// fmt.Println(err)
			req, _ := http.NewRequest("GET", url, nil)
			req = req.WithContext(ctx)
			resp, err := client.Do(req)
			fmt.Println(resp, err)
			result = Result{Response: resp, Error: err, Status: resp.Status}
			// fmt.Println("Status:", result.Response.Status)
			// if result.Response.Status == "200 OK" {
			// 	fmt.Println("Result Success:", result)
			// 	resultsSuccesfull <- result
			// } else {
			// 	fmt.Println("Result Failed:", result)
			// 	resultsFailed <- result
			// }
			select {
			// case <-done:
			// 	return
			case <-time.After(10 * time.Second):
				fmt.Println("overslept")
			case <-ctx.Done():
				fmt.Println("ERR", ctx.Err()) // prints "context deadline exceeded"
			case resultsSuccesfull <- result:
				fmt.Println("Result:", result)
			}
		}
	}()

	// <-resultsFailed

	return resultsSuccesfull
}

func main() {
	done := make(chan struct{})
	defer close(done)

	urls := []string{"https://www.bbas.com", "https://www.avito.ru/", "https://www.ozon.ru/", "https://vk.com/", "https://yandex.ru/", "https://www.google.com/", "https://github.com/", "http://medium.com/", "https://golang.org/"}
	for response := range checkStatus(done, urls...) {
		fmt.Printf("Response: %v\n", response.Response.Status)
		_ = response.Response.Body.Close()
	}
}
