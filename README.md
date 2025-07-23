# Go 服务管理工具

基于 Go 构建的综合服务管理平台，提供基于 Web 的服务监控、控制和管理功能，具备实时健康检查和性能监控能力。

## ✨ 功能特性

### 🚀 服务管理

- **多服务控制**: 一键启动、停止、重启和终止服务
- **批量操作**: 同时管理多个服务
- **自动重启**: 服务故障时自动恢复
- **进程监控**: 实时进程状态和资源使用情况
- **端口管理**: 自动端口检测和冲突解决

## 📸 界面截图

![控制台](https://raw.githubusercontent.com/crazykun/go_service/main/static/img/image.png)
---

## 🚀 快速开始

### 环境要求

- Go 1.19+
- MySQL 5.7+ 或 MariaDB 10.3+
- Linux/macOS (支持 Windows)

### 安装步骤

1. **克隆仓库**

```bash
git clone https://github.com/crazykun/go_service.git
cd go_service
```

2. **安装依赖**

```bash
go mod download
```

3. **配置数据库**

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE go_service;"

# 导入数据库结构
mysql -u root -p go_service < db.sql
```

4. **配置应用**

```bash
# 编辑配置文件
cp config.yml.example config.yml
vim config.yml
```

5. **运行应用**

```bash
# 开发模式
go run main.go
```

```bash
# 生产构建
go build -o go_service main.go
./go_service
```

```bash
# 使用重启脚本
chmod +x restart.sh
./restart.sh
```

### Docker 部署

```bash
# 使用 Docker Compose
docker-compose up -d

# 或手动构建
docker build -t go-service .
docker run -p 10000:10000 go-service
```

## 📖 使用说明

### Web 界面

访问 Web 控制台: `http://localhost:10000`

### API 接口

#### 服务管理

```bash
# 添加新服务
curl -X POST http://localhost:10000/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app",
    "title": "我的应用",
    "dir": "/path/to/app",
    "cmd_start": "npm start",
    "port": 3000
  }'

# 获取所有服务
curl http://localhost:10000/api/v1/services

# 启动服务
curl -X POST http://localhost:10000/api/v1/operations/start/1

# 获取服务状态
curl http://localhost:10000/api/v1/operations/status/1
```


#### 批量操作

```bash
# 启动所有服务
curl -X POST http://localhost:10000/api/v1/batch/start-all

# 批量操作指定服务
curl -X POST http://localhost:10000/api/v1/batch/operation \
  -H "Content-Type: application/json" \
  -d '{
    "service_ids": [1, 2, 3],
    "operation": "restart"
  }'
```

完整的 API 文档请参考 [API_EXAMPLES.md](API_EXAMPLES.md)

## ⚙️ 配置说明

### 基础配置 (config.yml)

```
app:
  name: "go_service"
  mode: "debug"  # debug, release, test
  version: "1.0.0"

server:
  host: "127.0.0.1"
  port: "10000"
  read_timeout: "30s"
  write_timeout: "30s"

database:
  default:
    host: "127.0.0.1"
    port: 3306
    user: "root"
    pwd: "password"
    name: "go_service"

monitor:
  enabled: true
  check_interval: "30s"
  retention_days: 7

security:
  rate_limit_enabled: true
  rate_limit_rps: 100
  enable_auth: false
```

### 环境变量

```bash
export GIN_MODE=release
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=password
export DB_NAME=go_service
```


## 🛠️ 开发指南

#### 项目结构

```html
go_service/
├── app/
│   ├── controller/     # HTTP 处理器
│   ├── service/        # 业务逻辑
│   ├── model/          # 数据模型
│   ├── middleware/     # HTTP 中间件
│   └── global/         # 全局配置
├── pkg/
│   ├── utils/          # 工具函数
│   └── pool/           # 连接池
├── static/             # 静态资源
├── template/           # HTML 模板
├── config.yml          # 配置文件
└── main.go            # 应用入口
```

#### 从源码构建

```bash
# 安装依赖
go mod tidy
```

```bash
# 运行测试
go test ./...
```

```bash
# 构建当前平台
go build -o go_service main.go
```

```bash
# 交叉编译 Linux 版本
GOOS=linux GOARCH=amd64 go build -o go_service-linux main.go
```

```bash
# 优化构建
go build -ldflags="-s -w" -o go_service main.go
```


#### 贡献代码

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature')
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request


### 🐛 故障排查

#### 常见问题

**服务无法启动**

```bash
# 检查端口是否被占用
netstat -tulpn | grep :10000

# 查看日志
tail -f logs/app.log

# 验证配置
go run main.go --config-check
```

**数据库连接失败**

```bash
# 测试数据库连接
mysql -h localhost -u root -p go_service
```

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🤝 技术支持

- **文档**: 查看我们的详细指南
- **问题反馈**: 在 [GitHub Issues](https://github.com/crazykun/go_service/issues) 报告 Bug
- **讨论交流**: 加入 [GitHub Discussions](https://github.com/crazykun/go_service/discussions)



⭐ **如果这个项目对您有帮助，请给个 Star！**
