package executor

import (
	"agent-workspace-manager/internal/database"
	"agent-workspace-manager/internal/models"
	"agent-workspace-manager/internal/services/realtime"
	"agent-workspace-manager/internal/utils"
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// CompletionCallback 定義執行完成後的回呼函式型別
// 參數:
//   - execution: 指向已完成的 Execution 模型的指標
type CompletionCallback func(*models.Execution)

// executionLocks 用於控制每個專案的並行執行
// 鍵為專案 ID，值為該專案的互斥鎖
var executionLocks = make(map[uint]*sync.Mutex)
var mapLock sync.Mutex

// DefaultTimeout 預設執行超時時間 (30分鐘)
const DefaultTimeout = 30 * time.Minute

// Log 是 Executor 服務專用的 Logger
var Log *slog.Logger

// SetLogger 設定 Executor 服務的 Logger
func SetLogger(logger *slog.Logger) {
	Log = logger
}

// getProjectLock 取得專案的執行鎖
//
// 參數:
//   - projectID: 專案 ID
//
// 返回:
//   - *sync.Mutex: 該專案的互斥鎖
//
// 說明:
//   使用 mapLock 確保並發安全地存取 executionLocks map。
//   如果該專案的鎖不存在，則創建一個新的。
func getProjectLock(projectID uint) *sync.Mutex {
	mapLock.Lock()
	defer mapLock.Unlock()

	if _, exists := executionLocks[projectID]; !exists {
		executionLocks[projectID] = &sync.Mutex{}
	}
	return executionLocks[projectID]
}

// streamWriter 是一個自定義 Writer，用於將輸出同時寫入 Broker 和底層 Writer
type streamWriter struct {
	executionID uint
	writer      io.Writer // 用於收集完整日誌 (例如 bytes.Buffer)
}

// Write 實作 io.Writer 介面
//
// 功能:
//  1. 將數據發送到即時系統 (Realtime Broker) 以供前端串流顯示。
//  2. 將數據寫入底層 writer 以供後續儲存。
func (sw *streamWriter) Write(p []byte) (n int, err error) {
	// 發送到即時系統
	if realtime.Broker != nil {
		realtime.Broker.Publish(sw.executionID, string(p))
	}
	// 寫入底層 writer
	return sw.writer.Write(p)
}

// ExecuteCommand 執行 AI Agent 指令的核心邏輯
//
// 參數:
//   - projectID: 目標專案 ID。
//   - userCommand: 使用者輸入的指令或提示詞。
//   - onComplete: 執行完成後的回呼函式 (可選)。
//
// 流程:
//  1. 並行控制: 嘗試取得專案鎖，若失敗則返回 "Project is busy"。
//  2. 準備資料: 獲取專案資訊、建立執行記錄 (Running 狀態)、獲取歷史記錄。
//  3. 建構指令: 組合 Prompt、解析 CLI 模版、替換參數。
//  4. 執行環境: 設定 Context (Timeout)、工作目錄。
//  5. 執行程序: 啟動外部指令，並透過 Pipe 即時讀取 Stdout/Stderr。
//  6. 串流輸出: 將輸出即時推送到 Realtime Broker，同時收集完整日誌。
//  7. 結果處理: 等待指令結束，解析輸出 (JSON)，更新執行記錄狀態 (Completed/Failed)。
//  8. 收尾: 釋放鎖，呼叫 onComplete。
func ExecuteCommand(projectID uint, userCommand string, onComplete CompletionCallback) {
	// 確保 Logger 已初始化
	if Log == nil {
		Log = slog.Default()
	}

	// 0. 並行控制：嘗試取得鎖
	lock := getProjectLock(projectID)
	if !lock.TryLock() {
		Log.Warn("Project is busy", "project_id", projectID)
		execution := models.Execution{
			ProjectID:    projectID,
			Command:      userCommand,
			Status:       models.StatusFailed,
			ErrorMessage: "Project is busy (concurrency limit)",
			StartTime:    time.Now(),
			EndTime:      time.Now(),
		}
		database.DB.Create(&execution)
		if onComplete != nil {
			onComplete(&execution)
		}
		return
	}
	defer lock.Unlock()

	// 1. 取得專案資訊
	var project models.Project
	if err := database.DB.First(&project, projectID).Error; err != nil {
		Log.Error("Project not found", "project_id", projectID)
		return
	}

	// 2. 建立執行記錄
	execution := models.Execution{
		ProjectID: projectID,
		Command:   userCommand,
		Status:    models.StatusRunning,
		StartTime: time.Now(),
	}
	database.DB.Create(&execution)

	// 2.5 取得最近 5 筆執行記錄 (作為 Context)
	var history []models.Execution
	database.DB.Where("project_id = ? AND id != ? AND status = ?", projectID, execution.ID, models.StatusCompleted).
		Order("created_at desc").
		Limit(5).
		Find(&history)

	// 3. 建構完整指令內容
	promptContent := utils.BuildPrompt(userCommand, history, project)
	
	// 解析 CLI 指令模版
	templateParts := strings.Fields(project.AICliCommand)
	if len(templateParts) == 0 {
		execution.Status = models.StatusFailed
		execution.ErrorMessage = "Empty AI CLI command configuration"
		execution.EndTime = time.Now()
		database.DB.Save(&execution)
		if onComplete != nil {
			onComplete(&execution)
		}
		return
	}
	// 取得指令執行檔
	exe := templateParts[0]
	var args []string
	// 處理參數
	for _, part := range templateParts[1:] {
		args = append(args, part)
	}
	// 加入提示詞作為最後一個參數
	args = append(args, promptContent)

	// 4. 準備執行 Context (Timeout)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	Log.Debug("Command", "exe", exe, "args", args)
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir = project.DirectoryPath

	// 設定輸出串流
	// 使用 Pipe 讀取 Stdout/Stderr，因為我們需要即時串流，而不僅僅是最後收集
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		finalizeExecution(&execution, models.StatusFailed, fmt.Sprintf("Failed to create stdout pipe: %v", err), "", onComplete)
		return
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		finalizeExecution(&execution, models.StatusFailed, fmt.Sprintf("Failed to create stderr pipe: %v", err), "", onComplete)
		return
	}

	// 啟動 Goroutine 讀取輸出並廣播
	var outputBuilder strings.Builder
	var wg sync.WaitGroup
	wg.Add(2)

	// 定義讀取並廣播的輔助函式
	readAndBroadcast := func(r io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			text := scanner.Text()
			// 廣播到前端
			if realtime.Broker != nil {
				realtime.Broker.Publish(execution.ID, text)
			}
			// 收集到 Buffer
			outputBuilder.WriteString(text + "\n")
		}
	}

	go readAndBroadcast(stdoutPipe)
	go readAndBroadcast(stderrPipe)

	// 5. 啟動指令
	Log.Info("Starting execution", "execution_id", execution.ID, "project_id", projectID, "command", exe)
	if err := cmd.Start(); err != nil {
		finalizeExecution(&execution, models.StatusFailed, fmt.Sprintf("Failed to start command: %v", err), "", onComplete)
		return
	}

	// 等待指令完成
	err = cmd.Wait()
	wg.Wait() // 確保所有輸出都已讀取完畢

	// 關閉 Broker (通知前端串流結束)
	if realtime.Broker != nil {
		realtime.Broker.CloseExecution(execution.ID)
	}

	fullOutput := outputBuilder.String()
	execution.Details = fullOutput
	execution.EndTime = time.Now()

	// 檢查 Timeout
	if ctx.Err() == context.DeadlineExceeded {
		finalizeExecution(&execution, models.StatusFailed, "Execution timed out", fullOutput, onComplete)
		return
	}

	if err != nil {
		finalizeExecution(&execution, models.StatusFailed, err.Error(), fullOutput, onComplete)
		return
	}
	Log.Debug("Command output", "execution_id", execution.ID, "output", fullOutput)
	
	// 6. 解析輸出 (嘗試從輸出中提取 JSON 結果)
	parsedOutput, err := utils.ParseOutput(fullOutput)
	if err != nil {
		// 即使解析失敗，也視為完成，但在狀態上標記為 ParseFailed
		execution.Status = models.StatusParseFailed
		execution.ErrorMessage = fmt.Sprintf("Output parsing failed: %v", err)
	} else {
		execution.Status = models.StatusCompleted
		execution.Summary = parsedOutput.Summary
		execution.ModifiedFiles = parsedOutput.ModifiedFiles
		execution.CreatedFiles = parsedOutput.CreatedFiles
		execution.DeletedFiles = parsedOutput.DeletedFiles
	}

	database.DB.Save(&execution)
	Log.Info("Execution completed", "execution_id", execution.ID, "status", execution.Status)

	if onComplete != nil {
		onComplete(&execution)
	}
}

// finalizeExecution 輔助函式：統一處理執行失敗或異常結束的狀態更新
//
// 參數:
//   - execution: 執行記錄物件。
//   - status: 最終狀態。
//   - errorMsg: 錯誤訊息。
//   - details: 執行詳細輸出 (Log)。
//   - onComplete: 回呼函式。
func finalizeExecution(execution *models.Execution, status, errorMsg, details string, onComplete CompletionCallback) {
	execution.Status = status
	execution.ErrorMessage = errorMsg
	if details != "" {
		execution.Details = details
	}
	execution.EndTime = time.Now()
	database.DB.Save(execution)
	
	Log.Error("Execution failed", "execution_id", execution.ID, "error", errorMsg)

	if realtime.Broker != nil {
		realtime.Broker.Publish(execution.ID, fmt.Sprintf("Error: %s", errorMsg))
		realtime.Broker.CloseExecution(execution.ID)
	}

	if onComplete != nil {
		onComplete(execution)
	}
}
