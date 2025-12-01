package scheduler

import (
	"agent-workspace-manager/internal/database"
	"agent-workspace-manager/internal/models"
	"agent-workspace-manager/internal/services/executor"
	"agent-workspace-manager/internal/services/telegram"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// Cron 是全域的排程器實例
var Cron *cron.Cron

// InitScheduler 初始化排程器服務
//
// 功能:
//  1. 建立並啟動一個支援秒級精度的 Cron 排程器。
//  2. 從資料庫載入所有狀態為 Pending 的排程任務。
//  3. 將這些任務重新加入排程系統，確保伺服器重啟後任務不丟失。
func InitScheduler() {
	Cron = cron.New(cron.WithSeconds())
	Cron.Start()

	// 載入等待中的排程
	var schedules []models.Schedule
	if err := database.DB.Where("status = ?", models.SchedulePending).Find(&schedules).Error; err != nil {
		log.Printf("Failed to load pending schedules: %v", err)
		return
	}

	for _, s := range schedules {
		ScheduleJob(s)
	}
}

// ScheduleJob 將單個任務加入排程
//
// 參數:
//   - s: 排程任務模型物件。
//
// 邏輯:
//  1. 檢查預定時間是否已過。
//  2. 若時間已過，立即執行任務。
//  3. 若時間未到，計算剩餘時間，使用 time.AfterFunc 設定定時器。
func ScheduleJob(s models.Schedule) {
	now := time.Now()
	if s.ScheduledTime.Before(now) {
		// 如果時間已過，立即執行
		log.Printf("Schedule %d is in the past, running immediately", s.ID)
		runJob(s.ID)
		return
	}

	duration := s.ScheduledTime.Sub(now)
	// 使用 time.AfterFunc 在指定時間後執行
	time.AfterFunc(duration, func() {
		runJob(s.ID)
	})
	
	log.Printf("Scheduled job %d for %v", s.ID, s.ScheduledTime)
}

// runJob 執行排程任務的實際邏輯
//
// 參數:
//   - scheduleID: 排程任務 ID。
//
// 流程:
//  1. 從資料庫查詢排程任務，確認其存在且狀態為 Pending。
//  2. 將排程狀態更新為 Completed (表示已觸發)。
//  3. 查詢關聯的專案資訊。
//  4. 呼叫 executor.ExecuteCommand 執行 AI 指令。
//  5. 設定回呼函式，在執行完成後透過 Telegram 發送通知。
func runJob(scheduleID uint) {
	var s models.Schedule
	if err := database.DB.First(&s, scheduleID).Error; err != nil {
		log.Printf("Schedule %d not found during execution", scheduleID)
		return
	}

	if s.Status != models.SchedulePending {
		return
	}

	// 更新狀態為已完成 (表示已觸發執行)
	s.Status = models.ScheduleCompleted
	database.DB.Save(&s)

	// 觸發執行
	var project models.Project
	if err := database.DB.First(&project, s.ProjectID).Error; err != nil {
		log.Printf("Project %d not found for schedule %d", s.ProjectID, s.ID)
		return
	}

	log.Printf("Executing scheduled job %d: %s", s.ID, s.Command)
	
	// 使用 executor 執行指令
	executor.ExecuteCommand(s.ProjectID, s.Command, func(execution *models.Execution) {
		msg := fmt.Sprintf("Scheduled Task Executed\nProject: %s\nStatus: %s\nSummary: %s", project.Name, execution.Status, execution.Summary)
		if execution.Status == models.StatusFailed || execution.Status == models.StatusParseFailed {
			msg += fmt.Sprintf("\nError: %s", execution.ErrorMessage)
		}
		telegram.SendNotification(msg)
	})
}
