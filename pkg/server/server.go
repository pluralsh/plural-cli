package server

import (
	"github.com/gin-gonic/gin"
)

func Run() error {
	r := gin.Default()
	v1 := r.Group("/v1")
	{
		v1.POST("/setup", serverFunc(setupCli))
	}
	
	return r.Run(":8080")
}