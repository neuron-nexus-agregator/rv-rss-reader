package implementation

import (
	"context"
	"encoding/xml"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"gafarov/rss-reader/internal/core/cache"
	"gafarov/rss-reader/internal/model/rss"
)

type RssReader struct {
	cache    cache.ICache
	lastGuid string
	output   chan rss.Item
	stopOnce sync.Once
	stopChan chan struct{}
	feeds    map[string]struct{}
	client   http.Client
	mu       sync.Mutex
	wg       sync.WaitGroup
	isStoped *atomic.Bool
}

func New(cache cache.ICache) *RssReader {

	isStoped := atomic.Bool{}
	isStoped.Store(false)

	return &RssReader{
		cache:    cache,
		lastGuid: "",
		output:   make(chan rss.Item, 500),
		feeds:    make(map[string]struct{}),
		stopChan: make(chan struct{}),
		client:   http.Client{Timeout: 10 * time.Second},
		isStoped: &isStoped,
	}
}

func (r *RssReader) Stop() error {
	if r.isStoped.Load() {
		return ErrClosed
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
		return ErrClosed
	}

	if r.isInProcessOrRegister(url) {
		return ErrAlreadyStarted
	}

	err := r.startOnce(url, ctx)
	if err != nil && err != ErrNoItemsFound {
		return err
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
				err := r.startOnce(url, ctx)
				if err != nil && err != ErrNoItemsFound {
					return
				}
			}
		}
	}(url, delay, ctx)
	return nil
}

func (r *RssReader) startOnce(url string, ctx context.Context) error {
	var lastGuid string = ""

	if r.lastGuid == "" {
		rLastGuid, err := r.getLastReadGuid()
		if err != nil {
			lastGuid = ""
		}
		lastGuid = rLastGuid
	} else {
		lastGuid = r.lastGuid
	}

	items, err := r.ParseOnce(url, ctx)
	if err == ErrNoItemsFound {
		return nil
	} else if err != nil {
		return err
	}

	if lastGuid == "" && len(items) > 0 {
		lastGuid = items[0].Guid
		r.lastGuid = lastGuid
		err = r.saveLastReadGuid(lastGuid)
		if err != nil {
			// TODO handle error
		}
		return nil
	}

	for _, item := range items {

		if item.Guid == lastGuid {
			break
		}

		select {
		case r.output <- *item:
		default:
			// skip if channel is blocked
			continue
		}
	}

	if len(items) > 0 {
		r.lastGuid = items[0].Guid
		err = r.saveLastReadGuid(r.lastGuid)
		if err != nil {
			// TODO handle error
		}
	}

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
		return nil, ErrNoItemsFound
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
