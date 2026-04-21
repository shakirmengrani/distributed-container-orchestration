package services

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shakirmengrani/distributed_docker/cmd"
	"github.com/shakirmengrani/distributed_docker/helper"
	"github.com/shakirmengrani/distributed_docker/types"
)

type docker_service struct {
	etcd   cmd.Etcd
	docker cmd.Docker
	scorer *NodeScorer
}

type Container struct {
	Name   string
	Image  string
	Lables map[string]string
}

func NewDockerService(server cmd.Server, etcd cmd.Etcd, docker cmd.Docker, scorer *NodeScorer) {
	svc := docker_service{etcd: etcd, docker: docker, scorer: scorer}
	r := make(map[string]gin.HandlerFunc)
	r["/container"] = svc.CreateContainer
	r["/container/list"] = svc.List
	r["/container/remove"] = svc.RemoveContainer
	r["/container/connect"] = svc.ConnectDomain
	server.AddRoutes(r)
}

func (svc *docker_service) CreateContainer(c *gin.Context) {
	switch c.Request.Method {
	case "POST":
		sysRes, err := helper.GetSystemResources()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		body, err := helper.ReadBodyStruct[types.ContainerConfig](c)
		if err != nil && (body.Name == "") {
			c.JSON(http.StatusBadRequest, errors.New("Bad request"))
		}
		isCapacity := helper.ComputeCapacity(*sysRes)
		if !isCapacity {
			node, body, err := svc.forwardContainer(body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			// add container info with node info to etcd for future reference
			b, err := helper.ReadBodyStruct[types.ContainerConfig](c)
			if err != nil {
				c.JSON(http.StatusBadRequest, errors.New("Bad request"))
				return
			}
			err = svc.etcd.AddContainer(node, b.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			c.JSON(http.StatusAccepted, gin.H{
				"message": "Container scheduled on another node",
				"node":    node,
				"details": string(body),
			})
			return
		}
		_, _, err = svc.docker.Create(body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		err = svc.etcd.AddTraefik(&body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		err = svc.etcd.AddContainer(os.Getenv("PREFIX"), body.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"message": "Container created successfully",
			"node":    os.Getenv("PREFIX"),
			"details": body,
		})
		return
	case "GET":
		if c.Query("name") == "" {
			c.JSON(http.StatusBadRequest, errors.New("Bad request"))
			return
		}
		container, err := svc.docker.Find(c.Query("name"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if len(container.Items) <= 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Container not found"})
			return
		}
		details, err := svc.docker.Details(container.Items[0].ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		labels, err := svc.etcd.TraefikList(fmt.Sprintf("//%s", c.Query("name")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"Container": details.Container,
			"Labels":    labels,
		})
		return
	default:
		c.AbortWithError(http.StatusForbidden, errors.New("Access forbidden"))
		return
	}
}

func (svc *docker_service) RemoveContainer(c *gin.Context) {
	if c.Request.Method == "POST" {
		body, err := helper.ReadBodyStruct[types.ContainerConfig](c)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.New("Bad request"))
		}
		node, _, err := svc.forwardRemoveContainer(body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		err = svc.etcd.RemoveContainer(body.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if node != os.Getenv("PREFIX") {
			c.JSON(http.StatusAccepted, gin.H{
				"message": "Container removal forwarded to another node",
				"node":    node,
				"details": body,
			})
			return
		}
		err = svc.docker.Remove(body.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		err = svc.etcd.RemoveTraefik(&body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusAccepted, body)
		return
	}
	c.AbortWithError(http.StatusForbidden, errors.New("Access forbidden"))
}

func (svc *docker_service) ConnectDomain(c *gin.Context) {
	if c.Request.Method == "POST" {
		body, err := helper.ReadBodyStruct[types.ContainerConfig](c)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.New("Bad request"))
		}
		node, _, err := svc.forwardConnectDomain(body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if node != os.Getenv("PREFIX") {
			c.JSON(http.StatusAccepted, gin.H{
				"message": "Domain connection forwarded to another node",
				"node":    node,
				"details": body,
			})
			return
		}
		err = svc.etcd.AddTraefik(&body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusAccepted, body)
		return
	}
	c.AbortWithError(http.StatusForbidden, errors.New("Access forbidden"))
}

func (svc *docker_service) List(c *gin.Context) {
	var conts []Container
	containers, err := svc.docker.Find("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	for _, item := range containers.Items {
		labelList, err := svc.etcd.TraefikList(fmt.Sprintf("/%s", item.Names[0]))
		if err != nil {
			continue
		}
		if len(labelList) > 0 {
			c := Container{
				Name:   item.Names[0],
				Image:  item.Image,
				Lables: labelList,
			}
			conts = append(conts, c)
		}
	}
	c.JSON(http.StatusAccepted, conts)
}

func (svc *docker_service) forwardContainer(cfg types.ContainerConfig) (string, []byte, error) {
	nodes, err := svc.scorer.Rank()
	if err != nil {
		return "", nil, err
	}
	if len(nodes) == 0 {
		return "", nil, err
	}
	for _, node := range nodes {
		nodeResp, err := helper.Forward(node, cfg, false, false)
		if err == nil {
			return node, nodeResp, nil
		}
	}
	return "", nil, errors.New("No node available")
}

func (svc *docker_service) forwardRemoveContainer(cfg types.ContainerConfig) (string, []byte, error) {
	node, err := svc.etcd.ContainerNode(cfg.Name)
	if err != nil {
		return "", nil, err
	}
	if node == "" {
		return "", nil, errors.New("Container not found in any node")
	}
	if node != os.Getenv("PREFIX") {
		nodeResp, err := helper.Forward(node, cfg, true, false)
		if err != nil {
			return "", nil, err
		}
		return node, nodeResp, nil
	}
	return os.Getenv("PREFIX"), nil, nil
}

func (svc *docker_service) forwardConnectDomain(cfg types.ContainerConfig) (string, []byte, error) {
	node, err := svc.etcd.ContainerNode(cfg.Name)
	if err != nil {
		return "", nil, err
	}
	if node == "" {
		return "", nil, errors.New("Container not found in any node")
	}
	if node != os.Getenv("PREFIX") {
		nodeResp, err := helper.Forward(node, cfg, false, true)
		if err != nil {
			return "", nil, err
		}
		return node, nodeResp, nil
	}
	return os.Getenv("PREFIX"), nil, nil
}
