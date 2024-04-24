package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// 捕获gin 500错误中间件
func ExceptErr() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 发生了 panic，返回 500 错误
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
				// 或者记录错误日志
				fmt.Println("Recovered from panic:", r)
				debug.PrintStack()
			}
		}()
		c.Next()
	}
}
