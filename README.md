# Agent Workspace Manager (AI 代理工作區管理器)

一個透過 Telegram 和 Web 介面遠端管理多個 AI Agent 專案的系統。

## 功能特色
- **專案管理**：建立與管理多個專案，每個專案都有獨立的目錄與 AI CLI 指令。
- **遠端執行**：透過 Telegram (`/run`) 或 Web 介面觸發 AI 指令。
- **排程功能**：為專案排程一次性任務。
- **執行歷史**：查看所有執行的詳細日誌。
- **Telegram 整合**：透過 Telegram Bot 接收通知並管理任務。

## 安裝設定

### 前置需求
- Go 1.23+
- Node.js 16+
- SQLite (內嵌)

### 後端設定
1. 進入 `backend/` 目錄。
2. 複製 `.env.example` 為 `.env` (如果有提供) 或自行建立：
   ```env
   PORT=8080
   DATABASE_URL=../data/app.db
   TELEGRAM_BOT_TOKEN=your_bot_token
   TELEGRAM_WHITELIST=your_telegram_id
   ```
3. 啟動伺服器：
   ```bash
   go run cmd/server/main.go
   ```

### 前端設定
1. 進入 `frontend/` 目錄。
2. 安裝依賴套件：
   ```bash
   npm install
   ```
3. 啟動開發伺服器：
   ```bash
   npm run dev
   ```

## 使用說明

### Telegram 指令
- `/help`：顯示可用指令。
- `/pp [page]`：列出專案。
- `/run [project_name] [command]`：執行指令。
- `/status [project_name]`：檢查最後一次執行的狀態。

### Web 介面
- 預設存取網址：`http://localhost:5173` (Vite 預設埠口)。
- 可建立專案、查看歷史記錄與排程任務。

## 開發
- 執行測試：在 `backend/` 目錄下執行 `go test ./...`。

## 關於本專案
本專案是在 AI 模型的協助下開發完成。

## 授權條款
本專案採用 [Apache License 2.0](LICENSE) 授權。
