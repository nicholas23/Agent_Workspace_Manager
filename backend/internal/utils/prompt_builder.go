package utils

import (
	"agent-workspace-manager/internal/models"
	"fmt"
	"strings"
)

// SystemInstructions 定義了給 AI 模型的系統提示詞，包含格式要求與安全規範
// 注意：Go 的反引號字串常數中無法直接包含反引號，需使用 + 連接
const SystemInstructions = `
【安全警告】
1. 你只能存取、修改或建立位於工作目錄內的檔案。
2. 嚴禁使用絕對路徑 (如 /etc/passwd)。
3. 嚴禁使用上層目錄路徑 (如 ../secret.txt)。
4. 所有檔案路徑必須是相對於工作目錄的相對路徑。
5. 嚴禁執行系統管理指令 (如 rm -rf, shutdown, reboot) 或網路指令 (curl, wget)。
6. 嚴禁輸出或洩漏任何環境變數 (Environment Variables)。

請執行以下任務,並在完成後以一定要以JSON格式輸出結果。不需回應其他內容只輸出JSON，內容如下 

【重要】JSON 格式規範:
{
  "status": "success 或 failed",
  "summary": "50字以內的執行摘要",
  "details": "完整執行過程",
  "modified_files": ["相對路徑1", "相對路徑2"],
  "created_files": ["相對路徑3"],
  "deleted_files": ["相對路徑4"]
}

`

// BuildPrompt 建構完整的 Prompt 內容 (不包含 CLI 指令部分)
// userPrompt: 使用者輸入的提示詞
// history: 最近的執行記錄
// project: 專案資訊
func BuildPrompt(userPrompt string, history []models.Execution, project models.Project) string {
	// 建構歷史記錄字串
	var historyBuilder strings.Builder
	if len(history) > 0 {
		var no int = 1
		historyBuilder.WriteString("過去執行歷史 (Context)】\n")
		// 反轉順序，讓舊的在前，新的在後，符合閱讀邏輯 (假設傳入的是倒序)
		for i := len(history) - 1; i >= 0; i-- {
			exec := history[i]
			historyBuilder.WriteString(fmt.Sprintf("- No. %d \n\t- 時間: %s\n\t- 指令: %s\n\t- 結果: %s\n\t- 摘要: %s\n",no,exec.StartTime,exec.Command, exec.Status, exec.Summary))
			no++
		}
	} 

	// 組合系統指令、專案資訊、歷史記錄與使用者提示詞
	projectInfo := fmt.Sprintf("【專案資訊】\n- 專案名稱: %s\n- 工作目錄: %s\n", project.Name, project.DirectoryPath)
	
	fullPrompt := fmt.Sprintf("%s\n%s\n%s\n【任務內容】\n%s", SystemInstructions, historyBuilder.String(), projectInfo, userPrompt)
	
	
	return fullPrompt
}
