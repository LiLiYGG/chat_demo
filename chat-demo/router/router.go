package router

import (
	"chat-demo/api"
	"chat-demo/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()
	r.Use()
	v1 := r.Group("/")
	{
		// 测试连接
		v1.GET("ping", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, "success")
		})
		// 用户注册
		v1.POST("user/register", api.UserRegister)
		v1.GET("ws", service.Handler)

	}
	return r
}
