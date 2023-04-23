package main

import (
	"fmt"
	"net/http"
)

type Result struct {
	Response *http.Response
	Error    error
	// add retry field ?
}

// Ретраи перекладывать в другую горутину и делать запросы в другой горутине ?
// Писать 200 ОК в один канал, при попытках ретарая писать результат в другой канал, после слить их в один и вернуть
func checkStatus(done <-chan struct{}, urls ...string) <-chan Result {
	results := make(chan Result)
	go func() {
		defer close(results)
		for _, url := range urls {
			var result Result
			resp, err := http.Get(url)
			result = Result{Error: err, Response: resp}
			if result.Response.Status == "200 OK" {
				select {
				case <-done:
					return
				case results <- result:
				}
			}
		}
	}()

	return results
}

func main() {
	done := make(chan struct{})
	defer close(done)

	urls := []string{"https://www.avito.ru/", "https://www.ozon.ru/", "https://vk.com/", "https://yandex.ru/", "https://www.google.com/"}
	for response := range checkStatus(done, urls...) {
		fmt.Printf("Response: %v\n", response.Response.Status)
		_ = response.Response.Body.Close()
	}
}
