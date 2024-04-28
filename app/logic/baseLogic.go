package logic

import (
	"go_service/app/global"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Logic struct {
	c  *gin.Context
	db *gorm.DB
}

// NewLogic
func NewLogic() *Logic {
	return &Logic{c: global.ServerContext, db: global.GetDefaultDb()}
}
