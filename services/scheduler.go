package services

import (
	"log"
	"os"
	"time"

	"github.com/shakirmengrani/distributed_docker/cmd"
	"github.com/shakirmengrani/distributed_docker/helper"
)

type Scheduler struct {
	etcd    cmd.Etcd
	metrics MetricsService
}

func NewScheduler(etcd cmd.Etcd, metrics MetricsService) Scheduler {
	return Scheduler{
		etcd:    etcd,
		metrics: metrics,
	}
}

func (s *Scheduler) CollectMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := s.metrics.Collect(); err != nil {
			log.Printf("[metrics] collection failed: %v", err)
		}
	}

}

func (s *Scheduler) Heartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := helper.SendHeartbeat(os.Getenv("CONTROL_PLANE"), map[string]string{
			"id":      os.Getenv("PREFIX"),
			"address": os.Getenv("IP"),
		}); err != nil {
			log.Printf("[heartbeat] failed: %v", err)
		}
	}
}
