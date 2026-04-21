package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakirmengrani/distributed_docker/types"
)

func ReadBodyMap(c *gin.Context) (map[string]any, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	err = json.Unmarshal(body, &data)
	return data, err
}

func ReadBodyStruct[T any](c *gin.Context) (T, error) {
	var data T
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(body, &data)
	return data, err
}

func ToJsonString(data any) (string, error) {
	jsonData, err := json.Marshal(data)
	return string(jsonData), err
}

func NodeInfo(address string) (map[string]any, error) {
	client := &http.Client{}
	resp, err := client.Get(fmt.Sprintf("http://%s/info", address))
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	err = json.Unmarshal(body, &data)
	return data, err
}

func SendHeartbeat(address string, cfg map[string]string) error {
	client := &http.Client{}
	_, err := client.Post(
		fmt.Sprintf("http://%s/member", address),
		"application/json",
		bytes.NewReader([]byte(fmt.Sprintf(`{"id": "%s", "address": "%s"}`, cfg["id"], cfg["address"]))),
	)
	return err
}

func Forward(address string, cfg types.ContainerConfig, remove bool, connect bool) ([]byte, error) {
	client := &http.Client{}
	payload, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/container", address)
	if remove {
		url = fmt.Sprintf("http://%s/container/remove", address)
	}
	if connect {
		url = fmt.Sprintf("http://%s/container/connect", address)
	}
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("node returned %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
