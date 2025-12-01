package telegram

import (
	"agent-workspace-manager/internal/config"
	"agent-workspace-manager/internal/database"
	"agent-workspace-manager/internal/models"
	"agent-workspace-manager/internal/services/executor"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot 是全域的 Telegram Bot 實例，用於發送訊息和接收更新。
var Bot *tgbotapi.BotAPI

// allowedUserIDs 儲存允許使用 Bot 的 Telegram 使用者 ID 列表 (白名單)。
var allowedUserIDs []int64

// Log 是 Telegram 服務專用的 Logger
var Log *slog.Logger

// InitBot 初始化 Telegram Bot 服務
//
// 參數:
//   - cfg: 系統配置物件，包含 Telegram Bot Token 和白名單設定。
//   - logger: 專用的 slog.Logger 實例。
//
// 功能:
//  1. 設定 Logger。
//  2. 檢查是否設定了 Bot Token，若無則跳過初始化。
//  3. 使用 Token 建立 Bot API 實例。
//  4. 解析並設定使用者白名單。
//  5. 啟動一個 goroutine 來監聽並處理 Telegram 更新 (訊息/指令)。
func InitBot(cfg *config.Config, logger *slog.Logger) {
	Log = logger
	if cfg.TelegramBotToken == "" {
		Log.Warn("Telegram Bot Token not set, skipping Telegram integration")
		return
	}

	var err error
	Bot, err = tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		Log.Error("Failed to init Telegram Bot", "error", err)
		return
	}

	Log.Info("Authorized on account", "username", Bot.Self.UserName)
	
	// Log masked token for debugging
	if len(cfg.TelegramBotToken) > 4 {
		maskedToken := cfg.TelegramBotToken[:4] + "..." + cfg.TelegramBotToken[len(cfg.TelegramBotToken)-4:]
		Log.Info("Telegram Bot Token loaded", "token", maskedToken)
	}

	// 解析白名單
	if cfg.TelegramWhitelist != "" {
		ids := strings.Split(cfg.TelegramWhitelist, ",")
		for _, idStr := range ids {
			if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil {
				allowedUserIDs = append(allowedUserIDs, id)
			}
		}
		Log.Info("Telegram whitelist configured", "count", len(allowedUserIDs))
	} else {
		Log.Warn("Telegram whitelist is empty")
	}

	// 啟動更新監聽迴圈
	go listenForUpdates()
}

// listenForUpdates 監聽並處理 Telegram 更新 (Long Polling)
//
// 功能:
//  1. 設定更新配置 (Timeout 為 60 秒)。
//  2. 透過 Channel 接收更新。
//  3. 驗證發送者是否在白名單中，若不在則拒絕存取。
//  4. 辨識並分派指令給對應的處理函數。
func listenForUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		Log.Info("Received message", "user_id", update.Message.From.ID, "text", update.Message.Text)

		// 檢查使用者是否在白名單中
		if !isUserAllowed(update.Message.From.ID) {
			Log.Warn("User not allowed", "user_id", update.Message.From.ID, "username", update.Message.From.UserName)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are not authorized to use this bot.")
			Bot.Send(msg)
			continue
		}

		// 處理指令
		if update.Message.IsCommand() {
			Log.Info("Handling command", "command", update.Message.Command(), "args", update.Message.CommandArguments())
			handleCommand(update.Message)
		} else {
			Log.Debug("Message is not a command", "text", update.Message.Text)
		}
	}
}

// isUserAllowed 檢查使用者 ID 是否在白名單中
//
// 參數:
//   - userID: Telegram 使用者 ID。
//
// 返回:
//   - bool: 如果使用者在白名單中則返回 true，否則返回 false。
func isUserAllowed(userID int64) bool {
	if len(allowedUserIDs) == 0 {
		Log.Warn("Telegram whitelist is empty. Denying access", "user_id", userID)
		return false
	}

	for _, id := range allowedUserIDs {
		if id == userID {
			return true
		}
	}
	Log.Warn("User ID not in whitelist", "user_id", userID)
	return false
}

// handleCommand 分派指令處理邏輯
//
// 支援的指令:
//   - /help: 顯示可用指令列表。
//   - /pp [page]: 列出專案列表 (分頁)。
//   - /run [project_name] [command]: 執行指定專案的 AI 指令。
//   - /status [project_name]: 查詢指定專案的最後一次執行狀態。
func handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "help":
		msg := tgbotapi.NewMessage(msg.Chat.ID, "Available commands:\n/pp [page] - List projects\n/run [project_name] [command] - Run command\n/status [project_name] - Check status")
		Bot.Send(msg)
	case "pp":
		handleListProjects(msg)
	case "run":
		handleRun(msg)
	case "status":
		handleStatus(msg)
	default:
		Log.Warn("Unknown command received", "command", msg.Command())
		msg := tgbotapi.NewMessage(msg.Chat.ID, "Unknown command")
		Bot.Send(msg)
	}
}

// handleListProjects 處理 /pp 指令：列出專案
//
// 參數:
//   - msg: Telegram 訊息物件，可能包含頁碼參數。
//
// 功能:
//   - 解析頁碼參數 (預設為第 1 頁)。
//   - 從資料庫分頁查詢專案列表。
//   - 格式化輸出專案名稱、ID 和描述。
func handleListProjects(msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())
	page := 1
	if len(args) > 0 {
		if p, err := strconv.Atoi(args[0]); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 10
	offset := (page - 1) * pageSize

	var projects []models.Project
	var total int64
	database.DB.Model(&models.Project{}).Count(&total)
	database.DB.Order("created_at desc").Limit(pageSize).Offset(offset).Find(&projects)

	if len(projects) == 0 {
		Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("No projects found on page %d.", page)))
		return
	}

	var response strings.Builder
	response.WriteString(fmt.Sprintf("Projects (Page %d/%d):\n", page, (total+int64(pageSize)-1)/int64(pageSize)))
	for _, p := range projects {
		response.WriteString(fmt.Sprintf("- %s (ID: %d)\n  %s\n", p.Name, p.ID, p.Description))
	}
	Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response.String()))
}

// handleRun 處理 /run 指令：執行專案指令
//
// 參數:
//   - msg: Telegram 訊息物件，必須包含 [project_name] 和 [command]。
//
// 功能:
//   - 驗證參數完整性。
//   - 根據專案名稱查詢專案 ID。
//   - 呼叫 executor 服務非同步執行指令。
//   - 執行完成後透過 SendNotification 發送結果通知。
func handleRun(msg *tgbotapi.Message) {
	args := strings.SplitN(msg.CommandArguments(), " ", 2)
	if len(args) < 2 {
		Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Usage: /run [project_name] [command]"))
		return
	}

	projectName := args[0]
	command := args[1]

	var project models.Project
	if err := database.DB.Where("name = ?", projectName).First(&project).Error; err != nil {
		Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Project not found"))
		return
	}

	// 非同步執行並在完成後通知
	go executor.ExecuteCommand(project.ID, command, func(execution *models.Execution) {
		msg := fmt.Sprintf("Project: %s\nStatus: %s\nSummary: %s", project.Name, execution.Status, execution.Summary)
		if execution.Status == models.StatusFailed || execution.Status == models.StatusParseFailed {
			msg += fmt.Sprintf("\nError: %s", execution.ErrorMessage)
		}
		SendNotification(msg)
	})
	Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Command execution started."))
}

// handleStatus 處理 /status 指令：查詢專案狀態
//
// 參數:
//   - msg: Telegram 訊息物件，必須包含 [project_name]。
//
// 功能:
//   - 根據專案名稱查詢專案。
//   - 查詢該專案最新的一筆執行記錄。
//   - 回傳執行狀態、開始時間與結束時間。
func handleStatus(msg *tgbotapi.Message) {
	projectName := msg.CommandArguments()
	if projectName == "" {
		Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Usage: /status [project_name]"))
		return
	}

	var project models.Project
	if err := database.DB.Where("name = ?", projectName).First(&project).Error; err != nil {
		Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Project not found"))
		return
	}

	var lastExecution models.Execution
	if err := database.DB.Where("project_id = ?", project.ID).Order("created_at desc").First(&lastExecution).Error; err != nil {
		Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "No executions found"))
		return
	}

	response := fmt.Sprintf("Last Execution Status: %s\nStart Time: %s", lastExecution.Status, lastExecution.StartTime.Format(time.RFC3339))
	if !lastExecution.EndTime.IsZero() {
		response += fmt.Sprintf("\nEnd Time: %s", lastExecution.EndTime.Format(time.RFC3339))
	}
	Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
}

// SendNotification 發送通知給所有白名單使用者
//
// 參數:
//   - message: 要發送的訊息內容。
//
// 功能:
//   - 遍歷 allowedUserIDs 列表，向每個使用者發送訊息。
//   - 用於系統通知 (如伺服器啟動/關閉) 或執行結果通知。
func SendNotification(message string) {
	if Bot == nil {
		return
	}

	for _, chatID := range allowedUserIDs {
		Bot.Send(tgbotapi.NewMessage(chatID, message))
	}
}
