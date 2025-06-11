package cache

import (
	"log"
	"sync/atomic"
	"time"
)

var (
	userCacheHits   uint64
	userCacheMisses uint64
	catCacheHits    uint64
	catCacheMisses  uint64
)

// recordUserCacheHit increments the user cache hit counter
func recordUserCacheHit() {
	atomic.AddUint64(&userCacheHits, 1)
}

// recordUserCacheMiss increments the user cache miss counter
func recordUserCacheMiss() {
	atomic.AddUint64(&userCacheMisses, 1)
}

// recordCatCacheHit increments the category cache hit counter
func recordCatCacheHit() {
	atomic.AddUint64(&catCacheHits, 1)
}

// recordCatCacheMiss increments the category cache miss counter
func recordCatCacheMiss() {
	atomic.AddUint64(&catCacheMisses, 1)
}

// StartCacheMetricsReporter starts reporting cache metrics periodically
func StartCacheMetricsReporter() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			// Get current values atomically
			userHits := atomic.LoadUint64(&userCacheHits)
			userMisses := atomic.LoadUint64(&userCacheMisses)
			catHits := atomic.LoadUint64(&catCacheHits)
			catMisses := atomic.LoadUint64(&catCacheMisses)

			// Calculate hit rates
			userTotal := float64(userHits + userMisses)
			catTotal := float64(catHits + catMisses)

			var userHitRate, catHitRate float64
			if userTotal > 0 {
				userHitRate = float64(userHits) / userTotal * 100
			}
			if catTotal > 0 {
				catHitRate = float64(catHits) / catTotal * 100
			}

			// Log metrics
			log.Printf("Cache Metrics - User: %.2f%% hit rate (%d hits, %d misses), Category: %.2f%% hit rate (%d hits, %d misses)",
				userHitRate, userHits, userMisses,
				catHitRate, catHits, catMisses)
		}
	}()
} 