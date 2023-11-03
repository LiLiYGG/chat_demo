package api

import (
	"chat-demo/service"
	"net/http"

	"github.com/gin-gonic/gin"
	logging "github.com/sirupsen/logrus"
)

func UserRegister(ctx *gin.Context) {
	// 定义一个用户注册服务
	var UserRegisterService service.UserRegisterService
	if err := ctx.ShouldBind(&UserRegisterService); err == nil {
		res := UserRegisterService.Register()
		ctx.JSON(http.StatusOK, res)
	} else {
		ctx.JSON(400, ErrorResponse(err))
		logging.Info(err)
	}
}
