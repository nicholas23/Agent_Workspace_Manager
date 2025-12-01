package handlers

import (
	"agent-workspace-manager/internal/database"
	"agent-workspace-manager/internal/models"
	"agent-workspace-manager/internal/services/executor"
	"agent-workspace-manager/internal/services/telegram"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RunProjectCommand 執行專案指令
func RunProjectCommand(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var input struct {
		Command string `json:"command" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 非同步執行指令
	go executor.ExecuteCommand(uint(projectID), input.Command, func(execution *models.Execution) {
		// 執行完成後發送 Telegram 通知
		var project models.Project
		if err := database.DB.First(&project, execution.ProjectID).Error; err == nil {
			msg := fmt.Sprintf("Project: %s\nStatus: %s\nSummary: %s", project.Name, execution.Status, execution.Summary)
			if execution.Status == models.StatusFailed || execution.Status == models.StatusParseFailed {
				msg += fmt.Sprintf("\nError: %s", execution.ErrorMessage)
			}
			telegram.SendNotification(msg)
		}
	})

	c.JSON(http.StatusAccepted, gin.H{"message": "Command execution started"})
}

// GetExecution 取得單一執行記錄詳情
func GetExecution(c *gin.Context) {
	executionID := c.Param("execution_id")
	var execution models.Execution
	if err := database.DB.First(&execution, executionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}
	c.JSON(http.StatusOK, execution)
}
