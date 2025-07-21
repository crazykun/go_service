package config

import (
	"fmt"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	App      AppConfig                 `mapstructure:"app"`
	Server   ServerConfig              `mapstructure:"server"`
	Database map[string]DatabaseConfig `mapstructure:"database"`
	Redis    map[string]RedisConfig    `mapstructure:"redis"`
	Log      LogConfig                 `mapstructure:"log"`
	Monitor  MonitorConfig             `mapstructure:"monitor"`
	Security SecurityConfig            `mapstructure:"security"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Mode        string `mapstructure:"mode"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"pwd"`
	Name            string        `mapstructure:"name"`
	MaxIdle         int           `mapstructure:"maxIdle"`
	MaxOpen         int           `mapstructure:"maxOpen"`
	MaxLifetime     time.Duration `mapstructure:"maxLifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"connMaxIdleTime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Password    string        `mapstructure:"pwd"`
	DB          int           `mapstructure:"db"`
	MaxIdle     int           `mapstructure:"maxIdle"`
	MaxActive   int           `mapstructure:"maxActive"`
	IdleTimeout time.Duration `mapstructure:"idleTimeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	HealthCheckPath string        `mapstructure:"health_check_path"`
	MetricsPath     string        `mapstructure:"metrics_path"`
	CheckInterval   time.Duration `mapstructure:"check_interval"`
	Timeout         time.Duration `mapstructure:"timeout"`
	AlertWebhook    string        `mapstructure:"alert_webhook"`
	RetentionDays   int           `mapstructure:"retention_days"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableAuth       bool          `mapstructure:"enable_auth"`
	JWTSecret        string        `mapstructure:"jwt_secret"`
	TokenExpiry      time.Duration `mapstructure:"token_expiry"`
	RateLimitEnabled bool          `mapstructure:"rate_limit_enabled"`
	RateLimitRPS     int           `mapstructure:"rate_limit_rps"`
	AllowedIPs       []string      `mapstructure:"allowed_ips"`
	TLSEnabled       bool          `mapstructure:"tls_enabled"`
	CertFile         string        `mapstructure:"cert_file"`
	KeyFile          string        `mapstructure:"key_file"`
}

var (
	GlobalConfig *Config
	configPath   string
)

// InitConfig 初始化配置 - 优化版本
func InitConfig(configFile string) error {
	if configFile == "" {
		configFile = "config.yml"
	}
	configPath = configFile

	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，创建默认配置文件
		if os.IsNotExist(err) {
			if createErr := createDefaultConfig(configFile); createErr != nil {
				return fmt.Errorf("创建默认配置文件失败: %v", createErr)
			}
			if err := viper.ReadInConfig(); err != nil {
				return fmt.Errorf("读取默认配置文件失败: %v", err)
			}
		} else {
			return fmt.Errorf("读取配置文件失败: %v", err)
		}
	}

	// 解析配置
	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}

	// 验证配置
	if err := validateConfig(GlobalConfig); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	// 监听配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("配置文件发生变化: %s\n", e.Name)
		oldConfig := GlobalConfig
		newConfig := &Config{}
		if err := viper.Unmarshal(newConfig); err != nil {
			fmt.Printf("重新加载配置失败: %v\n", err)
			return
		}

		// 验证新配置
		if err := validateConfig(newConfig); err != nil {
			fmt.Printf("新配置验证失败: %v\n", err)
			return
		}

		GlobalConfig = newConfig
		fmt.Printf("配置重新加载成功\n")

		// 可以在这里添加配置变更的回调处理
		handleConfigChange(oldConfig, newConfig)
	})

	return nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 应用默认配置
	viper.SetDefault("app.name", "go_service")
	viper.SetDefault("app.mode", "debug")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")

	// 服务器默认配置
	viper.SetDefault("server.host", "127.0.0.1")
	viper.SetDefault("server.port", "10000")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")

	// 数据库默认配置
	viper.SetDefault("database.default.host", "127.0.0.1")
	viper.SetDefault("database.default.port", 3306)
	viper.SetDefault("database.default.maxIdle", 10)
	viper.SetDefault("database.default.maxOpen", 100)
	viper.SetDefault("database.default.maxLifetime", "1h")
	viper.SetDefault("database.default.connMaxIdleTime", "10m")

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 3)
	viper.SetDefault("log.max_age", 28)
	viper.SetDefault("log.compress", true)

	// 监控默认配置
	viper.SetDefault("monitor.enabled", true)
	viper.SetDefault("monitor.health_check_path", "/health")
	viper.SetDefault("monitor.metrics_path", "/metrics")
	viper.SetDefault("monitor.check_interval", "30s")
	viper.SetDefault("monitor.timeout", "10s")
	viper.SetDefault("monitor.retention_days", 7)

	// 安全默认配置
	viper.SetDefault("security.enable_auth", false)
	viper.SetDefault("security.token_expiry", "24h")
	viper.SetDefault("security.rate_limit_enabled", true)
	viper.SetDefault("security.rate_limit_rps", 100)
	viper.SetDefault("security.tls_enabled", false)
}

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	// 验证服务器配置
	if cfg.Server.Port == "" {
		return fmt.Errorf("服务器端口不能为空")
	}

	// 验证数据库配置
	if len(cfg.Database) == 0 {
		return fmt.Errorf("至少需要配置一个数据库")
	}

	for name, db := range cfg.Database {
		if db.Host == "" || db.Name == "" || db.User == "" {
			return fmt.Errorf("数据库 %s 配置不完整", name)
		}
	}

	// 验证安全配置
	if cfg.Security.EnableAuth && cfg.Security.JWTSecret == "" {
		return fmt.Errorf("启用认证时必须配置JWT密钥")
	}

	if cfg.Security.TLSEnabled && (cfg.Security.CertFile == "" || cfg.Security.KeyFile == "") {
		return fmt.Errorf("启用TLS时必须配置证书文件")
	}

	return nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return GlobalConfig
}

// IsDevelopment 是否为开发环境
func IsDevelopment() bool {
	return GlobalConfig.App.Environment == "development"
}

// IsProduction 是否为生产环境
func IsProduction() bool {
	return GlobalConfig.App.Environment == "production"
}

// GetDatabaseConfig 获取数据库配置
func GetDatabaseConfig(name string) (*DatabaseConfig, error) {
	if cfg, ok := GlobalConfig.Database[name]; ok {
		return &cfg, nil
	}
	return nil, fmt.Errorf("数据库配置 %s 不存在", name)
}

// GetRedisConfig 获取Redis配置
func GetRedisConfig(name string) (*RedisConfig, error) {
	if cfg, ok := GlobalConfig.Redis[name]; ok {
		return &cfg, nil
	}
	return nil, fmt.Errorf("redis配置 %s 不存在", name)
}

// ReloadConfig 重新加载配置
func ReloadConfig() error {
	return InitConfig(configPath)
} // c
// reateDefaultConfig 创建默认配置文件
func createDefaultConfig(configFile string) error {
	defaultConfig := `# Go服务管理工具配置文件

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
    pwd: ""
    name: "go_service"
    maxIdle: 10
    maxOpen: 100
    maxLifetime: "1h"
    connMaxIdleTime: "10m"

log:
  level: "info"
  format: "json"
  output: "stdout"
  max_size: 100
  max_backups: 3
  max_age: 28
  compress: true

monitor:
  enabled: true
  health_check_path: "/health"
  metrics_path: "/metrics"
  check_interval: "30s"
  timeout: "10s"
  retention_days: 7
  alert_webhook: ""

security:
  enable_auth: false
  jwt_secret: ""
  token_expiry: "24h"
  rate_limit_enabled: true
  rate_limit_rps: 100
  allowed_ips: []
  tls_enabled: false
  cert_file: ""
  key_file: ""
`

	return os.WriteFile(configFile, []byte(defaultConfig), 0644)
}

// handleConfigChange 处理配置变更
func handleConfigChange(oldConfig, newConfig *Config) {
	// 检查关键配置是否发生变化
	if oldConfig.Server.Port != newConfig.Server.Port {
		fmt.Printf("服务器端口从 %s 变更为 %s，需要重启服务器\n", oldConfig.Server.Port, newConfig.Server.Port)
	}

	if oldConfig.Monitor.Enabled != newConfig.Monitor.Enabled {
		if newConfig.Monitor.Enabled {
			fmt.Printf("监控功能已启用\n")
		} else {
			fmt.Printf("监控功能已禁用\n")
		}
	}

	if oldConfig.Security.RateLimitEnabled != newConfig.Security.RateLimitEnabled {
		if newConfig.Security.RateLimitEnabled {
			fmt.Printf("限流功能已启用，RPS: %d\n", newConfig.Security.RateLimitRPS)
		} else {
			fmt.Printf("限流功能已禁用\n")
		}
	}
}

// UpdateConfig 更新配置
func UpdateConfig(key string, value interface{}) error {
	viper.Set(key, value)

	// 重新解析配置
	newConfig := &Config{}
	if err := viper.Unmarshal(newConfig); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}

	// 验证新配置
	if err := validateConfig(newConfig); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	// 保存配置到文件
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	oldConfig := GlobalConfig
	GlobalConfig = newConfig

	// 处理配置变更
	handleConfigChange(oldConfig, newConfig)

	return nil
}

// GetConfigValue 获取配置值
func GetConfigValue(key string) interface{} {
	return viper.Get(key)
}

// GetConfigString 获取字符串配置值
func GetConfigString(key string) string {
	return viper.GetString(key)
}

// GetConfigInt 获取整数配置值
func GetConfigInt(key string) int {
	return viper.GetInt(key)
}

// GetConfigBool 获取布尔配置值
func GetConfigBool(key string) bool {
	return viper.GetBool(key)
}

// GetConfigDuration 获取时间间隔配置值
func GetConfigDuration(key string) time.Duration {
	return viper.GetDuration(key)
}

// GetAllConfig 获取所有配置
func GetAllConfig() map[string]interface{} {
	return viper.AllSettings()
}

// ValidateConfigFile 验证配置文件格式
func ValidateConfigFile(configFile string) error {
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}

	return validateConfig(config)
}

// BackupConfig 备份当前配置
func BackupConfig() error {
	if configPath == "" {
		return fmt.Errorf("配置文件路径未设置")
	}

	backupPath := configPath + ".backup." + time.Now().Format("20060102150405")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("创建备份文件失败: %v", err)
	}

	fmt.Printf("配置文件已备份到: %s\n", backupPath)
	return nil
}

// RestoreConfig 恢复配置
func RestoreConfig(backupPath string) error {
	// 验证备份文件
	if err := ValidateConfigFile(backupPath); err != nil {
		return fmt.Errorf("备份文件验证失败: %v", err)
	}

	// 备份当前配置
	if err := BackupConfig(); err != nil {
		fmt.Printf("警告: 备份当前配置失败: %v\n", err)
	}

	// 复制备份文件到当前配置文件
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("恢复配置文件失败: %v", err)
	}

	// 重新加载配置
	if err := ReloadConfig(); err != nil {
		return fmt.Errorf("重新加载配置失败: %v", err)
	}

	fmt.Printf("配置已从 %s 恢复\n", backupPath)
	return nil
}
