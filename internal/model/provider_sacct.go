package model

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
)

type SacctProvider struct {
	BaseProvider[*TableData]
	cache     *SacctCache
	lastFetch time.Time
}

func NewSacctProvider() *SacctProvider {
	cache, err := NewSacctCache()
	if err != nil {
		log.Fatalf("Failed to initialize sacct cache: %v", err)
	}

	p := SacctProvider{
		BaseProvider: BaseProvider[*TableData]{},
		cache:        cache,
		lastFetch:    time.Time{}, // Zero time indicates no fetch performed yet
	}
	return &p
}

// Fetches data into cached file. Does NOT make it available to user, use Fetch() for that.
func (p *SacctProvider) FetchToCache(since time.Duration) error {
	if since == 0 {
		logger.Debugf("Ignoring sacct cache refresh as '%s'=0", config.CONFIG_OPTION_NAME_LOAD_SACCT_CACHE_SINCE)
		return nil
	}

	// By default, we assume that we'll fetch data starting from the requested duration.
	fetchSince := since
	cacheStartTime := time.Now().Add(-fetchSince)
	rewriteEntireCache := true
	msg := ""

	// If this is not the first fetch, use a more efficient strategy
	isFirstFetch := p.lastFetch.IsZero()
	if !isFirstFetch {
		// For subsequent fetches, only get data since the last fetch
		// This avoids repeatedly fetching the entire duration
		sinceLastFetch := time.Since(p.lastFetch)
		if sinceLastFetch > 0 {
			fetchSince = sinceLastFetch
			rewriteEntireCache = false
			msg = fmt.Sprintf("subsequent fetch, only getting data since last fetch (%s ago)", sinceLastFetch.Truncate(time.Second))
		}
	} else if p.cache.IsUsable {
		// First fetch but we have a usable cache, check its state
		cacheAge := time.Since(p.cache.Content.StartTime).Truncate(time.Second)
		cacheEndAge := time.Since(p.cache.Content.EndTime).Truncate(time.Second)

		if since > cacheAge {
			// Requested duration is older than the cache start time, full refresh needed.
			rewriteEntireCache = true
			msg = fmt.Sprintf("requested refresh-since of (%s) is older than the cache start time (%s). Performing full refresh.", since.Truncate(time.Second), cacheAge)
		} else if cacheAge >= since && since >= cacheEndAge {
			// Requested duration falls within the cache range, additive fetch.
			rewriteEntireCache = false
			cacheStartTime = p.cache.Content.StartTime
			msg = fmt.Sprintf("requested refresh-since of (%s) falls within the cache range (%s to %s). Performing additive fetch.", since.Truncate(time.Second), cacheAge, cacheEndAge)
		} else if since < cacheEndAge {
			// Requested duration is more recent than the cache end time, partial fetch.
			rewriteEntireCache = false
			fetchSince = cacheEndAge
			cacheStartTime = p.cache.Content.StartTime
			msg = fmt.Sprintf("requested refresh-since of (%s) is more recent than the cache end time (%s). Adjusting fetch duration to %s.", since.Truncate(time.Second), cacheEndAge, fetchSince)
		}
	} else {
		// First fetch with no usable cache, full refresh needed.
		msg = fmt.Sprintf("first fetch, cache is not usable. Performing full refresh with requested duration (%s).", since.Truncate(time.Second))
	}

	logger.Debugf("sacct cache: %s", msg)

	// Get fresh data and cache it
	data, err := GetSacctData(fetchSince)
	if err != nil {
		p.updateError(err)
		return err
	}

	// Update the last fetch time
	p.lastFetch = time.Now()

	err = p.cache.WriteToCache(
		data,
		cacheStartTime, // from
		p.lastFetch,    // to
		rewriteEntireCache,
	)
	if err != nil {
		return err
	}
	logger.Debugf("sacct cache: updated range (%s - %s), updated %d rows", p.cache.Content.StartTime.Format(time.RFC3339), p.cache.Content.EndTime.Format(time.RFC3339), len(data.Rows))
	return nil
}

func (p *SacctProvider) Fetch() error {
	p.updateData(&TableData{
		Headers: &[]config.ColumnConfig{},
		Rows:    [][]string{},
	})
	data, err := p.cache.GetFromCache()
	if err != nil {
		return err
	}
	p.updateData(data)
	return nil
}

func (p *SacctProvider) FilteredData(filter string) *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()

	var rows [][]string
	for _, row := range data.Rows {
		// Ignore row if partition filter doesn't match
		if filter != "" {
			if !strings.Contains(row[config.SacctViewColumnsPartitionIndex], filter) {
				continue
			}
		}

		// Ignore row if state filter doesn't match
		if config.JobStateCurrentChoice != "(all)" {
			if !strings.Contains(row[config.SacctViewColumnsStateIndex], config.JobStateCurrentChoice) {
				continue
			}
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers: data.Headers,
		Rows:    rows,
	}
}
