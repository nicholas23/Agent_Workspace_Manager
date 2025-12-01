package models

import "gorm.io/gorm"

// Setting 代表系統設定
type Setting struct {
	gorm.Model
	// Key 是設定鍵名，必須唯一
	Key string `json:"key" gorm:"unique;not null"`
	// Value 是設定值
	Value string `json:"value"`
	// Description 是設定描述
	Description string `json:"description"`
}
