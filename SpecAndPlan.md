# AI Agent 管理系統 - Project Specification

## 1. 專案概述

### 1.1 核心目標
本系統是一個透過 Telegram 遠端管理多個 AI Agent 專案的系統。使用者可以透過 Telegram 查看每個 Project 的執行狀況,並下達指令讓 AI Agent 執行任務,無需一直在電腦前操作。

### 1.2 主要特性
- 同時管理多個 Project/目錄
- 每個 Project 指定專屬的 AI Agent
- AI Agent 對目錄內檔案進行處理(程式碼修改、資料分析、文件處理等)
- 透過 Telegram 下達指令與監控執行狀態
- 執行完成後透過 Telegram 回傳結果
- 提供 Web 介面進行系統管理與歷史查詢

---

## 2. Project 管理

### 2.1 Project 基本資訊
每個 Project 包含以下資訊:
- **Project 名稱**: 僅允許英文、數字、底線
- **簡短描述**: 描述該 Project 的用途
- **AI Agent CLI 命令**: 指定使用的 AI CLI 工具及參數格式(如 `gemini {prompt}`)
- **目錄路徑**: 絕對路徑,系統會驗證目錄是否存在
- **建立日期**: 自動記錄

### 2.2 AI Agent 配置
- 用戶必須指定每個 Project 使用哪個 AI Agent
- AI Agent 透過 CLI 工具執行(如 `gemini`, `aider` 等)
- 用戶需預先在系統環境設定好 API Keys 等環境變數
- System prompt 由各 AI Agent CLI 自行處理,系統不介入

### 2.3 工作範圍限制
- AI Agent 的工作目錄為該 Project 目錄
- 只能存取該目錄內的檔案
- 不能存取目錄外的檔案

---

## 3. Telegram 互動介面

### 3.1 Bot 設定
- 用戶需自行建立 Telegram Bot 並提供 Bot Token
- 用戶需將 Bot 加為好友
- 系統可設定白名單(僅處理特定 Telegram User ID 的訊息)
- Bot 分享給他人的管控不在系統範圍內

### 3.2 指令列表

#### 3.2.1 查看 Project 列表
```
/pp 1  # 顯示第 1-10 筆 Project
/pp 2  # 顯示第 11-20 筆 Project
```
顯示內容:
- Project 名稱
- 簡短描述
- 建立日期
- 按建立日期排序

#### 3.2.2 下達執行指令
```
/run project_name 要執行的動作
```
範例: `/run project_name_1 分析最新的 CSV 檔案`

#### 3.2.3 查詢執行狀態
```
/status project_name
```
顯示內容:
- 執行狀態(等待中/執行中/已完成/失敗/未收到結果)
- 執行時間

#### 3.2.4 查看系統說明
```
help
```
顯示所有可用指令說明

### 3.3 執行結果通知
當任務完成後,系統自動推送訊息:
```
project_name AI Agent 的 50 字簡述
```

### 3.4 錯誤處理
- 執行失敗時透過 Telegram 即時通知錯誤訊息
- 同一 Project 執行中時收到新指令 → 回覆「命令執行中」

---

## 4. 執行流程與狀態管理

### 4.1 指令執行流程
1. 用戶透過 Telegram 或 Web 介面下達指令
2. 系統組合完整 prompt:
   - JSON 回傳格式規範說明
   - Project ID/名稱
   - 用戶的實際指令
3. 系統切換到 Project 目錄
4. 執行 AI CLI 命令
5. 系統監控 CLI 程序執行狀態並捕捉 STDOUT
6. CLI 執行完成後,從 STDOUT 解析 JSON 結果
7. 驗證並儲存執行結果
8. 透過 Telegram 通知用戶

### 4.2 執行狀態
- **等待中 (Pending)**: 任務已建立,等待執行
- **執行中 (Running)**: AI Agent 正在執行
- **已完成 (Completed)**: 執行成功完成
- **失敗 (Failed)**: 執行過程發生錯誤
- **解析失敗 (Parse Failed)**: CLI 執行完成但 JSON 解析失敗

### 4.3 並行控制
- 同一 Project 同時只能執行一個指令
- 執行中收到新指令 → 直接回覆「命令執行中」並拒絕

### 4.4 超時處理
- 不設定超時限制
- 讓 AI Agent 執行到完成

### 4.5 STDOUT 結果解析
- CLI 執行完成後,從 STDOUT 讀取完整輸出
- 解析輸出中的 JSON 格式結果
- 如果解析失敗,標記為「解析失敗」狀態

---

## 5. AI Agent 輸出格式規範

### 5.1 JSON 回傳格式
AI Agent 必須在執行完成後,透過 STDOUT 輸出以下 JSON 格式:

```json
{
  "status": "success",
  "summary": "分析了 3 個 CSV 檔案,生成統計報表並儲存至 report.md",
  "details": "完整的執行過程描述...",
  "modified_files": ["data/report.md", "config.json"],
  "created_files": ["output/summary.txt"],
  "deleted_files": ["temp/old_data.csv"]
}
```

**欄位說明:**
- `status` (必填): 執行狀態,值為 `"success"` 或 `"failed"`
- `summary` (必填): 50 字以內的簡述
- `details` (必填): 完整的執行內容描述
- `modified_files` (選填): 修改的檔案列表,使用相對路徑
- `created_files` (選填): 新增的檔案列表,使用相對路徑
- `deleted_files` (選填): 刪除的檔案列表,使用相對路徑

### 5.2 Prompt 範本
系統會在用戶指令前加入以下 prompt:

```
請執行以下任務,並在完成後以 JSON 格式輸出結果。

【重要】JSON 格式規範:
{
  "status": "success 或 failed",
  "summary": "50字以內的執行摘要",
  "details": "完整執行過程",
  "modified_files": ["相對路徑1", "相對路徑2"],
  "created_files": ["相對路徑3"],
  "deleted_files": ["相對路徑4"]
}

【專案資訊】
- Project ID: {project_name}
- 工作目錄: {project_directory}

【任務內容】
{user_command}
```

### 5.3 解析規則
- 系統會從 STDOUT 中尋找 JSON 格式內容
- 支援 JSON 前後有其他文字(例如 AI 的說明文字)
- 使用正則表達式或 JSON 邊界偵測提取 JSON 區塊
- 如果找到多個 JSON 區塊,取最後一個
- 解析失敗則標記為「解析失敗」狀態,並保留原始 STDOUT 內容

### 5.4 錯誤處理
如果 AI Agent 無法按格式回傳:
- CLI 執行成功但 JSON 解析失敗 → 標記為「解析失敗」
- 保留完整 STDOUT 內容供用戶在 Web 介面查看
- Telegram 通知:「執行完成但結果格式異常,請至 Web 介面查看」

---

## 6. Web 管理介面

### 6.1 認證方式
- 本地系統,不需要登入認證

### 6.2 功能列表

#### 6.2.1 Project 管理
- 建立 Project
  - 輸入名稱、描述、AI CLI 命令、目錄路徑
- 編輯 Project 資訊
- 刪除 Project
- 查看所有 Project 列表

#### 6.2.2 執行管理
- 手動下達執行指令
- 查看執行歷史
  - 按狀態篩選
  - 按日期範圍篩選
  - 點擊單筆記錄查看完整詳情

#### 6.2.3 排程管理
- 設定一次性排程任務
  - 選擇 Project
  - 使用日期時間選擇器設定執行時間
  - 輸入執行指令
- 每個 Project 只能設定一個排程
- 執行後標記完成,不重複執行

#### 6.2.4 系統設定
- 管理 Telegram Bot Token
- 管理 Telegram 白名單 User IDs

---

## 7. 資料儲存

### 7.1 資料庫結構

#### 7.1.1 Projects 表
- id
- name (唯一)
- description
- ai_cli_command
- directory_path
- created_at

#### 7.1.2 Executions 表
- id
- project_id (外鍵)
- command
- status
- start_time
- end_time
- summary (50字簡述)
- details (完整執行內容)
- modified_files (JSON)
- created_files (JSON)
- deleted_files (JSON)
- error_message

#### 7.1.3 Schedules 表
- id
- project_id (外鍵)
- scheduled_time
- command
- status (pending/completed)

#### 7.1.4 Settings 表
- key
- value
- 儲存項目:
  - telegram_bot_token
  - telegram_whitelist (JSON 陣列)

### 7.2 執行記錄
- 所有執行記錄完整儲存在資料庫
- Telegram 查詢只顯示最後一次執行結果
- Web 介面可查看完整歷史記錄

---

## 8. 錯誤處理與日誌

### 8.1 錯誤類型
- 檔案不存在
- AI API 呼叫失敗
- CLI 工具執行錯誤
- 權限問題
- MCP 未收到回傳

### 8.2 錯誤處理機制
1. 透過 Telegram 即時通知用戶
2. 記錄在執行日誌中
3. 標記為「失敗」或「未收到結果」狀態
4. Web 介面可查看詳細錯誤訊息

---

## 9. User Stories

### 9.1 建立與管理 Project
**As a** 用戶  
**I want to** 在 Web 介面建立新的 Project  
**So that** 我可以讓 AI Agent 管理特定目錄的任務

**Acceptance Criteria:**
- 可以輸入 Project 名稱(英文、數字、底線)
- 可以輸入簡短描述
- 可以設定 AI CLI 命令格式
- 可以輸入絕對路徑,系統驗證目錄存在
- 建立成功後顯示在 Project 列表

---

### 9.2 透過 Telegram 查看 Project 列表
**As a** 用戶  
**I want to** 在 Telegram 輸入 `/pp 1`, `/pp 2` 查看 Project  
**So that** 我可以快速瀏覽所有專案

**Acceptance Criteria:**
- `/pp 1` 顯示第 1-10 筆
- `/pp 2` 顯示第 11-20 筆
- 顯示內容包含名稱、描述、建立日期
- 按建立日期排序

---

### 9.3 透過 Telegram 下達執行指令
**As a** 用戶  
**I want to** 在 Telegram 輸入 `/run project_name 執行指令`  
**So that** 我可以遠端觸發 AI Agent 執行任務

**Acceptance Criteria:**
- 系統識別 Project 名稱
- 組合完整 prompt(含 MCP 說明)
- 切換到 Project 目錄執行 CLI
- 如果該 Project 正在執行中,回覆「命令執行中」
- 開始執行後回覆確認訊息

---

### 9.4 查詢執行狀態
**As a** 用戶  
**I want to** 在 Telegram 輸入 `/status project_name`  
**So that** 我可以知道任務目前的執行狀況

**Acceptance Criteria:**
- 顯示執行狀態(等待中/執行中/已完成/失敗/解析失敗)
- 顯示執行開始時間
- 如果已完成,顯示結束時間

---

### 9.5 接收執行結果通知
**As a** 用戶  
**I want to** 任務完成後自動收到 Telegram 通知  
**So that** 我不需要持續查詢就能知道結果

**Acceptance Criteria:**
- 任務完成後自動推送訊息
- 格式: `project_name [50字簡述]`
- 成功與失敗都會通知
- 失敗時包含錯誤訊息

---

### 9.6 設定排程任務
**As a** 用戶  
**I want to** 在 Web 介面設定排程任務  
**So that** 我可以在特定時間自動執行指令

**Acceptance Criteria:**
- 選擇 Project
- 使用日期時間選擇器設定執行時間
- 輸入要執行的指令
- 每個 Project 只能設定一個排程
- 時間到後自動執行
- 執行後標記完成,不重複

---

### 9.7 查看執行歷史
**As a** 用戶  
**I want to** 在 Web 介面查看執行歷史  
**So that** 我可以在 Web 介面查看執行歷史
**So that** 我可以追蹤所有任務的執行記錄

**Acceptance Criteria:**
- 顯示該 Project 所有執行記錄
- 可按狀態篩選
- 可按日期範圍篩選
- 點擊單筆記錄查看完整詳情(修改的檔案、完整輸出等)

---

### 9.8 管理 Telegram Bot 設定
**As a** 用戶  
**I want to** 在 Web 介面設定 Telegram Bot Token 和白名單  
**So that** 系統可以接收我的 Telegram 訊息

**Acceptance Criteria:**
- 可以輸入/修改 Bot Token
- 可以新增/刪除白名單 User ID
- 儲存後立即生效
- 非白名單訊息不處理

---

### 9.9 AI Agent 透過 STDOUT 回傳結果
**As an** AI Agent  
**I want to** 透過 STDOUT 以 JSON 格式輸出執行結果  
**So that** 系統可以解析、記錄並通知用戶

**Acceptance Criteria:**
- 使用指定的 JSON 格式輸出
- 包含 status, summary, details
- 包含修改/新增/刪除的檔案列表(相對路徑)
- 系統成功解析 JSON 並儲存到資料庫
- 透過 Telegram 通知用戶
- 如果 JSON 格式錯誤,標記為「解析失敗」

---

### 9.10 處理執行錯誤
**As a** 用戶  
**I want to** 當執行失敗時收到詳細錯誤訊息  
**So that** 我可以了解問題並修正

**Acceptance Criteria:**
- Telegram 即時通知錯誤
- 錯誤記錄在資料庫
- 標記為「失敗」狀態
- Web 介面可查看詳細錯誤內容
- 如果 CLI 完成但 JSON 解析失敗,標記為「解析失敗」並保留原始輸出

---

## 10. 技術需求

### 10.1 環境需求
- 本地執行環境
- 需安裝各 AI CLI 工具
- 需設定好各 AI 的環境變數(API Keys)
- **Go 1.23+**
- Node.js 16+ (用於 Vue 前端開發)

### 10.2 技術選型

#### 10.2.1 後端技術
- **語言**: Go 1.23+
- **專案管理**: Go Modules (`go mod`)
- **Web 框架**: Gin (`github.com/gin-gonic/gin`)
  - 高效能、輕量級的 Web 框架
  - **Server-Sent Events (SSE)**: 用於即時推播執行 Log 與狀態更新 (取代輪詢)
- **MCP Server**: Gin Handler
- **背景任務 & 排程**: Goroutines + `robfig/cron` (v3)
  - 處理排程任務
  - 監控執行中的 CLI 程序
- **Telegram 整合**: `go-telegram-bot-api/telegram-bot-api` (v5)
- **資料庫 ORM**: GORM (`gorm.io/gorm`)
- **資料庫驅動**: `gorm.io/driver/sqlite`
- **資料庫遷移**: `golang-migrate` 或 GORM AutoMigrate
- **程序管理**: `os/exec` + Goroutines
- **日誌系統**:
  - **File Logging**: 使用 `log/slog` (Go 1.21+) 寫入 `backend/logs/`
  - **Debug Mode**: 動態調整 Log Level
- **安全機制**:
  - **Path Validation**: 使用 `filepath.Clean` 與 `strings.HasPrefix` 驗證路徑，防止 Directory Traversal
  - **Command Filtering**: 實作指令黑名單
  - **Resource Limits**: Context Timeout 控制執行時間
  - **Strict Environment Isolation**: 白名單機制過濾環境變數
  - **Context Awareness**: 自動注入最近 5 筆執行紀錄
  - **Prompt Security**: 系統層級 Prompt 注入安全規範

#### 10.2.2 資料庫
- **SQLite**
  - 輕量級,無需額外安裝
  - 適合本地單用戶使用
  - 檔案: `app.db`

#### 10.2.3 前端技術
- **框架**: Vue 3 (Composition API)
- **UI 框架**: Element Plus 或 Naive UI
- **狀態管理**: Pinia
- **HTTP 客戶端**: Axios
- **路由**: Vue Router
- **建構工具**: Vite

#### 10.2.4 開發工具
- **程式碼格式化**: `gofmt`, `goimports`, Prettier (Vue)
- **Linting**: `golangci-lint`, ESLint (Vue)
- **版本控制**: Git
- **模組管理**: Go Modules

### 10.3 外部相依
- Telegram Bot API
- 各 AI 廠商的 CLI 工具

### 10.4 專案結構
```
ai-agent-manager/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go             # 應用程式入口
│   ├── internal/
│   │   ├── api/                    # API 層
│   │   │   ├── handlers/           # HTTP Handlers (Gin)
│   │   │   │   ├── project.go
│   │   │   │   ├── execution.go
│   │   │   │   ├── schedule.go
│   │   │   │   └── setting.go
│   │   │   ├── middleware/         # Middleware (Auth, CORS, etc.)
│   │   │   └── routes.go           # 路由定義
│   │   ├── config/                 # 配置管理 (Viper or env)
│   │   ├── database/               # 資料庫連接與遷移
│   │   ├── models/                 # GORM Models
│   │   │   ├── project.go
│   │   │   ├── execution.go
│   │   │   ├── schedule.go
│   │   │   └── setting.go
│   │   ├── services/               # 業務邏輯
│   │   │   ├── executor/           # CLI 執行引擎
│   │   │   ├── telegram/           # Telegram Bot 服務
│   │   │   └── scheduler/          # 排程服務
│   │   └── utils/
│   │       ├── prompt_builder.go   # Prompt 組合
│   │       ├── output_parser.go    # STDOUT JSON 解析
│   │       └── validators.go       # 驗證工具
│   ├── pkg/                        # 可共用套件 (Optional)
│   ├── migrations/                 # SQL Migration 檔案
│   ├── go.mod
│   ├── go.sum
│   └── .env.example
├── frontend/
│   ├── src/
│   │   ├── main.js
│   │   ├── App.vue
│   │   ├── router/
│   │   ├── stores/                 # Pinia stores
│   │   ├── views/                  # 頁面組件
│   │   │   ├── ProjectList.vue
│   │   │   ├── ProjectDetail.vue
│   │   │   ├── ExecutionHistory.vue
│   │   │   ├── ScheduleManager.vue
│   │   │   └── Settings.vue
│   │   ├── components/             # 可複用組件
│   │   └── api/                    # API 客戶端
│   ├── package.json
│   └── vite.config.js
├── data/
│   └── app.db                      # SQLite 資料庫
├── logs/                           # 執行日誌
├── README.md
└── docker-compose.yml              # (可選) Docker 部署
```
---

## 11. 實作計劃

### 11.1 Phase 1: 核心功能

#### 1.1 環境建置
- 建立專案目錄結構
- 初始化 Go Module (`go mod init`)
- 安裝後端相依套件 (`go get ...`)
- 設定 Gin 專案結構
- 設定 SQLite 資料庫連接 (GORM)

#### 1.2 資料庫設計與實作
- 定義 GORM Models (Structs)
  - Project
  - Execution
  - Schedule
  - Setting
- 設定 AutoMigrate 或 `golang-migrate`
- 測試資料庫 CRUD 操作

#### 1.3 Project 管理 API
- 定義 Request/Response Structs
- 實作 Project CRUD Handlers
  - POST /api/projects
  - GET /api/projects
  - GET /api/projects/:id
  - PUT /api/projects/:id
  - DELETE /api/projects/:id
- 實作目錄路徑驗證
- 撰寫 API 測試 (`net/http/httptest`)

#### 1.4 執行引擎核心
- 實作 Executor Service
  - 使用 `os/exec` 執行 CLI
  - 使用 Goroutines 處理並行與背景執行
  - 捕捉 STDOUT/STDERR
  - 錯誤處理
- 實作 Prompt Builder
  - JSON 格式規範說明生成
  - Project 識別資訊注入
- 實作 Output Parser
  - 解析 STDOUT JSON
  - JSON Unmarshal 到 Struct
- 實作執行狀態管理
  - 使用 Mutex 或 Channel 控制並行
- 測試 CLI 執行流程

#### 1.5 結果解析與儲存
- 設計 JSON Struct tags
- 實作解析邏輯
  - 提取 JSON 字串
  - 處理多區塊
- 實作執行結果儲存
  - 使用 GORM 儲存 Execution
  - 處理檔案列表 (JSON 序列化儲存)
- 測試各種 STDOUT 格式情境

#### 1.6 進階功能 (Advanced Features)
- **即時更新 (SSE)**
  - 實作 Gin SSE Handler
  - 使用 Channel 廣播 Log 事件
- **檔案日誌 (File Logging)**
  - 設定 `slog` 輸出至 `backend/logs/`
- **除錯模式 (Debug Mode)**
  - 動態調整 Log Level
  - 關鍵路徑加入 Debug Log

**Milestone 1 完成標準:**
- 可透過 API 建立 Project
- 可透過 API 觸發 AI CLI 執行
- AI Agent 輸出的 JSON 可被正確解析
- 執行結果正確儲存到資料庫
- 支援即時 Log 查看與詳細除錯

---

### 11.2 Phase 2: Telegram 整合

#### 2.1 Telegram Bot 基礎
- 初始化 `telegram-bot-api`
- 實作 Bot Token 設定與驗證
- 實作白名單 Middleware
- 建立 Update Loop (Goroutine)

#### 2.2 Telegram 指令實作
- 實作 `/help`
- 實作 `/pp 1`, `/pp 2`
- 實作 `/status`
- 實作 `/run project_name 指令` 觸發器
  - 呼叫 Executor Service
- 回覆確認訊息

#### 2.3 通知機制
- 實作 `SendMessage` 封裝
- 執行完成後透過 Channel 通知 Bot Service 發送訊息
- 處理「命令執行中」狀態回覆

#### 2.4 整合測試
- 測試 Telegram 指令處理
- 測試 Mock Bot API

**Milestone 2 完成標準:**
- 可透過 Telegram 查看 Project 列表、下達指令、查詢狀態
- 執行完成後自動推送通知

---

### 11.3 Phase 3: Web 管理介面

#### 3.1 前端環境建置
- 建立 Vue 3 專案 (Vite)
- 安裝 UI 框架與 Axios

#### 3.2 Project 管理頁面
- 實作列表、建立、編輯、詳情頁面
- 整合 Backend API

#### 3.3 執行歷史頁面
- 實作歷史記錄列表與詳情彈窗
- 顯示完整輸出與檔案變更

#### 3.4 系統設定頁面
- 設定 Telegram Bot Token 與白名單

#### 3.5 使用者體驗優化
- Loading, Error Handling, RWD

**Milestone 3 完成標準:**
- Web 介面功能完整，可管理專案與查看歷史

---

### 11.4 Phase 4: 排程功能

#### 4.1 排程管理 API
- 實作 Schedule CRUD API
- 驗證邏輯

#### 4.2 排程執行引擎
- 整合 `robfig/cron`
- 啟動時載入資料庫中的排程
- 動態新增/移除 Cron Job
- 執行時間到觸發 Executor

#### 4.3 排程管理前端
- 實作排程管理介面

#### 4.4 整合測試
- 測試 Cron Job 觸發與執行

**Milestone 4 完成標準:**
- 可設定與執行排程任務，並收到通知

---

### 11.5 Phase 5: 測試與優化

#### 5.1 整合測試
- 端到端測試 (End-to-End)
- 錯誤情境測試

#### 5.2 效能優化
- 資料庫索引
- Goroutine 洩漏檢查 (`runtime/pprof`)

#### 5.3 文件撰寫
- README.md
- Swagger API 文件 (`swaggo/swag`)

#### 5.4 部署準備
- Dockerfile (Multi-stage build)
- docker-compose.yml

**Milestone 5 完成標準:**
- 系統穩定，文件完整，準備部署

---

## 12. 開發注意事項

### 12.1 安全性考量
- **環境變數**: 使用 `os.Setenv` 僅在 Cmd 執行期間設定，並使用 Allowlist
- **Telegram**: 嚴格檢查 `Update.Message.From.ID`
- **路徑**: `filepath.Clean` 與 `filepath.Rel` 檢查
- **SQL Injection**: GORM 參數化查詢
- **Command Injection**: `os/exec` 不透過 Shell 執行 (除非必要且嚴格過濾)

### 12.2 錯誤處理
- 使用 `error` 介面傳遞錯誤
- 統一的 Error Handler Middleware
- 區分 User Error 與 System Error

### 12.3 日誌記錄
- 使用 `slog`
- 結構化日誌 (Structured Logging)

### 12.4 測試策略
- `go test ./...`
- Table-driven tests
- Mocking interfaces (`mockery` or manual mocks)

### 12.5 程式碼品質
- 遵循 `Effective Go`
- `golangci-lint` 檢查
- 註解規範 (GoDoc)

---

## 13. 技術風險與應對

### 13.1 AI CLI 工具相容性
**風險**: 不同 AI CLI 工具的輸出格式、錯誤處理可能不一致，導致解析失敗。

**應對**:
- 設計彈性的 Output Parser，支援多種常見格式。
- 提供豐富的錯誤日誌與原始輸出保留，方便除錯。
- 在文件中列出已測試相容的 CLI 工具版本。

### 13.2 MCP 整合複雜度
**風險**: AI Agent 可能無法正確理解 JSON 回傳格式的要求，導致任務執行成功但系統無法判定。

**應對**:
- 提供清晰且包含範例的 Prompt Template。
- 設計容錯性高的 JSON 提取邏輯（如 regex 搜尋）。
- 實作「未收到結果」的降級處理機制，並通知用戶手動確認。

### 13.3 長時間執行任務
**風險**: 雖然 Goroutines 輕量，但大量長時間運行的 CLI 程序仍可能耗盡系統資源或遇到 Context Timeout。

**應對**:
- 使用 `context.WithTimeout` 設定合理的執行時限。
- 實作 Semaphore 或 Worker Pool 限制同時執行的最大 CLI 程序數量。
- 監控系統資源使用率。

### 13.4 Telegram Bot 限制
**風險**: Telegram API 有速率限制 (Rate Limits) 與訊息長度限制。

**應對**:
- 實作 Rate Limiter Middleware，控制發送頻率。
- 長訊息自動截斷，並提供連結引導至 Web 介面查看完整內容。
- 實作訊息佇列 (Message Queue) 機制，平滑化突發流量。

---

## 14. 未來擴充考量
- gRPC 介面
- 多節點部署 (需更換資料庫與 Session 管理)
- WebSocket 雙向溝通 (取代 SSE)