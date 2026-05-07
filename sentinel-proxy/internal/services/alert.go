package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/omar/sentinel-proxy/internal/metrics"
)

var lastAlertTime time.Time
var mu sync.Mutex

func CheckAlerts() {
	mu.Lock()
	defer mu.Unlock()
	topIP, count := metrics.GetTopIP()

	stats := metrics.GetStats()

	blockRate := float64(stats.Blocked) / float64(stats.Total) * 100

	// cooldown (don’t spam alerts)
	if time.Since(lastAlertTime) < 10*time.Second {
		return
	}

	if blockRate > 30 {
		fmt.Printf("ALERT: High block rate: %.1f%%\n", blockRate)
		lastAlertTime = time.Now()
	}

	if count > 20 {
		fmt.Printf("ALERT: Suspicious IP: %s requests: %d\n", topIP, count)
	}

	if stats.Total > 20 && blockRate > 30 {
		return // avoid noise
	}
}
