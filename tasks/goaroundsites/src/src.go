package src

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	// You can limit concurrent net request. It's optional
	MaxGoroutines = 1
	// timeout for net requests
	Timeout = 2 * time.Second
)

type SiteStatus struct {
	Name          string
	StatusCode    int
	TimeOfRequest time.Time
}

type Monitor struct {
	StatusMap        map[string]SiteStatus
	Mtx              *sync.Mutex
	G                errgroup.Group
	Sites            []string
	RequestFrequency time.Duration
}

func NewMonitor(sites []string, requestFrequency time.Duration) *Monitor {
	return &Monitor{
		StatusMap:        make(map[string]SiteStatus),
		Mtx:              &sync.Mutex{},
		Sites:            sites,
		RequestFrequency: requestFrequency,
	}
}

func (m *Monitor) String() string {
	return fmt.Sprintf("TEST: %s", m.Sites)
}

func (m *Monitor) Run(ctx context.Context) error {
	// Run printStatuses(ctx) and checkSite(ctx) for m.Sites
	// Renew sites requests to map every m.RequestFrequency
	// Return if context closed

	fmt.Println("SITES:", m.Sites)

	m.G.SetLimit(2)

	m.G.Go(func() error {
		for _, site := range m.Sites {
			// m.Mtx.Lock()
			// defer m.Mtx.Unlock()
			m.checkSite(ctx, site)
		}

		return nil
	})

	m.G.Go(func() error {
		err := m.printStatuses(ctx)

		return err
	})

	if err := m.G.Wait(); err != nil {
		log.Fatalf("Fatal err from errG: %s", err.Error())
	}

	return nil
}

func (m *Monitor) checkSite(ctx context.Context, site string) {
	// with http client go through site and write result to m.StatusMap
	fmt.Println("Check sites...")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, site, nil)
	if err != nil {
		fmt.Println("Err", err)
	}

	resp, err := http.DefaultClient.Do(request)

	if _, ok := m.StatusMap[site]; ok {
		m.Mtx.Lock()
		defer m.Mtx.Unlock()
		m.StatusMap[site] = SiteStatus{
			Name:          site,
			StatusCode:    resp.StatusCode,
			TimeOfRequest: time.Now(),
		}
	}

	m.Mtx.Lock()
	m.StatusMap[site] = SiteStatus{
		Name:          site,
		StatusCode:    resp.StatusCode,
		TimeOfRequest: time.Now(),
	}
	m.Mtx.Unlock()

}

func (m *Monitor) printStatuses(ctx context.Context) error {
	// print results of m.Status every second of until ctx cancelled
	ticker := time.NewTicker(1 * time.Second)

	select {
	case <-ctx.Done():
		fmt.Println("Context done for printStatuses")
		return nil
	case <-ticker.C:
		fmt.Println("Ticker...")
		// m.Mtx.Lock()
		for _, s := range m.StatusMap {
			fmt.Println(s.StatusCode)
		}
		// m.Mtx.Unlock()
	}

	return nil
}
