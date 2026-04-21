package helper

import (
	"os"

	"github.com/goccy/go-yaml"
)

func ReadYaml[T any](file string) (T, error) {
	var cfg T
	data, err := os.ReadFile(file)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}
