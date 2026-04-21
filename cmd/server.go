package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shakirmengrani/distributed_docker/helper"
)

type Server struct {
	gin *gin.Engine
}

func NewServer() Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
	})
	r.GET("/info", func(ctx *gin.Context) {
		sysRes, _ := helper.GetSystemResources()
		ctx.JSON(http.StatusOK, gin.H{
			"prefix":      os.Getenv("PREFIX"),
			"ip":          os.Getenv("IP"),
			"resources":   sysRes,
			"is_capacity": helper.ComputeCapacity(*sysRes),
		})
	})
	return Server{gin: r}
}

func (server *Server) Listen(addr string) error {
	log.Println(fmt.Sprintf("Listen on %s", addr))
	return server.gin.Run(addr)
}

func (server *Server) AddRoutes(routes map[string]gin.HandlerFunc) {
	for k, i := range routes {
		server.gin.Any(k, i)
	}
}
