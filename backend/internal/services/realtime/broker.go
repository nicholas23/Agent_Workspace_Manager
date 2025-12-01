package realtime

import (
	"sync"
)

// LogBroker 管理即時日誌的訂閱與發布 (SSE Broker)
// 使用 map 儲存每個 Execution ID 對應的訂閱者 channel 列表。
type LogBroker struct {
	subscribers map[uint][]chan string // ExecutionID -> List of channels
	mu          sync.RWMutex           // 讀寫鎖，保護 subscribers map
}

// Broker 是全域的 LogBroker 實例
var Broker *LogBroker

// InitBroker 初始化即時通訊 Broker
//
// 功能:
//   - 建立全域的 LogBroker 實例。
//   - 初始化 subscribers map。
func InitBroker() {
	Broker = &LogBroker{
		subscribers: make(map[uint][]chan string),
	}
}

// Subscribe 訂閱特定 Execution 的日誌串流
//
// 參數:
//   - executionID: 要訂閱的執行記錄 ID。
//
// 返回:
//   - chan string: 用於接收日誌訊息的 Channel。
//
// 說明:
//   - 建立一個帶緩衝 (Buffered) 的 channel，避免發布者因接收者處理過慢而阻塞。
//   - 將 channel 加入到該 executionID 的訂閱列表中。
func (b *LogBroker) Subscribe(executionID uint) chan string {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan string, 100) // Buffered channel to prevent blocking
	b.subscribers[executionID] = append(b.subscribers[executionID], ch)
	return ch
}

// Unsubscribe 取消訂閱 (簡單實作：通常由 Client 斷線觸發，這裡不特別處理清理，依賴 Close)
// 在實際生產環境需要更嚴謹的 ID 管理

// Publish 發布日誌訊息給所有訂閱者
//
// 參數:
//   - executionID: 目標執行記錄 ID。
//   - message: 要發布的日誌內容。
//
// 邏輯:
//   - 使用讀鎖 (RLock) 讀取訂閱列表。
//   - 嘗試將訊息寫入每個訂閱者的 channel。
//   - 使用 select + default 機制：如果 channel 已滿 (阻塞)，則丟棄該訊息，
//     確保日誌系統不會拖慢核心執行流程。
func (b *LogBroker) Publish(executionID uint, message string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if chans, ok := b.subscribers[executionID]; ok {
		for _, ch := range chans {
			select {
			case ch <- message:
			default:
				// 如果 channel 滿了，丟棄訊息以避免阻塞執行器
			}
		}
	}
}

// CloseExecution 關閉指定 Execution 的所有訂閱
//
// 參數:
//   - executionID: 要關閉的執行記錄 ID。
//
// 功能:
//   - 當執行結束時呼叫此函數。
//   - 關閉該 ID 下的所有 channel，這會通知前端 SSE 連線結束。
//   - 從 map 中移除該 ID 的紀錄，釋放資源。
func (b *LogBroker) CloseExecution(executionID uint) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if chans, ok := b.subscribers[executionID]; ok {
		for _, ch := range chans {
			close(ch)
		}
		delete(b.subscribers, executionID)
	}
}
