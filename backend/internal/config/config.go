package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config 結構體定義了應用程式的設定參數
type Config struct {
	Port             string // 伺服器埠口
	DatabaseURL      string // 資料庫連線字串
	TelegramBotToken string // Telegram Bot Token
	TelegramWhitelist string // Telegram 白名單 (逗號分隔)
}

// LoadConfig 從環境變數或 .env 檔案載入設定
func LoadConfig() *Config {
	// 嘗試載入 .env 檔案，如果不存在則忽略錯誤 (可能已設定環境變數)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "agent_workspace.db"),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramWhitelist: getEnv("TELEGRAM_WHITELIST", ""),
	}
}

// getEnv 取得環境變數，如果不存在則回傳預設值
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
