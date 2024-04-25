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
		return "", errors.New("port is in use")
	}

	if info.CmdStart == "" {
		return "", errors.New("start command not found")
	}

	if strings.Contains(info.CmdStart, "php") {
		args := strings.Split(info.CmdStart, " ")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = "/data/zhw/zhw_quick_qq"
		output, err := cmd.CombinedOutput()
		fmt.Println("Command output:", string(output))
		if err != nil {
			fmt.Println("Command execution error:", err)
			return "", fmt.Errorf(string(output), err)
		}
		fmt.Println("Command output:", string(output))
		return string(output), err
	} else {
		cmd := exec.Command("bash", "-c", info.CmdStart)
		cmd.Dir = info.Dir
		output, err := cmd.CombinedOutput()
		fmt.Println("Command output:", string(output))
		if err != nil {
			fmt.Println("Command execution error:", err)
			return string(output), err
		}
		fmt.Println("Command output:", string(output))
		return string(output), err
	}
}

func (l Logic) Stop(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("service not found")
	}
	port := strconv.Itoa(int(info.Port))
	isUse, _ := utils.IsPortInUse(port)
	if !isUse {
		return "", errors.New("port is not in use")
	}

	if info.CmdStop == "" {
		return utils.Kill(port)
	} else {
		// 进入目录并启动服务
		args := strings.Split(info.CmdStart, " ")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = info.Dir

		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}
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
	args := strings.Split(info.CmdRestart, " ")
	cmd := exec.Command(args[0], args[1:]...)
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
		utils.Kill(port)
	}

	// 进入目录并启动服务
	args := strings.Split(info.CmdStart, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = info.Dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (l Logic) Kill(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("service not found")
	}
	port := strconv.Itoa(int(info.Port))
	// 查询进程号
	out, err := utils.Kill(port)
	return string(out), err
}
