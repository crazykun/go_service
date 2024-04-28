package middleware

import (
	"go_service/app/global"

	"github.com/gin-gonic/gin"
)

func ServerContextHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		global.ServerContext = c
		c.Next()
	}
}
