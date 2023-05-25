package src

import (
	"context"
	"fmt"
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

func (m *Monitor) Run(ctx context.Context) error {
	// Run printStatuses(ctx) and checkSite(ctx) for m.Sites
	// Renew sites requests to map every m.RequestFrequency
	// Return if context closed

	for _, site := range m.Sites {
		site := site
		m.G.Go(func() error {
			ticker := time.NewTicker(m.RequestFrequency)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					m.checkSite(ctx, site)
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}

	m.G.Go(func() error {
		return m.printStatuses(ctx)
	})

	if err := m.G.Wait(); err != nil && err != context.Canceled {
		return err
		// log.Fatalf("Fatal err from errG: %s", err.Error())
	}

	return nil
}

func (m *Monitor) checkSite(ctx context.Context, site string) {
	// with http client go through site and write result to m.StatusMap
	client := http.Client{
		Timeout: Timeout,
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, site, nil)
	if err != nil {
		fmt.Println("Err", err)
	}

	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	m.Mtx.Lock()
	m.StatusMap[site] = SiteStatus{
		Name:          site,
		StatusCode:    resp.StatusCode,
		TimeOfRequest: time.Now().Truncate(time.Second),
	}
	m.Mtx.Unlock()

}

func (m *Monitor) printStatuses(ctx context.Context) error {
	// print results of m.Status every second of until ctx cancelled
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			m.Mtx.Lock()
			for _, s := range m.StatusMap {
				fmt.Println(s.Name, s.StatusCode, s.TimeOfRequest)
			}
			m.Mtx.Unlock()
		}
	}
}
