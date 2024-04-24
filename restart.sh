#!/bin/bash
echo "go build"
go mod tidy
go build -o go_service
chmod +x ./go_service
echo "kill go_service service"
killall go_service # kill go-admin service
filename="./logs/output.log"
nohup ./go_service >> $filename 2>&1 & #后台启动服务将日志写入output.log文件
echo "run go_service success"
ps -aux | grep go_service
