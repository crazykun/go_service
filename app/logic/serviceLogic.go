package logic

import (
	"context"
	"errors"
	"fmt"
	"go_service/app/config"
	"go_service/app/model"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ServiceLogic
type ServiceLogic struct {
	c  *gin.Context
	db *gorm.DB
}

// NewServiceLogic
func NewServiceLogic() *ServiceLogic {
	return &ServiceLogic{c: config.ServerContext, db: config.GetDefaultDb()}
}

func (s ServiceLogic) Add(ctx context.Context, info model.ServiceModel) (int64, error) {
	var in model.ServiceModel
	s.db.First(&in, "port", info.Port)
	if in.Port > 0 && in.Port == info.Port { //去重
		return 0, errors.New("port is exist")
	}
	s.db.Save(&info) //要使用指针
	return info.Id, nil
}

func (s ServiceLogic) GetByName(ctx context.Context, name string) model.ServiceModel {
	var info model.ServiceModel
	s.db.First(&info, "name", name)
	return info
}

func (s ServiceLogic) FindAll(ctx context.Context) ([]model.ServiceStatusModel, error) {
	var infos []model.ServiceModel
	var infoNew []model.ServiceStatusModel
	s.db.Find(&infos)
	if len(infos) == 0 {
		return infoNew, nil
	}

	out, err := exec.Command("netstat", "-nptl").Output()
	if err != nil {
		fmt.Println("Error:", err)
		return infoNew, err
	}
	outList := strings.Split(string(out), "\n")
	outList = outList[2:]
	portList := make(map[string]map[string]interface{}, 0)
	for _, line := range outList {
		if strings.Contains(line, "LISTEN") {
			parts := strings.Fields(line)
			address := parts[3]
			port := strings.Split(address, ":")[1]
			var pid, process_name string
			process := strings.Split(parts[6], "/")
			if len(process) != 2 {
				pid = ""
				process_name = ""
			} else {
				pid = process[0]
				process_name = process[1]
			}
			portList[port] = map[string]interface{}{
				"pid":     pid,
				"process": process_name,
			}
		}
	}

	for _, v := range infos {
		var status int
		var pid string
		var process string
		// 根据port查询服务是否启动
		port := strconv.Itoa(int(v.Port))
		var tmp = portList[port]
		if tmp != nil {
			fmt.Println("tmp:", tmp)
			status = 1
			pid = tmp["pid"].(string)
			process = tmp["process"].(string)
		}
		infoNew = append(infoNew, model.ServiceStatusModel{ServiceModel: v, Status: status, Pid: pid, Process: process})
	}
	return infoNew, nil
}

func (s ServiceLogic) DeleteById(ctx context.Context, id int64) bool {
	s.db.Delete(&model.ServiceModel{}, id)
	return true
}

func (s ServiceLogic) GetById(ctx context.Context, id int64) model.ServiceModel {
	var info model.ServiceModel
	s.db.First(&info, "id", id)
	return info
}

func (s ServiceLogic) UpdateById(ctx context.Context, info model.ServiceModel) bool {
	s.db.Model(&model.ServiceModel{}).Where("id", info.Id).Updates(info)
	return true
}
