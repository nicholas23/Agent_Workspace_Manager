package api

import (
	"agent-workspace-manager/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 設定所有 API 路由
func SetupRoutes(r *gin.Engine) {
	// 定義 API 路由群組
	api := r.Group("/api")
	{
		// 專案相關路由
		projects := api.Group("/projects")
		{
			projects.POST("", handlers.CreateProject)           // 建立專案
			projects.GET("", handlers.GetProjects)              // 取得專案列表
			projects.GET("/:id", handlers.GetProject)           // 取得單一專案
			projects.PUT("/:id", handlers.UpdateProject)        // 更新專案
			projects.DELETE("/:id", handlers.DeleteProject)     // 刪除專案
			projects.POST("/:id/run", handlers.RunProjectCommand) // 執行專案指令
			projects.GET("/:id/executions", handlers.GetProjectExecutions) // 取得專案執行記錄
			projects.POST("/:id/schedules", handlers.CreateSchedule) // 建立排程
			projects.GET("/:id/schedules", handlers.GetSchedules)    // 取得排程列表
		}

		// 執行記錄相關路由
		executions := api.Group("/executions")
		{
			executions.GET("/:execution_id", handlers.GetExecution) // 取得單一執行記錄
			// SSE 串流路由
			executions.GET("/:execution_id/stream", handlers.StreamExecutionLogs) 
		}

		// 系統設定相關路由
		settings := api.Group("/settings")
		{
			settings.GET("", handlers.GetSettings)       // 取得所有設定
			settings.PUT("/:key", handlers.UpdateSetting) // 更新設定
		}

		// 全域排程路由
		schedules := api.Group("/schedules")
		{
			schedules.GET("", handlers.GetAllSchedules) // 取得所有等待中的排程
		}
	}

	// 健康檢查端點
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}