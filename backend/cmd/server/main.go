package main

import (
	"agent-workspace-manager/internal/api"
	"agent-workspace-manager/internal/api/middleware"
	"agent-workspace-manager/internal/config"
	"agent-workspace-manager/internal/database"
	"agent-workspace-manager/internal/logger"
	"agent-workspace-manager/internal/services/executor"
	"agent-workspace-manager/internal/services/realtime"
	"agent-workspace-manager/internal/services/scheduler"
	"agent-workspace-manager/internal/services/telegram"
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// main 是應用程式的進入點
func main() {
	// 載入設定
	cfg := config.LoadConfig()

	// 初始化 Loggers
	// 假設 logs 目錄在 backend/logs
	if err := logger.InitLoggers("logs", true); err != nil {
		log.Fatalf("Failed to init loggers: %v", err)
	}
	slog.Info("Loggers initialized")

	// 設定 Gin 的預設 Writer 為 Web Logger
	// 注意：Gin 的 Logger 中介軟體會寫入 DefaultWriter
	// 我們需要自定義一個 Writer 來適配 slog 或者直接寫入 Web Logger 的底層 Writer
	// 這裡簡單起見，我們假設 logger.Web 已經配置好 MultiWriter
	// 但 slog.Logger 沒有直接暴露 Writer。
	// 更好的方式是自定義 Gin Middleware 使用 slog，但為了符合需求，
	// 我們可以讓 logger.InitLoggers 返回 Writer 或者直接在 logger 包中暴露 Writer。
	// 由於 logger.Web 是 *slog.Logger，我們無法直接取出 Writer。
	// 暫時方案：Gin 預設寫入 stdout，我們可以在 middleware 中使用 slog 記錄請求。
	// 或者，修改 logger 包以暴露 Writer。
	// 為了簡單且符合 "分開 log" 的需求，我們在 logger 包中設定了 lumberjack。
	// 但 slog 封裝了 writer。
	// 讓我們修改 InitLoggers 讓它設定 Gin DefaultWriter? 不，這會造成循環依賴。
	// 我們可以在 main 中設定 Gin DefaultWriter，如果我們能從 logger 包獲取 writer。
	// 讓我們假設 logger.Web 是主要的 web logger。
	
	// 初始化 Realtime Broker
	realtime.InitBroker()

	// 初始化資料庫連線
	database.Connect(cfg.DatabaseURL)

	// 初始化 Telegram Bot (注入 Telegram Logger)
	telegram.InitBot(cfg, logger.Telegram)
	
	// 初始化 Executor Logger
	executor.SetLogger(logger.Executor)

	// 初始化排程器
	scheduler.InitScheduler()

	// 設定 Gin 的預設 Writer 為 Web Logger
	gin.DefaultWriter = logger.WebWriter

	// 設定 Gin 路由器
	r := gin.Default()
	// 套用 CORS 中介軟體
	r.Use(middleware.CORSMiddleware())

	// 設定 API 路由
	api.SetupRoutes(r)

	// 設定 HTTP Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// 啟動伺服器 (在 goroutine 中)
	go func() {
		logger.Web.Info("Server starting", "port", cfg.Port)
		// 發送啟動通知
		telegram.SendNotification("🚀 Agent Workspace Manager Server Started")
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Web.Error("Server failed to start", "error", err)
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中斷信號以優雅地關閉伺服器 (設定 5 秒超時)
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Web.Info("Shutting down server...")

	// 發送關閉通知
	telegram.SendNotification("🛑 Agent Workspace Manager Server Stopped")

	// 設定 Context 用於優雅關機的超時控制
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Web.Error("Server forced to shutdown", "error", err)
		log.Fatal("Server forced to shutdown:", err)
	}

	logger.Web.Info("Server exiting")
}