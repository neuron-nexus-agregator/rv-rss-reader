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

	"go.uber.org/zap"
)

type RssReader struct {
	cache     cache.ICache
	output    chan rss.Item
	stopOnce  sync.Once
	stopChan  chan struct{}
	feeds     map[string]struct{}
	client    http.Client
	mu        sync.Mutex
	wg        sync.WaitGroup
	isStarted *atomic.Bool
	isStoped  *atomic.Bool
	logger    *zap.Logger
}

func New(cache cache.ICache, logger *zap.Logger) *RssReader {

	isStoped := atomic.Bool{}
	isStoped.Store(false)

	isStarted := atomic.Bool{}
	isStarted.Store(false)

	return &RssReader{
		cache:     cache,
		output:    make(chan rss.Item, 500),
		feeds:     make(map[string]struct{}),
		stopChan:  make(chan struct{}),
		client:    http.Client{Timeout: 10 * time.Second},
		isStarted: &isStarted,
		isStoped:  &isStoped,
		logger:    logger,
	}
}

func (r *RssReader) Stop() error {
	if r.isStoped.Load() {
		if r.logger != nil {
			r.logger.Error("reader already stopped", zap.Error(ErrClosed))
		}
		return ErrClosed
	}

	if r.logger != nil {
		r.logger.Info("stopping reader")
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

		if r.logger != nil {
			r.logger.Info("reader stopped")
		}
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

func (r *RssReader) StartParsing(url, name string, delay time.Duration, ctx context.Context) error {

	if r.IsStopped() {
		if r.logger != nil {
			r.logger.Error("reader already stopped", zap.Error(ErrClosed))
		}
		return ErrClosed
	}

	if r.isInProcessOrRegister(url) {
		if r.logger != nil {
			r.logger.Error("already parsing", zap.String("url", url))
		}
		return ErrAlreadyStarted
	} else if r.logger != nil {
		r.logger.Info("starting parsing", zap.String("url", url))
	}

	err := r.startOnce(url, name, ctx)
	if err != nil && err != ErrNoItemsFound {
		if r.logger != nil {
			r.logger.Error("failed to start parsing", zap.String("url", url), zap.Error(err))
		}
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
				if r.logger != nil {
					r.logger.Info("parsing stopped", zap.String("url", url))
				}
				return
			case <-r.stopChan:
				if r.logger != nil {
					r.logger.Info("reader stopped")
				}
				return
			case <-ticker.C:
				err := r.startOnce(url, name, ctx)
				if err != nil && err != ErrNoItemsFound {
					if r.logger != nil {
						r.logger.Error("failed to parsing", zap.String("url", url), zap.Error(err))
					}
				}
			}
		}
	}(url, delay, ctx)
	return nil
}

func (r *RssReader) startOnce(url, name string, ctx context.Context) error {
	items, err := r.ParseOnce(url, ctx)
	if err == ErrNoItemsFound {
		return nil
	} else if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	defer wg.Wait()

	if !r.isStarted.Load() {
		for _, item := range items {
			wg.Go(func() {
				if err := r.saveReadGuid(item.Guid, name, 3*24*time.Hour); err != nil {
					if r.logger != nil {
						r.logger.Error("failed to save last read guid", zap.Error(err))
					}
				}
			})
		}
		r.isStarted.Store(true)
	}

	for _, item := range items {

		if isProcessed, err := r.isProcessed(item.Guid, name); err != nil {
			if r.logger != nil {
				r.logger.Error("failed to check if item is processed", zap.Error(err))
			}
		} else if isProcessed {
			continue
		}

		select {
		case r.output <- *item:
			if err := r.saveReadGuid(item.Guid, name, 14*24*time.Hour); err != nil {
				if r.logger != nil {
					r.logger.Error("failed to save last read guid", zap.Error(err))
				}
			}
		default:
			continue
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

func (r *RssReader) GetChannel(url string, ctx context.Context) (*rss.Channel, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	response, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var channel rss.Rss
	if err := xml.NewDecoder(response.Body).Decode(&channel); err != nil {
		return nil, err
	}

	channel.Channel.Items = nil

	return &channel.Channel, nil
}
