package conf

import (
	"fmt"
	"os"

	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Host       string      `yaml:"Host"`
	Port       int         `yaml:"Port"`
	GinMode    string      `yaml:"GinMode"`
	ConfigMode string      `yaml:"ConfigMode"`
	TextConfig TextConfig  `yaml:"TextConfig"`
	P5cc       P5ccConfig  `yaml:"P5cc"`
	Wxapi      WxapiConfig `yaml:"WxApi"`
	AIConfig   AIConfig    `yaml:"AIConfig"`
	CHApi      CHApiConfig `yaml:"CHApi"`
}

type CHApiConfig struct {
	Enabled         bool   `yaml:"enabled"`
	DatabaseURL     string `yaml:"databaseURL"`
	SessionTTLHours int    `yaml:"sessionTTLHours"`
	PasswordPepper  string `yaml:"passwordPepper"`
	AllowedOrigin   string `yaml:"allowedOrigin"`
}

type TextConfig struct {
	HelloText string `yaml:"helloText"`
}

type P5ccConfig struct {
	AssetPath  string  `yaml:"assetPath"`
	FontSize   float64 `yaml:"fontSize"`
	FontFamily string  `yaml:"fontFamily"`
	Gutter     float64 `yaml:"gutter"`
	Padding    float64 `yaml:"padding"`
	TextAlign  string  `yaml:"textAlign"`
	RedProb    float64 `yaml:"redProb"`

	ShowLogo   bool    `yaml:"showLogo"`
	LogoScale  float64 `yaml:"logoScale"`
	LogoOffset int     `yaml:"logoOffset"`

	ShowWtm string `yaml:"showWtm"`
}

type WxapiConfig struct {
	ReplyPostURL    string           `yaml:"replyPostURL"`
	OfficialAccount offConfig.Config `yaml:"official_account"`
	Text            wxText           `yaml:"text"`
}
type wxText struct {
	HelloText   string `yaml:"helloText"`
	SimplayText string `yaml:"simplayText"`
	DefaultText string `yaml:"defaultText"`
}

type AIConfig struct {
	ChatGPTUrlProxy string `yaml:"ChatGPTUrlProxy"`
	DeepSeekUrl     string `yaml:"DeepSeekUrl"`
}

var AppConfig *Config

// 初始化配置
func InitConfig() error {
	if AppConfig != nil {
		return nil // 避免重复加载
	}

	// 读取配置文件
	data, err := os.ReadFile("conf/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析 YAML
	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	// 加载环境变量覆盖配置
	if AppConfig.ConfigMode == "env" {
		loadEnvConfig()
	}

	// 验证配置
	if AppConfig.Port <= 0 {
		return fmt.Errorf("invalid port: %d", AppConfig.Port)
	}

	return nil
}

func loadEnvConfig() {
	// 读取CHApi相关环境变量并覆盖配置
	AppConfig.CHApi.PasswordPepper = os.Getenv("CH_API_PASSWORD_PEPPER")
	AppConfig.CHApi.AllowedOrigin = os.Getenv("CH_API_ALLOWED_ORIGIN")
	AppConfig.CHApi.DatabaseURL = os.Getenv("CH_API_DATABASE_URL")
}
