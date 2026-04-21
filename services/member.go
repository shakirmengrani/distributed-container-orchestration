package services

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	cmd "github.com/shakirmengrani/distributed_docker/cmd"
	"github.com/shakirmengrani/distributed_docker/helper"
)

type member_service struct {
	etcd cmd.Etcd
}

func NewMemberService(server cmd.Server, etcd cmd.Etcd) {
	svc := member_service{etcd: etcd}
	r := make(map[string]gin.HandlerFunc)
	r["/member"] = svc.addMember
	r["/member/list"] = svc.list
	server.AddRoutes(r)
}

func (svc *member_service) addMember(c *gin.Context) {
	if c.Request.Method == "POST" {
		body, err := helper.ReadBodyMap(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.New("Bad request"))
		}
		svc.etcd.AddMember(body["id"].(string), body["address"].(string))
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Member added",
			"id":      body["id"].(string),
			"address": body["address"].(string),
		})
		return
	}
	c.AbortWithError(http.StatusForbidden, errors.New("Access forbidden"))
}

func (svc *member_service) list(c *gin.Context) {
	list, err := svc.etcd.NodeList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	kvs := make(map[string]string)
	for _, kv := range list.Kvs {
		kvs[string(kv.Key)] = string(kv.Value)
	}
	c.JSON(http.StatusOK, kvs)
}
