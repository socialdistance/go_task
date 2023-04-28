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

func checkStatus(urls ...string) <-chan Result {
	retry := 3
	wg := &sync.WaitGroup{}
	client := &http.Client{Timeout: 2 * time.Second}

	resultsSuccesfull := make(chan Result)
	resultsFailed := make(chan Result, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(resultsFailed)
		for _, url := range urls {
			resp, err := client.Get(url)
			if err != nil {
				resultsFailed <- Result{Response: nil, Error: err, URL: url}
				fmt.Println("failed")
			} else {
				resultsSuccesfull <- Result{Response: resp, Error: nil, URL: url}
				fmt.Println("success")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
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
		close(resultsSuccesfull)
	}()

	return resultsSuccesfull
}

func writeFile(lines []ResultFile) {
	f, err := os.Create("file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	buffer := bufio.NewWriterSize(f, 128)

	for _, line := range lines {
		_, err = buffer.WriteString(line.URL + line.Status + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := buffer.Flush(); err != nil {
		log.Fatal(err)
	}
}

const batchSize = 2

func process(data []ResultFile) {
	for start, end := 0, 0; start <= len(data)-1; start = end {
		end = start + batchSize
		if end > len(data) {
			end = len(data)
		}
		batch := data[start:end]
		writeFile(batch)
	}
	fmt.Println("done processing all data")
}

// func processBatch(list []ResultFile) {
// 	var wg sync.WaitGroup
// 	for _, i := range list {
// 		x := i
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			fmt.Println(x)
// 		}()
// 	}

// 	wg.Wait()
// }

func main() {
	sliceResultToFile := []ResultFile{}

	urls := []string{"https://www.bbas.com", "https://badhost", "https://someOtherSite", "https://www.avito.ru/", "https://www.ozon.ru/", "https://vk.com/", "https://yandex.ru/", "https://www.google.com/", "https://github.com/", "http://medium.com/", "https://golang.org/"}
	for response := range checkStatus(urls...) {
		if response.Error != nil {
			fmt.Printf("error: %v\n", response.Error)
			continue
		}
		fmt.Printf("Response: %s %v\n", response.URL, response.Response.Status)
		_ = response.Response.Body.Close()

		res := ResultFile{URL: response.URL, Status: response.Response.Status}

		sliceResultToFile = append(sliceResultToFile, res)
	}

	// writeFile(sliceResultToFile)
	process(sliceResultToFile)
}
