// Package config 负责从环境变量加载服务配置。
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Addr           string
	MaxPageSize    int
	AuthToken      string // 非空时启用 Bearer 鉴权
	RateLimitPerIP int    // 每 IP 每秒请求数上限，<=0 不限流
	JudgeDelayMs   int    // 模拟判题耗时（毫秒）
}

func Load() *Config {
	cfg := &Config{
		Addr:           ":" + getenv("PORT", "8080"),
		MaxPageSize:    getenvInt("MAX_PAGE_SIZE", 100),
		AuthToken:      os.Getenv("AUTH_TOKEN"),
		RateLimitPerIP: getenvInt("RATE_LIMIT_PER_IP", 0),
		JudgeDelayMs:   getenvInt("JUDGE_DELAY_MS", 0),
	}
	if v := os.Getenv("ADDR"); v != "" {
		cfg.Addr = v
	}
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func (c *Config) String() string {
	return fmt.Sprintf("addr=%s max_page_size=%d rate_limit=%d", c.Addr, c.MaxPageSize, c.RateLimitPerIP)
}
