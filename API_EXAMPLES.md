# Go服务管理工具 API 使用示例

## 🚀 快速开始

### 启动服务
```bash
# 编译并启动
go build -o go_service main.go
./go_service

# 或直接运行
go run main.go
```

服务启动后，访问 http://localhost:10000 查看管理界面。

## 📋 API 接口文档

### 1. 服务管理

#### 添加服务
```bash
curl -X POST http://localhost:10000/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-web-service",
    "title": "我的Web服务",
    "dir": "/home/user/my-app",
    "cmd_start": "npm start",
    "cmd_stop": "npm stop",
    "port": 3000,
    "health_check_url": "http://localhost:3000/health",
    "auto_restart": true,
    "remark": "Node.js Web应用"
  }'
```

#### 获取服务列表
```bash
# 获取所有服务
curl http://localhost:10000/api/v1/services

# 分页查询
curl "http://localhost:10000/api/v1/services?page=1&page_size=10"

# 按名称搜索
curl "http://localhost:10000/api/v1/services?name=web"

# 按状态过滤 (0:停止, 1:运行)
curl "http://localhost:10000/api/v1/services?status=1"
```

#### 获取单个服务
```bash
curl http://localhost:10000/api/v1/services/1
```

#### 更新服务
```bash
curl -X PUT http://localhost:10000/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "name": "my-web-service",
    "title": "我的Web服务(更新)",
    "dir": "/home/user/my-app",
    "cmd_start": "npm start",
    "port": 3000,
    "remark": "更新后的备注"
  }'
```

#### 删除服务
```bash
curl -X DELETE http://localhost:10000/api/v1/services/1
```

### 2. 服务操作

#### 启动服务
```bash
curl -X POST http://localhost:10000/api/v1/operations/start/1
```

#### 停止服务
```bash
curl -X POST http://localhost:10000/api/v1/operations/stop/1
```

#### 重启服务
```bash
curl -X POST http://localhost:10000/api/v1/operations/restart/1
```

#### 强制重启服务
```bash
curl -X POST http://localhost:10000/api/v1/operations/force-restart/1
```

#### 强制终止服务
```bash
curl -X POST http://localhost:10000/api/v1/operations/kill/1
```

#### 获取服务状态
```bash
curl http://localhost:10000/api/v1/operations/status/1
```

### 3. 批量操作

#### 批量启动服务
```bash
curl -X POST http://localhost:10000/api/v1/batch/operation \
  -H "Content-Type: application/json" \
  -d '{
    "service_ids": [1, 2, 3],
    "operation": "start"
  }'
```

#### 启动所有服务
```bash
curl -X POST http://localhost:10000/api/v1/batch/start-all
```

#### 停止所有服务
```bash
curl -X POST http://localhost:10000/api/v1/batch/stop-all
```

#### 获取批量状态
```bash
curl http://localhost:10000/api/v1/batch/status
```


## 📊 响应格式

### 成功响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 具体数据
  }
}
```

### 错误响应
```json
{
  "code": 1001,
  "message": "参数错误",
  "data": null
}
```

### 分页响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "size": 20
  }
}
```

## 🔧 配置示例

### config.yml
```yaml
app:
  name: "go_service"
  mode: "debug"
  version: "1.0.0"
  environment: "development"

server:
  host: "127.0.0.1"
  port: "10000"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"

database:
  default:
    host: "127.0.0.1"
    port: 3306
    user: "root"
    pwd: "password"
    name: "go_service"
    maxIdle: 10
    maxOpen: 100
    maxLifetime: "1h"
    connMaxIdleTime: "10m"

monitor:
  enabled: true
  check_interval: "30s"
  timeout: "10s"
  retention_days: 7

security:
  rate_limit_enabled: true
  rate_limit_rps: 100
```

## 🐳 Docker 部署

### Dockerfile
```dockerfile
FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o go_service main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/go_service .
COPY --from=builder /app/config.yml .
COPY --from=builder /app/template ./template
COPY --from=builder /app/static ./static

EXPOSE 10000

CMD ["./go_service"]
```

### docker-compose.yml
```yaml
version: '3.8'

services:
  go-service:
    build: .
    ports:
      - "10000:10000"
    environment:
      - GIN_MODE=release
    volumes:
      - ./config.yml:/root/config.yml
      - ./logs:/root/logs
    depends_on:
      - mysql
    restart: unless-stopped

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: go_service
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./db_migration.sql:/docker-entrypoint-initdb.d/init.sql
    restart: unless-stopped

volumes:
  mysql_data:
```

## 🚀 部署脚本

### deploy.sh
```bash
#!/bin/bash

# 构建应用
echo "构建应用..."
go build -o go_service main.go

# 停止旧服务
echo "停止旧服务..."
pkill -f go_service || true

# 备份配置
echo "备份配置..."
cp config.yml config.yml.backup.$(date +%Y%m%d%H%M%S)

# 启动新服务
echo "启动新服务..."
nohup ./go_service > logs/app.log 2>&1 &

# 等待服务启动
sleep 3

# 健康检查
echo "健康检查..."
if curl -f http://localhost:10000/health; then
    echo "服务启动成功!"
else
    echo "服务启动失败!"
    exit 1
fi
```


