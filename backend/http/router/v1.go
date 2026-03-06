package router

import (
	"q-dev/http/api"

	"github.com/gin-gonic/gin"
)

func RegisterV1(api *gin.RouterGroup) {
	v1 := api.Group("/v1")

	registerHelloWorld(v1)
}

func registerHelloWorld(rg *gin.RouterGroup) {
	h := &api.HelloWorldAPI{}
	g := rg.Group("/hello-world")
	g.GET("", h.List)
	g.GET("/:id", h.Get)
}
