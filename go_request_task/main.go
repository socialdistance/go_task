package main

import (
	"errors"
	"fmt"
	"net/http"
	"syscall"
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
	resultsSuccesfull := make(chan Result)

	go func() {
		defer close(resultsSuccesfull)
		for _, url := range urls {
			var result Result
			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Get(url)
			if errors.Is(err, syscall.ECONNREFUSED) {
				fmt.Println("Timeout:", url)
				result = Result{Error: err, Response: nil}
				continue
			}
			if err != nil {
				result = Result{Error: err, Response: nil}
			} else {
				result = Result{Error: nil, Response: resp}
			}
			select {
			case <-done:
				return
			case resultsSuccesfull <- result:
			}
		}
	}()

	// <-resultsFailed

	return resultsSuccesfull
}

func main() {
	done := make(chan struct{})
	defer close(done)

	urls := []string{"https://www.bbas.com", "https://badhost", "https://www.avito.ru/", "https://www.ozon.ru/", "https://vk.com/", "https://yandex.ru/", "https://www.google.com/", "https://github.com/", "http://medium.com/", "https://golang.org/"}
	for response := range checkStatus(done, urls...) {
		if response.Error != nil {
			fmt.Printf("error: %v\n", response.Error)
			continue
		}
		fmt.Printf("Response: %v\n", response.Response.Status)
		_ = response.Response.Body.Close()
	}
}
