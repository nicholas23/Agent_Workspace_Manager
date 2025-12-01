package database

import (
	"agent-workspace-manager/internal/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB 是全域的資料庫連線實例
var DB *gorm.DB

// Connect 建立資料庫連線並執行自動遷移
func Connect(databaseURL string) {
	var err error
	// 開啟 SQLite 連線
	DB, err = gorm.Open(sqlite.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	// 自動遷移資料庫結構 (建立表格)
	err = DB.AutoMigrate(
		&models.Project{},
		&models.Execution{},
		&models.Schedule{},
		&models.Setting{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")
}
