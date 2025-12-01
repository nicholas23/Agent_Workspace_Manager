package handlers

import (
	"agent-workspace-manager/internal/services/realtime"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// StreamExecutionLogs 處理 SSE 連線，即時回傳執行日誌
func StreamExecutionLogs(c *gin.Context) {
	executionIDStr := c.Param("execution_id")
	executionID, err := strconv.ParseUint(executionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	// 訂閱日誌
	logChan := realtime.Broker.Subscribe(uint(executionID))

	// 設定 SSE Header
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 監聽 Client 斷線
	clientGone := c.Writer.CloseNotify()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case msg, ok := <-logChan:
			if !ok {
				// Channel closed (Execution finished)
				c.SSEvent("end", "Execution finished")
				return false
			}
			// 發送 log 事件
			c.SSEvent("log", msg)
			return true
		}
	})
}
