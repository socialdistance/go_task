package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	Response *http.Response
	Error    error
	URL      string
}

func checkStatus(done chan struct{}, urls ...string) <-chan Result {
	retry := 3
	wg := &sync.WaitGroup{}
	client := &http.Client{Timeout: 2 * time.Second}

	resultsSuccesfull := make(chan Result)
	resultsFailed := make(chan Result, 1)
	// testCh := make(chan struct{})

	// wg.Add(1)
	// go func() {
	// 	// defer wg.Done()
	// 	defer close(resultsSuccesfull)
	// 	// defer close(testCh)
	// 	for _, url := range urls {
	// 		resp, err := client.Get(url)
	// 		if err != nil {
	// 			resultsFailed <- Result{Response: nil, Error: err, URL: url}
	// 			fmt.Println("failed")
	// 			// testCh <- struct{}{}
	// 		} else {
	// 			resultsSuccesfull <- Result{Response: resp, Error: nil, URL: url}
	// 			fmt.Println("success")
	// 		}
	// 	}

	// }()

	// // for retry requests
	// // wg.Add(1)
	// go func() {
	// 	// defer wg.Done()
	// 	defer close(resultsFailed)
	// 	// select {
	// 	// case <-testCh:
	// 	for urlFailed := range resultsFailed {
	// 		// testCh <- struct{}{}
	// 		fmt.Println("URL", urlFailed)
	// 		for i := 0; i < retry; i++ {
	// 			resp, err := client.Get(urlFailed.URL)
	// 			if err != nil {
	// 				fmt.Printf("WARNING: Error job in attempt no %d: %s - retrying...\n", i+1, err)
	// 			} else {
	// 				resultsSuccesfull <- Result{Response: resp, Error: err}
	// 				continue
	// 			}
	// 		}
	// 		<-resultsFailed
	// 	}
	// 	// }
	// }()

	wg.Add(len(urls))
	go func() {
		defer wg.Done()
		defer close(resultsSuccesfull)
		for _, url := range urls {
			resp, err := client.Get(url)
			if err != nil {
				resultsFailed <- Result{Response: nil, Error: err, URL: url}
				fmt.Println("failed")
				// testCh <- struct{}{}
			} else {
				resultsSuccesfull <- Result{Response: resp, Error: nil, URL: url}
				fmt.Println("success")
			}
		}
	}()

	wg.Add(len(resultsFailed))
	go func() {
		defer wg.Done()
		defer close(resultsFailed)
		for urlFailed := range resultsFailed {
			for i := 0; i < retry; i++ {
				resp, err := client.Get(urlFailed.URL)
				if err != nil {
					fmt.Printf("WARNING: Error job in attempt no %d: %s - retrying...\n", i+1, err)
				} else {
					resultsSuccesfull <- Result{Response: resp, Error: err}
					continue
				}
			}
			<-resultsFailed
		}
	}()

	go func() {
		// close(resultsSuccesfull)
		// close(resultsFailed)
		wg.Wait()
	}()
	// <-testCh
	return resultsSuccesfull
}

func main() {
	done := make(chan struct{})
	defer close(done)

	// "https://someOtherSite.com", "https://www.bbas.com", "https://badhost", "https://www.avito.ru/", "https://www.ozon.ru/"
	urls := []string{"https://badhost", "https://someOtherSite", "https://www.avito.ru/", "https://www.ozon.ru/", "https://vk.com/", "https://yandex.ru/", "https://www.google.com/", "https://github.com/", "http://medium.com/", "https://golang.org/"}
	for response := range checkStatus(done, urls...) {
		// time.Sleep(5 * time.Second)
		if response.Error != nil {
			fmt.Printf("error: %v\n", response.Error)
			continue
		}
		fmt.Printf("Response: %v\n", response.Response.Status)
		_ = response.Response.Body.Close()
	}
}
