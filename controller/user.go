package controller

import (
	"github.com/distribyted/distribyted/server"
	"github.com/gin-gonic/gin"
)

func UserLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	data, err := server.UserLogin(username, password)
	if err != nil {
		ErrorJson(c, data, err.Error())
	}
	SuccessJson(c, data, "")
}

func UserInfo(c *gin.Context) {
	token := c.Query("token")
	data, err := server.UserInfo(token)
	if err != nil {
		ErrorJson(c, data, err.Error())
	}
	SuccessJson(c, data, "")
}
