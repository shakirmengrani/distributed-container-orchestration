package services

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shakirmengrani/distributed_docker/cmd"
	"github.com/shakirmengrani/distributed_docker/helper"
)

type MetricsService struct {
	memTotal  *prometheus.GaugeVec
	memUsed   *prometheus.GaugeVec
	memFree   *prometheus.GaugeVec
	diskTotal *prometheus.GaugeVec
	diskUsed  *prometheus.GaugeVec
	diskFree  *prometheus.GaugeVec
	cpuTotal  *prometheus.GaugeVec
	cpuFree   *prometheus.GaugeVec
	capacity  *prometheus.GaugeVec
}

const labelPrefix = "prefix"

func NewMetricsService(server cmd.Server) *MetricsService {
	svc := &MetricsService{
		memTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "node_mem_total_bytes",
				Help: "Total memory in bytes",
			},
			[]string{labelPrefix}, // 👈 label defined here
		),

		memUsed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "node_mem_used_bytes",
				Help: "Used memory in bytes",
			},
			[]string{labelPrefix},
		),

		memFree: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "node_mem_free_bytes",
				Help: "Free memory in bytes",
			},
			[]string{labelPrefix},
		),

		diskTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "node_disk_total_bytes",
				Help: "Total disk in bytes",
			},
			[]string{labelPrefix},
		),

		diskUsed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "node_disk_used_bytes",
				Help: "Used disk in bytes",
			},
			[]string{labelPrefix},
		),

		diskFree: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "node_disk_free_bytes",
				Help: "Free disk in bytes",
			},
			[]string{labelPrefix},
		),

		cpuTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "node_cpu_total_cores",
				Help: "Total CPU cores",
			},
			[]string{labelPrefix},
		),

		cpuFree: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "node_cpu_free_cores",
				Help: "Free CPU cores",
			},
			[]string{labelPrefix},
		),

		capacity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "node_has_capacity",
				Help: "1 if node has capacity, 0 otherwise",
			},
			[]string{labelPrefix},
		),
	}

	prometheus.MustRegister(
		svc.memTotal, svc.memUsed, svc.memFree,
		svc.diskTotal, svc.diskUsed, svc.diskFree,
		svc.cpuTotal, svc.cpuFree,
		svc.capacity,
	)

	server.AddRoutes(map[string]gin.HandlerFunc{
		"/metrics": gin.WrapH(promhttp.Handler()),
	})

	return svc
}

// Collect reads current system resources and updates all gauges.
// Call this on a ticker in a goroutine.
func (m *MetricsService) Collect() error {
	prefix := os.Getenv("PREFIX")
	res, err := helper.GetSystemResources()
	if err != nil {
		return err
	}

	m.memTotal.With(prometheus.Labels{"prefix": prefix}).Set(float64(res.Mem.Total))
	m.memUsed.With(prometheus.Labels{"prefix": prefix}).Set(float64(res.Mem.Used))
	m.memFree.With(prometheus.Labels{"prefix": prefix}).Set(float64(res.Mem.Free))

	m.diskTotal.With(prometheus.Labels{"prefix": prefix}).Set(float64(res.Disk.Total))
	m.diskUsed.With(prometheus.Labels{"prefix": prefix}).Set(float64(res.Disk.Used))
	m.diskFree.With(prometheus.Labels{"prefix": prefix}).Set(float64(res.Disk.Free))

	m.cpuTotal.With(prometheus.Labels{"prefix": prefix}).Set(float64(res.CPU.Total))
	m.cpuFree.With(prometheus.Labels{"prefix": prefix}).Set(float64(res.CPU.Free))

	if helper.ComputeCapacity(*res) {
		m.capacity.With(prometheus.Labels{"prefix": prefix}).Set(1)
	} else {
		m.capacity.With(prometheus.Labels{"prefix": prefix}).Set(0)
	}

	return nil
}
