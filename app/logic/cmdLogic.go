package logic

import (
	"context"
	"errors"
	"fmt"
	"go_service/pkg/utils"
	"os/exec"
	"strconv"
	"strings"
)

func (l Logic) Start(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("service not found")
	}
	port := strconv.Itoa(int(info.Port))
	isUse, _ := utils.IsPortInUse(port)
	if isUse {
		// portList, _ := utils.GetPortList()
		// fmt.Println(portList[port])
		return "", errors.New("port is in use")
	}
	// 进入目录并启动服务 info.CmdStart =  php easyswoole server start -d
	args := strings.Split(info.CmdStart, " ")
	cmd := exec.Command(args[0], args[1:]...)
	// 设置工作目录
	cmd.Dir = info.Dir
	// 判断如果命令是php开头, 先查询php的路径
	if strings.HasPrefix(args[0], "php") {
		phpPath, err := exec.LookPath(args[0])
		if err != nil {
			return "", err
		}
		if phpPath == "" {
			return "", errors.New(args[0] + " not found ")
		}
		cmd.Path = phpPath
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Command output:", string(out))
		return "", err
	}
	return string(out), nil
}

func (l Logic) Stop(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("service not found")
	}
	port := strconv.Itoa(int(info.Port))
	isUse, _ := utils.IsPortInUse(port)
	if isUse {
		return "", errors.New("port is in use")
	}
	// 进入目录并启动服务
	cmd := exec.Command(info.CmdStart)
	cmd.Dir = info.Dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (l Logic) Restart(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("service not found")
	}
	port := strconv.Itoa(int(info.Port))
	isUse, _ := utils.IsPortInUse(port)
	if isUse {
		return "", errors.New("port is in use")
	}
	// 进入目录并启动服务
	cmd := exec.Command(info.CmdStart)
	cmd.Dir = info.Dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (l Logic) ForcedRestart(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("service not found")
	}
	port := strconv.Itoa(int(info.Port))
	isUse, _ := utils.IsPortInUse(port)
	if isUse {
		return "", errors.New("port is in use")
	}
	// 进入目录并启动服务
	cmd := exec.Command(info.CmdStart)
	cmd.Dir = info.Dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
