package common

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义业务错误码
const (
	ErrCodeSuccess         = 0
	ErrCodeInvalidParam    = 1001
	ErrCodeServiceNotFound = 1002
	ErrCodePortInUse       = 1003
	ErrCodeServiceRunning  = 1004
	ErrCodeServiceStopped  = 1005
	ErrCodeCommandFailed   = 1006
	ErrCodeDatabaseError   = 1007
	ErrCodePermissionDenied = 1008
)

// BusinessError 业务错误
type BusinessError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BusinessError) Error() string {
	return e.Message
}

// NewBusinessError 创建业务错误
func NewBusinessError(code int, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// 预定义的业务错误
var (
	ErrInvalidParam    = NewBusinessError(ErrCodeInvalidParam, "参数错误")
	ErrServiceNotFound = NewBusinessError(ErrCodeServiceNotFound, "服务不存在")
	ErrPortInUse       = NewBusinessError(ErrCodePortInUse, "端口已被占用")
	ErrServiceRunning  = NewBusinessError(ErrCodeServiceRunning, "服务正在运行")
	ErrServiceStopped  = NewBusinessError(ErrCodeServiceStopped, "服务已停止")
	ErrCommandFailed   = NewBusinessError(ErrCodeCommandFailed, "命令执行失败")
	ErrDatabaseError   = NewBusinessError(ErrCodeDatabaseError, "数据库操作失败")
	ErrPermissionDenied = NewBusinessError(ErrCodePermissionDenied, "权限不足")
)

// ErrorResponse 统一错误响应处理
func ErrorResponse(c *gin.Context, err error) {
	if bizErr, ok := err.(*BusinessError); ok {
		c.JSON(http.StatusOK, gin.H{
			"code":    bizErr.Code,
			"message": bizErr.Message,
			"data":    nil,
		})
		return
	}

	// 未知错误
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    http.StatusInternalServerError,
		"message": "内部服务器错误",
		"data":    nil,
	})
}

// SuccessResponse 统一成功响应处理
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    ErrCodeSuccess,
		"message": "success",
		"data":    data,
	})
}

// WrapError 包装错误信息
func WrapError(code int, message string, err error) *BusinessError {
	if err != nil {
		message = fmt.Sprintf("%s: %v", message, err)
	}
	return NewBusinessError(code, message)
}