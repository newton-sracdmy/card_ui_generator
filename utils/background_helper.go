package utils

import (
	"context"
	"log"
	"time"
)

func RunEvery(ctx context.Context, serviceName string, interval time.Duration, job func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[STOPPED] %s", serviceName)
			return
		case <-ticker.C:
			job()
		}
	}
}
