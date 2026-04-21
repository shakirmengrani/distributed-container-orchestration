package services

import (
	"github.com/shakirmengrani/distributed_docker/cmd"
	"github.com/shakirmengrani/distributed_docker/helper"
)

type NodeScorer struct {
	etcd cmd.Etcd
}

func NewNodeScorerService(etcd cmd.Etcd) NodeScorer {
	return NodeScorer{etcd: etcd}
}

func (n *NodeScorer) Rank() ([]string, error) {
	nodes, err := n.etcd.NodeList()
	if err != nil {
		return nil, err
	}
	var nodeList []string
	for _, kv := range nodes.Kvs {
		info, err := helper.NodeInfo(string(kv.Value))
		if err != nil {
			continue
		}
		hasCapacity, _ := info["is_capacity"].(bool)
		// mem := extractMemFree(info)
		if hasCapacity {
			nodeList = append(nodeList, string(kv.Value))
		}
	}
	return nodeList, nil
}

func extractMemFree(info map[string]any) uint64 {
	resources, ok := info["resources"].(map[string]any)
	if !ok {
		return 0
	}
	mem, ok := resources["mem"].(map[string]any)
	if !ok {
		return 0
	}
	free, _ := mem["free"].(float64)
	return uint64(free)
}
