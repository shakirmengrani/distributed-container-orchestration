package main

import (
	"strings"

	"github.com/shakirmengrani/distributed_docker/cmd"
	"github.com/shakirmengrani/distributed_docker/services"
)

func main() {
	cfg, err := cmd.NewConfig()
	if err != nil {
		panic(err)
	}
	etcd, err := cmd.NewEtcd()
	if err != nil {
		panic(err)
	}
	docker, err := cmd.NewDocker()
	if err != nil {
		panic(err)
	}
	http := cmd.NewServer()
	metrics := services.NewMetricsService(http)
	nodeScorer := services.NewNodeScorerService(*etcd)
	scheduler := services.NewScheduler(*etcd, *metrics)
	services.NewMemberService(http, *etcd)
	services.NewDockerService(http, *etcd, docker, &nodeScorer)
	go scheduler.Heartbeat()
	go scheduler.CollectMetrics()
	port := strings.Split(cfg.Address, ":")[1]
	err = http.Listen(":" + port)
	if err != nil {
		panic(err)
	}
}
