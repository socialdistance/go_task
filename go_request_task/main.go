package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Result struct {
	Response *http.Response
	Error    error
	URL      string
}

type ResultFile struct {
	URL    string
	Status string
}

const batchSize = 2
const timeout = 2 * time.Second

func checkStatus(wg *sync.WaitGroup, urls chan string) <-chan Result {
	retry := 3
	client := &http.Client{Timeout: timeout}

	resultsSuccesfull := make(chan Result)
	resultsFailed := make(chan Result, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(resultsFailed)
		for url := range urls {
			resp, err := client.Get(url)
			if err != nil {
				resultsFailed <- Result{Response: nil, Error: err, URL: url}
			} else {
				resultsSuccesfull <- Result{Response: resp, Error: nil, URL: url}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// defer close(resultsSuccesfull)
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
		}
	}()

	go func() {
		wg.Wait()
		// close(resultsFailed)
		close(resultsSuccesfull)
	}()

	return resultsSuccesfull
}

// func processBatch(list []ResultFile) {
// 	var wg sync.WaitGroup
// 	for _, i := range list {
// 		x := i
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			// do more complex things here
// 			writeFile(x)
// 		}()
// 	}
// 	wg.Wait()
// }

// func process(data []ResultFile) {
// 	for start, end := 0, 0; start <= len(data)-1; start = end {
// 		end = start + batchSize
// 		if end > len(data) {
// 			end = len(data)
// 		}
// 		batch := data[start:end]
// 		processBatch(batch)
// 	}
// 	fmt.Println("done processing all data")
// }

func batch(wg *sync.WaitGroup, data []ResultFile) {
	ch := make(chan struct{}, batchSize)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, i := range data {
			ch <- struct{}{}
			writeFile(i)
			<-ch
		}
	}()

	wg.Wait()
	fmt.Println("done processing all data")
}

func writeFile(line ResultFile) {
	f, err := os.OpenFile("file.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	buffer := bufio.NewWriterSize(f, 128)

	_, err = buffer.WriteString(line.URL + line.Status + "\n")
	if err != nil {
		log.Fatal(err)
	}

	if err := buffer.Flush(); err != nil {
		log.Fatal(err)
	}
}

func worker(wg *sync.WaitGroup, linkChan chan string, sliceResultToFile []ResultFile) {
	defer wg.Done()

	for response := range checkStatus(wg, linkChan) {
		if response.Error != nil {
			fmt.Printf("error: %v\n", response.Error)
			continue
		}

		fmt.Printf("Response: %s %v\n", response.URL, response.Response.Status)
		_ = response.Response.Body.Close()

		res := ResultFile{URL: response.URL, Status: response.Response.Status}

		sliceResultToFile = append(sliceResultToFile, res)
	}

	batch(wg, sliceResultToFile)
}

func main() {
	sliceResultToFile := []ResultFile{}
	wg := &sync.WaitGroup{}
	linkCh := make(chan string)

	urls := []string{"https://www.bbas.com", "https://badhost", "https://someOtherSite", "https://www.avito.ru/", "https://www.ozon.ru/", "https://vk.com/", "https://yandex.ru/", "https://www.google.com/", "https://github.com/", "http://medium.com/", "https://golang.org/"}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			// defer wg.Done()
			worker(wg, linkCh, sliceResultToFile)
		}()
	}

	for _, url := range urls {
		linkCh <- url
	}

	close(linkCh)
	wg.Wait()
}
