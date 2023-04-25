package main

import (
	"fmt"
	"net/http"
	"time"
)

type Result struct {
	Response *http.Response
	Error    error
	URL      string
	// add retry field ?
}

// Ретраи перекладывать в другую горутину и делать запросы в другой горутине ?
// Писать 200 ОК в один канал, при попытках ретарая писать результат в другой канал, после слить их в один и вернуть
func checkStatus(done chan struct{}, urls ...string) <-chan Result {
	retry := 3
	client := &http.Client{Timeout: 2 * time.Second}

	resultsSuccesfull := make(chan Result)
	resultsFailed := make(chan Result)

	go func() {
		defer close(resultsSuccesfull)
		for _, url := range urls {
			var resultSuccess, resultFailed Result
			resp, err := client.Get(url)
			if err != nil {
				resultFailed = Result{Response: nil, Error: err, URL: url}
				fmt.Println("failed", resultFailed)
			} else {
				resultSuccess = Result{Response: resp, Error: nil, URL: url}
				fmt.Println("success", resultSuccess)
			}

			// Запросы продолжают делаться в горутине для ретраев потому что приходит
			// {nil nil}, а это тип Result с nil в полях
			select {
			// case <-done:
			// 	return
			case resultsSuccesfull <- resultSuccess:
			case resultsFailed <- resultFailed:
			}
		}
	}()

	// for retry requests
	go func() {
		var result Result
		defer close(resultsFailed)
		for urlFailed := range resultsFailed {
			for i := 0; i < retry; i++ {
				resp, err := client.Get(urlFailed.URL)
				if err != nil {
					fmt.Printf("WARNING: Error job in attempt no %d: %s - retrying...\n", i+1, err)
					continue
				} else {
					result = Result{Response: resp, Error: err}
				}
				select {
				case resultsSuccesfull <- result:
				default:
					fmt.Println("Default")
				}
			}
		}

	}()

	return resultsSuccesfull
}

func main() {
	done := make(chan struct{})
	defer close(done)

	// "https://someOtherSite.com", "https://www.bbas.com", "https://badhost",
	urls := []string{"https://badhost", "https://www.avito.ru/", "https://www.ozon.ru/", "https://vk.com/", "https://yandex.ru/", "https://www.google.com/", "https://github.com/", "http://medium.com/", "https://golang.org/"}
	for response := range checkStatus(done, urls...) {
		time.Sleep(5 * time.Second)
		if response.Error != nil {
			fmt.Printf("error: %v\n", response.Error)
			continue
		}
		fmt.Printf("Response: %v\n", response.Response.Status)
		_ = response.Response.Body.Close()
	}
}
