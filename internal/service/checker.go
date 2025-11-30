package service

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"servicehealthchecker/internal/storage"
)

type Checker struct {
	storage *storage.Storage
	client  *http.Client
}

func NewChecker(s *storage.Storage) *Checker {
	return &Checker{
		storage: s,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *Checker) normalizeURL(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

func (c *Checker) checkURL(url string) string {
	normalizedURL := c.normalizeURL(url)

	resp, err := c.client.Get(normalizedURL)
	if err != nil {
		return "not_available"
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "available"
	}

	return "not_available"
}

func (c *Checker) CheckLinks(links []string) (map[string]string, int, error) {
	id, err := c.storage.CreateLinkSet(links)
	if err != nil {
		return nil, 0, err
	}

	statuses := c.checkLinksParallel(links)

	if err := c.storage.UpdateLinkSet(id, statuses); err != nil {
		return nil, 0, err
	}

	return statuses, id, nil
}

func (c *Checker) checkLinksParallel(links []string) map[string]string {
	var wg sync.WaitGroup
	var mu sync.Mutex

	statuses := make(map[string]string)

	for _, link := range links {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			status := c.checkURL(url)
			mu.Lock()
			statuses[url] = status
			mu.Unlock()
		}(link)
	}

	wg.Wait()
	return statuses
}

func (c *Checker) ProcessPendingTask(id int, links []string) error {
	statuses := c.checkLinksParallel(links)
	return c.storage.UpdateLinkSet(id, statuses)
}

