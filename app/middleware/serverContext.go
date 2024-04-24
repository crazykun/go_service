package middleware

import (
	"go_service/app/config"

	"github.com/gin-gonic/gin"
)

func ServerContextHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		config.ServerContext = c
		c.Next()
	}
}
