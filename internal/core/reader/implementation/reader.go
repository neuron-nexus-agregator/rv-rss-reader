package implementation

import (
	"context"
	"encoding/xml"
	"errors"
	"gafarov/rss-reader/internal/model/rss"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var ClosedError error = errors.New("reader is closed")
var NoItemsFoundError error = errors.New("no items found")
var AlreadyStartedError error = errors.New("already started")

type RssReader struct {
	output   chan rss.Item
	stopOnce sync.Once
	stopChan chan struct{}
	feeds    map[string]struct{}
	client   http.Client
	mu       sync.Mutex
	wg       sync.WaitGroup
	isStoped *atomic.Bool
}

func New() *RssReader {

	isStoped := atomic.Bool{}
	isStoped.Store(false)

	return &RssReader{
		output:   make(chan rss.Item, 100),
		feeds:    make(map[string]struct{}),
		stopChan: make(chan struct{}),
		client:   http.Client{Timeout: 10 * time.Second},
		isStoped: &isStoped,
	}
}

func (r *RssReader) Stop() error {
	if r.isStoped.Load() {
		return ClosedError
	}
	r.close()
	return nil
}

func (r *RssReader) Output() <-chan rss.Item {
	return r.output
}

func (r *RssReader) IsStopped() bool {
	return r.isStoped.Load()
}

func (r *RssReader) close() {
	r.stopOnce.Do(func() {
		r.isStoped.Store(true)
		close(r.stopChan)
		r.wg.Wait()
		close(r.output)
		r.feeds = make(map[string]struct{})
	})
}

func (r *RssReader) isInProcessOrRegister(url string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.feeds[url]
	if !ok {
		r.feeds[url] = struct{}{}
		return false
	}
	return true
}

func (r *RssReader) StartParsing(url string, delay time.Duration, ctx context.Context) error {

	if r.IsStopped() {
		return ClosedError
	}

	if r.isInProcessOrRegister(url) {
		return AlreadyStartedError
	}

	r.wg.Add(1)
	go func(url string, delay time.Duration, ctx context.Context) {
		defer r.wg.Done()
		ticker := time.NewTicker(delay)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				r.mu.Lock()
				delete(r.feeds, url)
				r.mu.Unlock()
				return
			case <-r.stopChan:
				return
			case <-ticker.C:
				items, err := r.ParseOnce(url, ctx)
				if err == NoItemsFoundError {
					continue
				}
				for _, item := range items {
					select {
					case r.output <- *item:
					default:
						// skip in channel is blocked
						continue
					}
				}
			}
		}
	}(url, delay, ctx)
	return nil
}

func (r *RssReader) ParseOnce(url string, ctx context.Context) ([]*rss.Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	response, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var channel rss.Rss
	if err := xml.NewDecoder(response.Body).Decode(&channel); err != nil {
		return nil, err
	}

	var items []*rss.Item

	if len(channel.Channel.Items) == 0 {
		return nil, NoItemsFoundError
	}

	for i := range channel.Channel.Items {
		itm := &channel.Channel.Items[i]
		date, parseErr := ParseRSSDate(itm.PubDate)
		if parseErr == nil {
			itm.PubTimeParsed = date
		}
		items = append(items, itm)
	}

	return items, nil
}
