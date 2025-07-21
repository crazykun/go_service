package logic

import (
	"context"
	"errors"
	"fmt"
	"go_service/pkg/utils"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func (l Logic) Start(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("服务不存在")
	}
	
	if info.CmdStart == "" {
		return "", errors.New("启动命令未配置")
	}

	port := strconv.Itoa(int(info.Port))
	isUse, _ := utils.IsPortInUse(port)
	if isUse {
		return "", errors.New("端口已被占用")
	}

	// 执行启动命令
	return l.executeCommand(info.CmdStart, info.Dir)
}

func (l Logic) Stop(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("服务不存在")
	}
	
	port := strconv.Itoa(int(info.Port))
	isUse, _ := utils.IsPortInUse(port)
	if !isUse {
		return "", errors.New("服务未运行")
	}

	// 如果有停止命令，使用停止命令，否则强制终止
	if info.CmdStop != "" {
		return l.executeCommand(info.CmdStop, info.Dir)
	}
	
	return utils.Kill(port)
}

func (l Logic) Restart(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("服务不存在")
	}
	
	port := strconv.Itoa(int(info.Port))
	isUse, _ := utils.IsPortInUse(port)
	if !isUse {
		return "", errors.New("服务未运行")
	}
	
	// 如果有重启命令，使用重启命令，否则先停止再启动
	if info.CmdRestart != "" {
		return l.executeCommand(info.CmdRestart, info.Dir)
	}
	
	// 没有重启命令，先停止再启动
	_, err := l.Stop(ctx, id)
	if err != nil {
		return "", fmt.Errorf("停止服务失败: %v", err)
	}
	
	// 等待端口释放后再启动
	return l.Start(ctx, id)
}

func (l Logic) ForcedRestart(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("服务不存在")
	}
	
	if info.CmdStart == "" {
		return "", errors.New("启动命令未配置")
	}
	
	port := strconv.Itoa(int(info.Port))
	isUse, _ := utils.IsPortInUse(port)
	if isUse {
		// 强制终止现有进程
		_, err := utils.Kill(port)
		if err != nil {
			return "", fmt.Errorf("强制终止进程失败: %v", err)
		}
	}

	// 执行启动命令
	return l.executeCommand(info.CmdStart, info.Dir)
}

func (l Logic) Kill(ctx context.Context, id int64) (string, error) {
	info := l.GetById(ctx, id)
	if info.Id == 0 {
		return "", errors.New("服务不存在")
	}
	port := strconv.Itoa(int(info.Port))
	// 查询进程号并强制终止
	out, err := utils.Kill(port)
	return out, err
}

// executeCommand 统一的命令执行方法
func (l Logic) executeCommand(command, workDir string) (string, error) {
	if command == "" {
		return "", errors.New("命令不能为空")
	}

	// 清理多余空格
	command = strings.TrimSpace(strings.ReplaceAll(command, "  ", " "))
	
	var cmd *exec.Cmd
	
	// 根据命令类型选择执行方式
	if strings.Contains(command, "php") || !strings.Contains(command, " ") {
		// PHP命令或单个命令，直接分割参数
		args := strings.Fields(command)
		if len(args) == 0 {
			return "", errors.New("无效的命令")
		}
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		// 复杂命令，使用bash执行
		cmd = exec.Command("bash", "-c", command)
	}
	
	// 设置工作目录
	if workDir != "" {
		cmd.Dir = workDir
	}
	
	// 设置子进程独立于父进程
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // 创建新进程组，子进程不再依赖父进程
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %v, 输出: %s", err, string(output))
	}
	
	return string(output), nil
}
