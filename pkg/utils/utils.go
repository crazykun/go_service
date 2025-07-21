package utils

import (
	"context"
	"errors"
	"fmt"
	"go_service/app/config"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	portListCache     map[string]map[string]interface{}
	portListCacheTime time.Time
	portListMutex     sync.RWMutex
	cacheDuration     = 1 * time.Second // 进一步优化缓存时间为1秒

	// 连接池优化
	tcpDialer = &net.Dialer{
		Timeout:   50 * time.Millisecond, // 减少超时时间
		KeepAlive: 30 * time.Second,
	}
)

// 命令安全验证正则
var (
	dangerousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`[;&|><$\\]`),          // 危险字符
		regexp.MustCompile(`\b(rm|del|format)\b`), // 危险命令
		regexp.MustCompile(`\.\./`),               // 路径遍历
	}
)

func IntToString(i int) string {
	return strconv.Itoa(i)
}

// IsPortInUse 检查指定端口是否正在使用 - 高性能版本
func IsPortInUse(port string) (bool, error) {
	// 验证端口号
	if err := validatePortString(port); err != nil {
		return false, err
	}

	// 优先检查本地回环地址，最常用
	conn, err := tcpDialer.Dial("tcp", "127.0.0.1:"+port)
	if err == nil {
		conn.Close()
		return true, nil
	}

	// 如果本地回环失败，再检查其他地址
	addresses := []string{
		"0.0.0.0:" + port,
		"[::]:" + port,
	}

	for _, addr := range addresses {
		conn, err := tcpDialer.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return true, nil
		}
	}

	return false, nil
}

// GetPortList 获取正在监听的端口列表 - 带缓存优化
func GetPortList() (map[string]map[string]interface{}, error) {
	portListMutex.RLock()
	if portListCache != nil && time.Since(portListCacheTime) < cacheDuration {
		defer portListMutex.RUnlock()
		return portListCache, nil
	}
	portListMutex.RUnlock()

	portListMutex.Lock()
	defer portListMutex.Unlock()

	// 双重检查
	if portListCache != nil && time.Since(portListCacheTime) < cacheDuration {
		return portListCache, nil
	}

	portList, err := getPortListFromSystem()
	if err != nil {
		return nil, err
	}

	portListCache = portList
	portListCacheTime = time.Now()
	return portList, nil
}

// getPortListFromSystem 从系统获取端口列表
func getPortListFromSystem() (map[string]map[string]interface{}, error) {
	portList := make(map[string]map[string]interface{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "netstat", "-nptl")
	out, err := cmd.Output()
	if err != nil {
		return portList, fmt.Errorf("执行netstat命令失败: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 3 {
		return portList, nil
	}

	cpid := os.Getpid()
	for _, line := range lines[2:] {
		if !strings.Contains(line, "LISTEN") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 7 {
			continue
		}

		address := parts[3]
		portParts := strings.Split(address, ":")
		port := portParts[len(portParts)-1]
		if port == "" {
			continue
		}

		var pid, processName string
		process := strings.SplitN(parts[6], "/", 2)
		if len(process) == 2 {
			if IntToString(cpid) == process[0] && port != config.GlobalConfig.Server.Port {
				pid = "0"
				processName = "starting..."
			} else {
				pid = process[0]
				processName = process[1]
			}
		}

		portList[port] = map[string]interface{}{
			"pid":     pid,
			"process": processName,
		}
	}

	return portList, nil
}

func Kill(port string) (string, error) {
	// 验证端口号
	if err := validatePortString(port); err != nil {
		return "", err
	}

	// 查询进程号
	portList, err := GetPortList()
	if err != nil {
		return "", fmt.Errorf("获取端口列表失败: %v", err)
	}

	if _, ok := portList[port]; !ok {
		return "", errors.New("端口未被占用")
	}

	pidStr, ok := portList[port]["pid"].(string)
	if !ok || pidStr == "" || pidStr == "0" {
		return "", errors.New("无法获取进程ID")
	}

	// 验证PID是否为数字
	pid, err := strconv.Atoi(pidStr)
	if err != nil || pid <= 0 {
		return "", errors.New("无效的进程ID")
	}

	// 检查是否为系统关键进程
	if pid == 1 || pid == os.Getpid() {
		return "", errors.New("不能终止系统关键进程")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 先尝试优雅终止
	cmd := exec.CommandContext(ctx, "kill", "-TERM", pidStr)
	out, err := cmd.CombinedOutput()
	if err == nil {
		// 等待进程优雅退出
		time.Sleep(2 * time.Second)
		if inUse, _ := IsPortInUse(port); !inUse {
			return string(out), nil
		}
	}

	// 强制终止
	cmd = exec.CommandContext(ctx, "kill", "-9", pidStr)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("终止进程失败: %v", err)
	}

	// 清除缓存
	clearPortListCache()
	return string(out), nil
}

// WaitForPortFree 等待端口释放
func WaitForPortFree(port string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		inUse, _ := IsPortInUse(port)
		if !inUse {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("等待端口释放超时")
}

// ValidatePort 验证端口号是否有效
func ValidatePort(port int64) error {
	if port <= 0 || port > 65535 {
		return errors.New("端口号必须在1-65535之间")
	}
	if port < 1024 {
		return errors.New("建议使用1024以上的端口号")
	}
	return nil
}

// GetProcessByPort 根据端口获取进程信息
func GetProcessByPort(port string) (map[string]interface{}, error) {
	portList, err := GetPortList()
	if err != nil {
		return nil, err
	}

	if info, ok := portList[port]; ok {
		return info, nil
	}

	return nil, errors.New("端口未被占用")
}

// validatePortString 验证端口字符串
func validatePortString(port string) error {
	if port == "" {
		return errors.New("端口号不能为空")
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return errors.New("端口号必须为数字")
	}

	return ValidatePort(int64(portNum))
}

// clearPortListCache 清除端口列表缓存
func clearPortListCache() {
	portListMutex.Lock()
	defer portListMutex.Unlock()
	portListCache = nil
}

// ValidateCommand 验证命令安全性 - 增强版本
func ValidateCommand(command string) error {
	if command == "" {
		return errors.New("命令不能为空")
	}

	// 长度限制
	if len(command) > 1000 {
		return errors.New("命令长度不能超过1000字符")
	}

	// 检查危险模式
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(command) {
			return fmt.Errorf("命令包含危险字符或操作")
		}
	}

	// 检查绝对路径和相对路径安全性
	if strings.Contains(command, "../") || strings.Contains(command, "..\\") {
		return errors.New("命令不能包含路径遍历字符")
	}

	// 检查是否包含敏感文件路径
	sensitivePaths := []string{"/etc/passwd", "/etc/shadow", "/proc/", "/sys/"}
	for _, path := range sensitivePaths {
		if strings.Contains(command, path) {
			return errors.New("命令不能访问敏感系统路径")
		}
	}

	return nil
}

// SanitizeCommand 清理命令字符串
func SanitizeCommand(command string) string {
	// 移除多余空格
	command = strings.TrimSpace(command)
	command = regexp.MustCompile(`\s+`).ReplaceAllString(command, " ")
	return command
}

// IsValidDirectory 验证目录是否存在且可访问
func IsValidDirectory(dir string) error {
	if dir == "" {
		return errors.New("目录路径不能为空")
	}

	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("目录不存在: %s", dir)
		}
		return fmt.Errorf("无法访问目录: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("路径不是目录: %s", dir)
	}

	return nil
}

// GetSystemStats 获取系统统计信息
func GetSystemStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取负载信息
	if loadavg, err := os.ReadFile("/proc/loadavg"); err == nil {
		parts := strings.Fields(string(loadavg))
		if len(parts) >= 3 {
			stats["load_1m"] = parts[0]
			stats["load_5m"] = parts[1]
			stats["load_15m"] = parts[2]
		}
	}

	// 获取内存信息
	if meminfo, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(meminfo), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					stats["memory_total"] = parts[1] + " kB"
				}
			} else if strings.HasPrefix(line, "MemAvailable:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					stats["memory_available"] = parts[1] + " kB"
				}
			}
		}
	}

	// 获取CPU使用率
	if cpustat, err := os.ReadFile("/proc/stat"); err == nil {
		lines := strings.Split(string(cpustat), "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "cpu ") {
			stats["cpu_info"] = strings.Fields(lines[0])[1:5] // user, nice, system, idle
		}
	}

	// 获取磁盘使用情况
	if diskUsage := getDiskUsage("/"); diskUsage != nil {
		stats["disk_usage"] = diskUsage
	}

	stats["timestamp"] = time.Now().Unix()
	return stats
}

// getDiskUsage 获取磁盘使用情况
func getDiskUsage(path string) map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "df", "-h", path)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return nil
	}

	fields := strings.Fields(lines[1])
	if len(fields) >= 5 {
		return map[string]interface{}{
			"total": fields[1],
			"used":  fields[2],
			"avail": fields[3],
			"use%":  fields[4],
		}
	}

	return nil
}

// BatchPortCheck 批量检查端口状态
func BatchPortCheck(ports []string) map[string]bool {
	result := make(map[string]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, port := range ports {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			inUse, _ := IsPortInUse(p)
			mu.Lock()
			result[p] = inUse
			mu.Unlock()
		}(port)
	}

	wg.Wait()
	return result
}

// GetNetworkConnections 获取网络连接信息
func GetNetworkConnections() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ss", "-tuln")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取网络连接失败: %v", err)
	}

	connections := make(map[string]int)
	lines := strings.Split(string(out), "\n")

	for _, line := range lines[1:] { // 跳过标题行
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			protocol := fields[0]
			connections[protocol]++
		}
	}

	return connections, nil
}
