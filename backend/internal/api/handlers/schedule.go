package handlers

import (
	"agent-workspace-manager/internal/database"
	"agent-workspace-manager/internal/models"
	"agent-workspace-manager/internal/services/scheduler"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateSchedule 建立新的排程任務
func CreateSchedule(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var input struct {
		Command       string    `json:"command" binding:"required"`
		ScheduledTime time.Time `json:"scheduled_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 檢查排程時間是否在未來
	if input.ScheduledTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Scheduled time must be in the future"})
		return
	}

	// 檢查是否已有等待中的排程 (每個專案只能有一個)
	var count int64
	database.DB.Model(&models.Schedule{}).Where("project_id = ? AND status = ?", projectID, models.SchedulePending).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Project already has a pending schedule"})
		return
	}

	schedule := models.Schedule{
		ProjectID:     uint(projectID),
		Command:       input.Command,
		ScheduledTime: input.ScheduledTime,
		Status:        models.SchedulePending,
	}

	if err := database.DB.Create(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule"})
		return
	}

	// 註冊到排程器
	scheduler.ScheduleJob(schedule)

	c.JSON(http.StatusCreated, schedule)
}

// GetSchedules 取得專案的排程列表
func GetSchedules(c *gin.Context) {
	projectID := c.Param("id")
	var schedules []models.Schedule
	// 依照預定時間倒序排列
	database.DB.Where("project_id = ?", projectID).Order("scheduled_time desc").Find(&schedules)
	c.JSON(http.StatusOK, schedules)
}

// GetAllSchedules 取得系統中所有等待中的排程
func GetAllSchedules(c *gin.Context) {
	var schedules []models.Schedule
	// Preload Project 資訊以便顯示
	database.DB.Preload("Project").Where("status = ?", models.SchedulePending).Order("scheduled_time asc").Find(&schedules)
	c.JSON(http.StatusOK, schedules)
}