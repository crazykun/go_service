package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// Error 错误响应
func Error(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 1,
		Msg:  msg,
		Data: gin.H{},
	})
}

// ErrorWithCode 带错误码的错误响应
func ErrorWithCode(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: gin.H{},
	})
}

// Paginate 分页响应
func Paginate(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: gin.H{
			"list":      data,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// HandleBusinessError 处理业务错误
func HandleBusinessError(c *gin.Context, err error) {
	if bizErr, ok := err.(*BusinessError); ok {
		ErrorWithCode(c, bizErr.Code, bizErr.Message)
		return
	}
	Error(c, err.Error())
}