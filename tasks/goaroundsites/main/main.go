package main

import (
	"context"
	"fmt"
	"go_task/tasks/goaroundsites/src"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	sites := []string{
		"https://www.avito.ru/",
		"https://www.ozon.ru/",
		"https://vk.com/",
		"https://yandex.ru/",
		"https://www.google.com/",
	}

	monitor := src.NewMonitor(sites, time.Millisecond*200)

	if err := monitor.Run(ctx); err != nil {
		fmt.Println("error:", err)
	}
}
