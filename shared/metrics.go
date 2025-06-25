package shared

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics collects application metrics
type Metrics struct {
	mu              sync.RWMutex
	httpRequests    map[string]*atomic.Int64 // endpoint -> count
	dbQueries       *atomic.Int64
	userLookups     *atomic.Int64
	cacheHits       *atomic.Int64
	cacheMisses     *atomic.Int64
	requestDuration map[string][]time.Duration
	enabled         bool
}

// NewMetrics creates a new metrics collector
func NewMetrics(config *Config) *Metrics {
	return &Metrics{
		httpRequests:    make(map[string]*atomic.Int64),
		dbQueries:       &atomic.Int64{},
		userLookups:     &atomic.Int64{},
		cacheHits:       &atomic.Int64{},
		cacheMisses:     &atomic.Int64{},
		requestDuration: make(map[string][]time.Duration),
		enabled:         config.App.Features["metrics_enabled"],
	}
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(endpoint string, duration time.Duration) {
	if !m.enabled {
		return
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, ok := m.httpRequests[endpoint]; !ok {
		m.httpRequests[endpoint] = &atomic.Int64{}
	}
	m.httpRequests[endpoint].Add(1)
	m.requestDuration[endpoint] = append(m.requestDuration[endpoint], duration)
}

// RecordDBQuery increments the database query counter
func (m *Metrics) RecordDBQuery() {
	if !m.enabled {
		return
	}
	m.dbQueries.Add(1)
}

// RecordUserLookup increments the user lookup counter
func (m *Metrics) RecordUserLookup() {
	if !m.enabled {
		return
	}
	m.userLookups.Add(1)
}

// RecordCacheHit increments the cache hit counter
func (m *Metrics) RecordCacheHit() {
	if !m.enabled {
		return
	}
	m.cacheHits.Add(1)
}

// RecordCacheMiss increments the cache miss counter
func (m *Metrics) RecordCacheMiss() {
	if !m.enabled {
		return
	}
	m.cacheMisses.Add(1)
}

// GetStats returns current metrics as a string
func (m *Metrics) GetStats() string {
	if !m.enabled {
		return "Metrics disabled"
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := "=== Application Metrics ===\n\n"
	
	// HTTP metrics
	stats += "HTTP Requests:\n"
	for endpoint, count := range m.httpRequests {
		avgDuration := time.Duration(0)
		if durations, ok := m.requestDuration[endpoint]; ok && len(durations) > 0 {
			total := time.Duration(0)
			for _, d := range durations {
				total += d
			}
			avgDuration = total / time.Duration(len(durations))
		}
		stats += fmt.Sprintf("  %s: %d requests (avg: %v)\n", endpoint, count.Load(), avgDuration)
	}
	
	// Database metrics
	stats += fmt.Sprintf("\nDatabase:\n  Queries: %d\n", m.dbQueries.Load())
	
	// Cache metrics
	hits := m.cacheHits.Load()
	misses := m.cacheMisses.Load()
	total := hits + misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}
	stats += fmt.Sprintf("\nCache:\n  Hits: %d\n  Misses: %d\n  Hit Rate: %.1f%%\n", 
		hits, misses, hitRate)
	
	// Business metrics
	stats += fmt.Sprintf("\nBusiness:\n  User Lookups: %d\n", m.userLookups.Load())
	
	return stats
}