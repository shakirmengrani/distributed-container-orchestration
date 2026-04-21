package cmd

import (
	"os"

	"github.com/shakirmengrani/distributed_docker/helper"
)

type Config struct {
	Id           string `yaml:"id"`
	Prefix       string `yaml:"prefix"`
	Ip           string `yaml:"ip"`
	Address      string `yaml:"address"`
	ControlPlane string `yaml:"control_plane"`
	Etcd         string `yaml:"etcd"`
	Docker       string `yaml:"docker"`
}

func NewConfig() (Config, error) {
	cfg, err := helper.ReadYaml[Config]("./config.yml")
	os.Setenv("PREFIX", cfg.Prefix)
	os.Setenv("IP", cfg.Address)
	if cfg.ControlPlane != "" {
		os.Setenv("CONTROL_PLANE", cfg.ControlPlane)
	}
	os.Setenv("ETCD_ENDPOINTS", cfg.Etcd)
	return cfg, err
}
