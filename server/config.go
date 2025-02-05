package server

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host       string     `yaml:"Host"`
	Port       int        `yaml:"Port"`
	GinMode   string     `yaml:"GinMode"`
	TextConfig TextConfig `yaml:"TextConfig"`
	P5cc       P5ccConfig `yaml:"p5cc"`
}

type TextConfig struct {
	HelloText string `yaml:"helloText"`
}

type P5ccConfig struct {
	FontSize   float64 `yaml:"fontSize"`
	FontFamily string  `yaml:"fontFamily"`
	Gutter     float64 `yaml:"gutter"`
	Padding    float64 `yaml:"padding"`
	TextAlign  string  `yaml:"textAlign"`

	ShowLogo   bool    `yaml:"showLogo"`
	LogoScale  float64 `yaml:"logoScale"`
	LogoOffset int     `yaml:"logoOffset"`

	ShowWtm string `yaml:"showWtm"`
}

var AppConfig *Config

// 初始化配置
func InitConfig(configPath string) error {
	if AppConfig != nil {
		return nil // 避免重复加载
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析 YAML
	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	// 验证配置
	if AppConfig.Port <= 0 {
		return fmt.Errorf("invalid port: %d", AppConfig.Port)
	}

	return nil
}
