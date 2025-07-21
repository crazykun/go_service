package logic

import (
	"context"
	"errors"
	"fmt"
	"go_service/app/model"
	"go_service/pkg/utils"
	"strconv"
)

func (l Logic) Add(ctx context.Context, info model.ServiceModel) (int64, error) {
	// 验证必要字段
	if info.Name == "" {
		return 0, errors.New("服务名称不能为空")
	}
	if info.Port <= 0 {
		return 0, errors.New("端口号无效")
	}
	if info.CmdStart == "" {
		return 0, errors.New("启动命令不能为空")
	}

	// 检查端口是否已存在
	var existing model.ServiceModel
	l.db.Where("port = ?", info.Port).First(&existing)
	if existing.Id > 0 {
		return 0, errors.New("端口已被其他服务占用")
	}

	// 检查服务名是否已存在
	l.db.Where("name = ?", info.Name).First(&existing)
	if existing.Id > 0 {
		return 0, errors.New("服务名称已存在")
	}

	// 保存服务信息
	result := l.db.Create(&info)
	if result.Error != nil {
		return 0, fmt.Errorf("保存服务失败: %v", result.Error)
	}
	
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
	// 验证必要字段
	if info.Id <= 0 {
		return false
	}
	if info.Name == "" || info.Port <= 0 || info.CmdStart == "" {
		return false
	}

	// 检查端口是否被其他服务占用
	var existing model.ServiceModel
	l.db.Where("port = ? AND id != ?", info.Port, info.Id).First(&existing)
	if existing.Id > 0 {
		return false
	}

	// 检查服务名是否被其他服务占用
	l.db.Where("name = ? AND id != ?", info.Name, info.Id).First(&existing)
	if existing.Id > 0 {
		return false
	}

	result := l.db.Model(&model.ServiceModel{}).Where("id = ?", info.Id).Updates(info)
	return result.Error == nil && result.RowsAffected > 0
}
