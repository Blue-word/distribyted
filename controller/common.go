package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func SuccessJson(c *gin.Context, data interface{}, msg string) {
	rsp := gin.H{
		"code":    20000,
		"message": msg,
		"data":    data,
	}
	c.JSON(http.StatusOK, rsp)
}

func ErrorJson(c *gin.Context, data interface{}, msg string) {
	rsp := gin.H{
		"code":    40000,
		"message": msg,
		"data":    data,
	}
	c.JSON(http.StatusOK, rsp)
}
