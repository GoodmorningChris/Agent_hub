package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
	Redis  RedisConfig
	JWT    JWTConfig
	Log    LogConfig
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
}

// DSN 返回 GORM MySQL DSN
func (c MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database, c.Charset)
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load 从 configs 目录加载配置，环境变量可覆盖
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("configs")
	v.AddConfigPath(".")
	v.AddConfigPath("../configs") // 从 cmd/server 运行时
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定环境变量
	bindEnv(v, "server.port", "SERVER_PORT")
	bindEnv(v, "mysql.host", "MYSQL_HOST")
	bindEnv(v, "mysql.port", "MYSQL_PORT")
	bindEnv(v, "mysql.user", "MYSQL_USER")
	bindEnv(v, "mysql.password", "MYSQL_PASSWORD")
	bindEnv(v, "mysql.database", "MYSQL_DATABASE")
	bindEnv(v, "redis.addr", "REDIS_ADDR")
	bindEnv(v, "redis.password", "REDIS_PASSWORD")
	bindEnv(v, "jwt.secret", "JWT_SECRET")
	bindEnv(v, "jwt.expire_hours", "JWT_EXPIRE_HOURS")
	bindEnv(v, "log.level", "LOG_LEVEL")
	bindEnv(v, "log.format", "LOG_FORMAT")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func bindEnv(v *viper.Viper, key, envKey string) {
	_ = v.BindEnv(key, envKey)
}
