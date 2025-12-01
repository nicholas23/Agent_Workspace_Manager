package models

import (
	"time"

	"gorm.io/gorm"
)

// 定義執行狀態常數
const (
	StatusRunning     = "running"     // 執行中
	StatusCompleted   = "completed"   // 已完成
	StatusFailed      = "failed"      // 失敗
	StatusParseFailed = "parse_failed" // 輸出解析失敗
)

// Execution 代表一次指令執行的記錄
type Execution struct {
	gorm.Model
	// ProjectID 是關聯的專案 ID
	ProjectID uint `json:"project_id"`
	// Command 是執行的具體指令內容
	Command string `json:"command"`
	// Status 是執行狀態
	Status string `json:"status"`
	// StartTime 是開始執行時間
	StartTime time.Time `json:"start_time"`
	// EndTime 是結束執行時間
	EndTime time.Time `json:"end_time"`
	// Summary 是執行的簡短摘要
	Summary string `json:"summary"`
	// Details 是執行的詳細輸出或日誌
	Details string `json:"details"`
	// ModifiedFiles 記錄被修改的檔案列表 (JSON 格式)
	ModifiedFiles []string `json:"modified_files" gorm:"serializer:json"`
	// CreatedFiles 記錄新建立的檔案列表 (JSON 格式)
	CreatedFiles []string `json:"created_files" gorm:"serializer:json"`
	// DeletedFiles 記錄被刪除的檔案列表 (JSON 格式)
	DeletedFiles []string `json:"deleted_files" gorm:"serializer:json"`
	// ErrorMessage 記錄錯誤訊息 (如果有)
	ErrorMessage string `json:"error_message"`
}
