package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shakirmengrani/distributed_docker/types"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Etcd struct {
	client *clientv3.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func NewEtcd() (*Etcd, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{os.Getenv("ETCD_ENDPOINTS")},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return &Etcd{
			ctx:    ctx,
			cancel: cancelFunc,
		}, err
	}
	return &Etcd{client: cli, ctx: ctx, cancel: cancelFunc}, nil
}

func (etcd *Etcd) AddMember(nodeId string, ip string) error {
	leaseResp, err := etcd.client.Lease.Grant(etcd.ctx, 30)
	if err != nil {
		return err
	}
	_, err = etcd.client.KV.Put(etcd.ctx, fmt.Sprintf("%s/nodes/%s", os.Getenv("PREFIX"), nodeId), ip, clientv3.WithLease(leaseResp.ID))
	return err
}

func (etcd *Etcd) RemoveMember(nodeId string) error {
	_, err := etcd.client.KV.Delete(etcd.ctx, fmt.Sprintf("%s/nodes/%s", os.Getenv("PREFIX"), nodeId))
	return err
}

func (etcd *Etcd) AddTraefik(containerConfig *types.ContainerConfig) error {
	var err error
	labels := etcd.makeTraefikLabels(containerConfig)
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	for _, key := range keys {
		_, err = etcd.client.KV.Put(etcd.ctx, key, labels[key])
	}
	return err
}

func (etcd *Etcd) RemoveTraefik(containerConfig *types.ContainerConfig) error {
	var err error
	labels := etcd.makeTraefikLabels(containerConfig)
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	for _, key := range keys {
		_, err = etcd.client.KV.Delete(etcd.ctx, key)
	}
	return err
}

func (etcd *Etcd) AddContainer(nodeAddress string, containerId string) error {
	_, err := etcd.client.KV.Put(etcd.ctx, fmt.Sprintf("%s/containers/%s", os.Getenv("PREFIX"), containerId), nodeAddress)
	return err
}

func (etcd *Etcd) RemoveContainer(containerId string) error {
	_, err := etcd.client.KV.Delete(etcd.ctx, fmt.Sprintf("%s/containers/%s", os.Getenv("PREFIX"), containerId))
	return err
}

func (etcd *Etcd) Containerlist() (map[string]string, error) {
	resp, err := etcd.client.KV.Get(etcd.ctx, fmt.Sprintf("%s/containers", os.Getenv("PREFIX")), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, item := range resp.Kvs {
		result[string(item.Key)] = string(item.Value)
	}
	return result, nil
}

func (etcd *Etcd) ContainerNode(containerId string) (string, error) {
	resp, err := etcd.client.KV.Get(etcd.ctx, fmt.Sprintf("%s/containers/%s", os.Getenv("PREFIX"), containerId))
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) > 0 {
		return string(resp.Kvs[0].Value), nil
	}
	return "", nil
}

func (etcd *Etcd) NodeList() (*clientv3.GetResponse, error) {
	return etcd.client.KV.Get(etcd.ctx, fmt.Sprintf("%s/nodes", os.Getenv("PREFIX")), clientv3.WithPrefix())
}

func (etcd *Etcd) TraefikList(prefix string) (map[string]string, error) {
	suffix := strings.TrimPrefix(prefix, "//")
	resp, err := etcd.client.KV.Get(etcd.ctx, fmt.Sprintf("%s/traefik/http/routers/%s", os.Getenv("PREFIX"), suffix), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, item := range resp.Kvs {
		result[string(item.Key)] = string(item.Value)
	}
	respSVC, err := etcd.client.KV.Get(etcd.ctx, fmt.Sprintf("%s/traefik/http/services/%s/loadBalancer/servers/0/url", os.Getenv("PREFIX"), fmt.Sprintf("service-%s", suffix)))
	if err != nil {
		return nil, err
	}
	for _, item := range respSVC.Kvs {
		result[string(item.Key)] = string(item.Value)
	}
	return result, nil
}

func (etcd *Etcd) makeTraefikLabels(containerConfig *types.ContainerConfig) map[string]string {
	prefix := fmt.Sprintf("%s/traefik/http/routers/%s", os.Getenv("PREFIX"), containerConfig.Name)
	labels := make(map[string]string)
	domains := ""
	for i, k := range containerConfig.Domain {
		if i == 0 {
			domains += fmt.Sprintf("Host(`%s`)", k)
		} else {
			domains += fmt.Sprintf(" || Host(`%s`)", k)
		}
	}
	labels[fmt.Sprintf("%s/rule", prefix)] = domains
	labels[fmt.Sprintf("%s/entrypoints", prefix)] = "websecure"
	labels[fmt.Sprintf("%s/tls/certresolver", prefix)] = "httpresolver"
	labels[fmt.Sprintf("%s/service", prefix)] = fmt.Sprintf("service-%s", containerConfig.Name)
	labels[fmt.Sprintf("%s/traefik/http/services/%s/loadBalancer/servers/0/url", os.Getenv("PREFIX"), fmt.Sprintf("service-%s", containerConfig.Name))] = fmt.Sprintf("http://%s:%d", containerConfig.Name, containerConfig.Port)
	return labels
}
