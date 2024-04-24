package logic

import (
	"context"
	"errors"
	"go_service/app/model"
	"go_service/pkg/utils"
	"strconv"
)

func (l Logic) Add(ctx context.Context, info model.ServiceModel) (int64, error) {
	var in model.ServiceModel
	l.db.First(&in, "port", info.Port)
	if in.Port > 0 && in.Port == info.Port { //去重
		return 0, errors.New("port is exist")
	}
	l.db.Save(&info) //要使用指针
	return info.Id, nil
}

func (l Logic) GetByName(ctx context.Context, name string) model.ServiceModel {
	var info model.ServiceModel
	l.db.First(&info, "name", name)
	return info
}

func (l Logic) FindAll(ctx context.Context) ([]model.ServiceStatusModel, error) {
	var infos []model.ServiceModel
	var infoNew []model.ServiceStatusModel
	l.db.Find(&infos)
	if len(infos) == 0 {
		return infoNew, nil
	}

	portList, err := utils.GetPortList()
	if err != nil {
		return infoNew, err
	}

	for _, v := range infos {
		var status int
		var pid string
		var process string
		// 根据port查询服务是否启动
		port := strconv.Itoa(int(v.Port))
		var tmp = portList[port]
		if tmp != nil {
			status = 1
			pid = tmp["pid"].(string)
			process = tmp["process"].(string)
		}
		infoNew = append(infoNew, model.ServiceStatusModel{ServiceModel: v, Status: status, Pid: pid, Process: process})
	}
	return infoNew, nil
}

func (l Logic) DeleteById(ctx context.Context, id int64) bool {
	l.db.Delete(&model.ServiceModel{}, id)
	return true
}

func (l Logic) GetById(ctx context.Context, id int64) model.ServiceModel {
	var info model.ServiceModel
	l.db.First(&info, "id", id)
	return info
}

func (l Logic) UpdateById(ctx context.Context, info model.ServiceModel) bool {
	l.db.Model(&model.ServiceModel{}).Where("id", info.Id).Updates(info)
	return true
}
