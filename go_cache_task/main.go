package main

import (
	"container/heap"
	"container/list"
	"fmt"
	"sync"
	"time"
)

type Node struct {
	value      any
	lastAccess int64
	keyPtr     *list.Element
}

type Cache struct {
	capacity int
	queue    *list.List
	storage  map[string]*Node
}

type itemQueue struct {
	key        string
	value      any
	priority   time.Time
	queueIndex int
}

type CacheMultithreading struct {
	mu        sync.RWMutex
	cache     Cache
	priorityQ priorityQueue
}

type priorityQueue struct {
	items []*itemQueue
}

func (pq priorityQueue) Len() int {
	length := len(pq.items)
	return length
}

// Less will consider items with time.Time default value (epoch start) as more than set items.
func (pq priorityQueue) Less(i, j int) bool {
	if pq.items[i].priority.IsZero() {
		return false
	}
	if pq.items[j].priority.IsZero() {
		return true
	}
	return pq.items[i].priority.Before(pq.items[j].priority)
}

func (pq priorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].queueIndex = i
	pq.items[j].queueIndex = j
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*itemQueue)
	item.queueIndex = len(pq.items)
	pq.items = append(pq.items, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	item.queueIndex = -1
	pq.items = old[0 : n-1]
	return item
}

func New(capacity int) *CacheMultithreading {
	priorityQ := &priorityQueue{}

	return &CacheMultithreading{
		mu: sync.RWMutex{},
		cache: Cache{
			capacity: capacity,
			queue:    list.New(),
			storage:  make(map[string]*Node, capacity),
		},
		priorityQ: *priorityQ,
	}
}

func (c *CacheMultithreading) Set(key string, value any) {
	c.mu.Lock()
	i := 0
	if item, ok := c.cache.storage[key]; !ok {
		if c.cache.capacity == len(c.cache.storage) {
			lastItem := c.cache.queue.Back()
			// itemQ := itemQueue{
			// 	key: key, value: lastItem.Value, queueIndex: 0,
			// }

			c.cache.queue.Remove(lastItem)
			delete(c.cache.storage, lastItem.Value.(string))
			c.priorityQ.Pop()
		}
		c.cache.storage[key] = &Node{value: value, keyPtr: c.cache.queue.PushFront(key), lastAccess: time.Now().Unix()}
		heap.Push(&c.priorityQ, &itemQueue{key: key, value: value, priority: time.Now(), queueIndex: i})
	} else {
		item.value = value
		item.lastAccess = time.Now().Unix()
		c.cache.storage[key] = item
		c.cache.queue.MoveToFront(item.keyPtr)
		heap.Push(&c.priorityQ, &itemQueue{key: key, value: value, priority: time.Now(), queueIndex: i + 1})
	}
	c.mu.Unlock()
}

func (c *CacheMultithreading) Get(key string) (value any, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	i := 0
	item, ok := c.cache.storage[key]
	if !ok {
		return nil, false
	}

	c.cache.storage[key] = &Node{
		value:      item.value,
		lastAccess: time.Now().Unix(),
		keyPtr:     c.cache.queue.PushFront(key),
	}

	c.cache.queue.MoveToFront(item.keyPtr)
	heap.Push(&c.priorityQ, &itemQueue{key: key, value: value, priority: time.Now(), queueIndex: i + 1})

	return item.value, ok
}

func (c *CacheMultithreading) Delete(key string) {
	c.mu.Lock()
	c.deleteMethod(key)
	c.mu.Unlock()
}

func (c *CacheMultithreading) deleteMethod(key string) {
	delete(c.cache.storage, key)
}

func (c *CacheMultithreading) clearTTL(ticker *time.Ticker, tickerStopCh chan struct{}) {
loop:
	for {
		select {
		case now := <-ticker.C:
			c.mu.Lock()
			heap.Init(&c.priorityQ)
			if len(c.cache.storage) == 0 && c.priorityQ.Len() == 0 {
				tickerStopCh <- struct{}{}
			}

			for _, v := range c.priorityQ.items {
				if now.Unix()-v.priority.Unix() > 1 {
					fmt.Println("Delete", v.key, v.value)
					fmt.Printf("%+v\n", v)
					c.deleteMethod(v.key)
					c.priorityQ.Pop()
				}

			}
			c.mu.Unlock()
		case <-tickerStopCh:
			ticker.Stop()
			break loop
		}
	}
}

func main() {
	mainRace()
}

func mainRace() {
	cache := New(2)
	wg := &sync.WaitGroup{}
	ticker := time.NewTicker(1 * time.Second)
	tickerStopCh := make(chan struct{}, 1)

	cache.Set("name", "Alex")
	cache.Set("hobby", "BJJ")
	cache.Set("Boba", "Biba")

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println(cache.Get("name"))
		fmt.Println(cache.Get("hobby"))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.Delete("hobby")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println(cache.Get("name"))
		fmt.Println(cache.Get("hobby"))
		fmt.Println(cache.Get("Boba"))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.clearTTL(ticker, tickerStopCh)
	}()

	wg.Wait()
}
