package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	Environment       string
	DeepSeekAPIKey    string
	DeepSeekAPIURL    string
	DatabaseURL       string
	JWTSecret         string
	JWTExpiration     string
	AgentSystemPrompt string
}

var AppConfig *Config

// LoadConfig 加载配置文件
func LoadConfig() {
	// 加载.env文件
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	AppConfig = &Config{
		Port:              getEnv("PORT", "8080"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		DeepSeekAPIKey:    getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekAPIURL:    getEnv("DEEPSEEK_API_URL", "https://api.deepseek.com/v1/chat/completions"),
		DatabaseURL:       getEnv("DATABASE_URL", "sqlite://./zippifi.db"),
		JWTSecret:         getEnv("JWT_SECRET", "default_jwt_secret"),
		JWTExpiration:     getEnv("JWT_EXPIRATION", "24h"),
		AgentSystemPrompt: getEnv("AGENT_SYSTEM_PROMPT", "你是一个金融AI代理，负责分析市场数据并提供投资建议。"),
	}

	log.Println("Configuration loaded successfully")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
