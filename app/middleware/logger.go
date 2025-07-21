package middleware

// logger.go zip日志中间件
import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 执行请求
		// 记录请求体
		request_body := ""
		if c.Request.Body != nil {
			body, err := c.GetRawData()
			if err == nil {
				request_body = string(body)
			} else {
				request_body = "body_err:" + err.Error()
			}
		}
		// 记录日志
		gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			log := fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s  %s | %s\n",
				param.TimeStamp.Format("2006-01-02 15:04:05"),
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
				request_body,
			)
			return log
		})
	}
}
