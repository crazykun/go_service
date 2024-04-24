package utils

import (
	"net"
	"os/exec"
	"strings"
)

// IsPortInUse 检查指定端口是否正在使用
func IsPortInUse(port string) (bool, error) {
	_, err := net.Listen("tcp", ":"+port)
	return err != nil, err
}

// GetPortList 获取正在监听的端口列表及其对应进程的信息。
// 返回一个映射，其键为端口号，值为包含进程ID和进程名称的映射，以及一个可能出现的错误。
// - port:[pid:进程id, pid:进程名] 端口到进程信息的映射。
func GetPortList() (map[string]map[string]interface{}, error) {
	portList := make(map[string]map[string]interface{}, 0)
	out, err := exec.Command("netstat", "-nptl").Output()
	if err != nil {
		return portList, err
	}
	outList := strings.Split(string(out), "\n")
	outList = outList[2:]
	for _, line := range outList {
		if strings.Contains(line, "LISTEN") {
			parts := strings.Fields(line)
			address := parts[3]
			ports := strings.Split(address, ":")
			port := ports[len(ports)-1]
			if port == "" || len(parts) < 7 {
				continue
			}
			var pid, process_name string
			process := strings.SplitN(parts[6], "/", 2)
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
	return portList, nil
}
