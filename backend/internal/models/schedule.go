package models

import (
	"time"

	"gorm.io/gorm"
)

// 定義排程狀態常數
const (
	SchedulePending   = "pending"   // 等待執行
	ScheduleCompleted = "completed" // 已執行
	ScheduleFailed    = "failed"    // 執行失敗
)

// Schedule 代表一個排程任務
type Schedule struct {
	gorm.Model
	// ProjectID 是關聯的專案 ID
	ProjectID uint `json:"project_id"`
	// Command 是要執行的指令
	Command string `json:"command"`
	// ScheduledTime 是預定執行時間
	ScheduledTime time.Time `json:"scheduled_time"`
	// Status 是排程狀態
	Status string `json:"status"`
}
