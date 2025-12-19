package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	WAF      WAFConfig      `yaml:"waf"`
	SSL      SSLConfig      `yaml:"ssl"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	AdminPort    int           `yaml:"admin_port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

type WAFConfig struct {
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	HTTPFlood HTTPFloodConfig `yaml:"http_flood"`
	AntiBot   AntiBotConfig   `yaml:"anti_bot"`
	GeoIP     GeoIPConfig     `yaml:"geoip"`
}

type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
}

type HTTPFloodConfig struct {
	Enabled              bool `yaml:"enabled"`
	MaxRequestsPerMinute int  `yaml:"max_requests_per_minute"`
	BlockDuration        int  `yaml:"block_duration"`
}

type AntiBotConfig struct {
	Enabled             bool     `yaml:"enabled"`
	ChallengeMode       string   `yaml:"challenge_mode"`
	WhitelistUserAgents []string `yaml:"whitelist_user_agents"`
	BlacklistUserAgents []string `yaml:"blacklist_user_agents"`
}

type GeoIPConfig struct {
	Enabled      bool   `yaml:"enabled"`
	DatabasePath string `yaml:"database_path"`
}

type SSLConfig struct {
	AutoCert bool   `yaml:"auto_cert"`
	CertDir  string `yaml:"cert_dir"`
}

type LoggingConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	// Try to load from YAML file first
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, err
			}
		}
	}

	// Override with environment variables
	cfg.overrideFromEnv()

	return &cfg, nil
}

func (c *Config) overrideFromEnv() {
	// Server
	if val := os.Getenv("SERVER_HOST"); val != "" {
		c.Server.Host = val
	}
	if val := os.Getenv("SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			c.Server.Port = port
		}
	}
	if val := os.Getenv("SERVER_ADMIN_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			c.Server.AdminPort = port
		}
	}
	if val := os.Getenv("SERVER_READ_TIMEOUT"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			c.Server.ReadTimeout = duration
		}
	}
	if val := os.Getenv("SERVER_WRITE_TIMEOUT"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			c.Server.WriteTimeout = duration
		}
	}

	// Database
	if val := os.Getenv("DATABASE_DRIVER"); val != "" {
		c.Database.Driver = val
	}
	if val := os.Getenv("DATABASE_HOST"); val != "" {
		c.Database.Host = val
	}
	if val := os.Getenv("DATABASE_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			c.Database.Port = port
		}
	}
	if val := os.Getenv("DATABASE_NAME"); val != "" {
		c.Database.Name = val
	}
	if val := os.Getenv("DATABASE_USER"); val != "" {
		c.Database.User = val
	}
	if val := os.Getenv("DATABASE_PASSWORD"); val != "" {
		c.Database.Password = val
	}
	if val := os.Getenv("DATABASE_SSLMODE"); val != "" {
		c.Database.SSLMode = val
	}

	// Redis
	if val := os.Getenv("REDIS_HOST"); val != "" {
		c.Redis.Host = val
	}
	if val := os.Getenv("REDIS_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			c.Redis.Port = port
		}
	}
	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		c.Redis.Password = val
	}
	if val := os.Getenv("REDIS_DB"); val != "" {
		if db, err := strconv.Atoi(val); err == nil {
			c.Redis.DB = db
		}
	}
	if val := os.Getenv("REDIS_POOL_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			c.Redis.PoolSize = size
		}
	}

	// WAF - Rate Limit
	if val := os.Getenv("WAF_RATE_LIMIT_ENABLED"); val != "" {
		c.WAF.RateLimit.Enabled = val == "true"
	}
	if val := os.Getenv("WAF_RATE_LIMIT_RPS"); val != "" {
		if rps, err := strconv.Atoi(val); err == nil {
			c.WAF.RateLimit.RequestsPerSecond = rps
		}
	}
	if val := os.Getenv("WAF_RATE_LIMIT_BURST"); val != "" {
		if burst, err := strconv.Atoi(val); err == nil {
			c.WAF.RateLimit.Burst = burst
		}
	}

	// WAF - HTTP Flood
	if val := os.Getenv("WAF_HTTP_FLOOD_ENABLED"); val != "" {
		c.WAF.HTTPFlood.Enabled = val == "true"
	}
	if val := os.Getenv("WAF_HTTP_FLOOD_MAX_RPM"); val != "" {
		if rpm, err := strconv.Atoi(val); err == nil {
			c.WAF.HTTPFlood.MaxRequestsPerMinute = rpm
		}
	}
	if val := os.Getenv("WAF_HTTP_FLOOD_BLOCK_DURATION"); val != "" {
		if duration, err := strconv.Atoi(val); err == nil {
			c.WAF.HTTPFlood.BlockDuration = duration
		}
	}

	// WAF - Anti Bot
	if val := os.Getenv("WAF_ANTI_BOT_ENABLED"); val != "" {
		c.WAF.AntiBot.Enabled = val == "true"
	}
	if val := os.Getenv("WAF_ANTI_BOT_CHALLENGE_MODE"); val != "" {
		c.WAF.AntiBot.ChallengeMode = val
	}

	// WAF - GeoIP
	if val := os.Getenv("WAF_GEOIP_ENABLED"); val != "" {
		c.WAF.GeoIP.Enabled = val == "true"
	}
	if val := os.Getenv("WAF_GEOIP_DATABASE_PATH"); val != "" {
		c.WAF.GeoIP.DatabasePath = val
	}

	// SSL
	if val := os.Getenv("SSL_AUTO_CERT"); val != "" {
		c.SSL.AutoCert = val == "true"
	}
	if val := os.Getenv("SSL_CERT_DIR"); val != "" {
		c.SSL.CertDir = val
	}

	// Logging
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		c.Logging.Level = val
	}
	if val := os.Getenv("LOG_FORMAT"); val != "" {
		c.Logging.Format = val
	}
	if val := os.Getenv("LOG_OUTPUT"); val != "" {
		c.Logging.Output = val
	}
	if val := os.Getenv("LOG_FILE_PATH"); val != "" {
		c.Logging.FilePath = val
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c *Config) GetAdminAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.AdminPort)
}
