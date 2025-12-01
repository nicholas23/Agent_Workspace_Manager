package handlers

import (
	"agent-workspace-manager/internal/config"
	"agent-workspace-manager/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSettings 取得所有設定
func GetSettings(c *gin.Context) {
	cfg := config.LoadConfig()
	
	settings := []models.Setting{
		{
			Key:         "TELEGRAM_BOT_TOKEN",
			Value:       cfg.TelegramBotToken,
			Description: "Telegram Bot Token (唯讀，請修改 .env 檔案)",
		},
		{
			Key:         "TELEGRAM_WHITELIST",
			Value:       cfg.TelegramWhitelist,
			Description: "Telegram 白名單 (唯讀，請修改 .env 檔案)",
		},
	}
	
	c.JSON(http.StatusOK, settings)
}

// UpdateSetting 更新單一設定 (已停用)
func UpdateSetting(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "Settings are read-only. Please update the .env file and restart the server."})
}
