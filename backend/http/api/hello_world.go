package api

import (
	"q-dev/http/common"

	"github.com/gin-gonic/gin"
)

type HelloWorldAPI struct{}

// List @Summary 列表
// @Tags    hello-world
// @Produce json
// @Success 200 {object} common.Response
// @Router  /v1/hello-world [get]
func (h *HelloWorldAPI) List(c *gin.Context) {
	common.OK(c, gin.H{"message": "hello world"})
}

// Get @Summary 详情
// @Tags    hello-world
// @Produce json
// @Param   id path string true "ID"
// @Success 200 {object} common.Response
// @Router  /v1/hello-world/{id} [get]
func (h *HelloWorldAPI) Get(c *gin.Context) {
	id := c.Param("id")
	common.OK(c, gin.H{"id": id, "message": "hello world"})
}
