package models

import "gorm.io/gorm"

// Project 代表一個 AI Agent 專案
type Project struct {
	gorm.Model
	// Name 是專案名稱，必須唯一
	Name string `json:"name" gorm:"unique;not null"`
	// Description 是專案描述
	Description string `json:"description"`
	// AICliCommand 是用於執行此專案的 AI CLI 指令模板
	AICliCommand string `json:"ai_cli_command"`
	// DirectoryPath 是專案在檔案系統中的絕對路徑
	DirectoryPath string `json:"directory_path" gorm:"not null"`
	// Executions 關聯到該專案的所有執行記錄
	Executions []Execution `json:"executions,omitempty" gorm:"foreignKey:ProjectID"`
}
