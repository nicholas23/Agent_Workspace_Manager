package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ParsedOutput 定義解析後的輸出結構
type ParsedOutput struct {
	Status        string   `json:"status"`
	Summary       string   `json:"summary"`
	ModifiedFiles []string `json:"modified_files"`
	CreatedFiles  []string `json:"created_files"`
	DeletedFiles  []string `json:"deleted_files"`
}

// ParseOutput 解析 AI Agent 的 JSON 輸出
// 支援從混合文字中提取 JSON 區塊
func ParseOutput(output string) (*ParsedOutput, error) {
	// 1. 嘗試直接解析 (如果整個輸出就是 JSON)
	var result ParsedOutput
	if err := json.Unmarshal([]byte(output), &result); err == nil {
		return &result, nil
	}

	// 2. 使用正規表達式尋找 JSON 區塊
	// 尋找以 { 開頭，以 } 結尾的區塊，並嘗試解析
	// (?s) 開啟 dot-matches-newline 模式
	re := regexp.MustCompile(`(?s)\{.*\}`)
	matches := re.FindAllString(output, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("no JSON block found in output")
	}

	// 取最後一個匹配的區塊 (假設 AI 可能會輸出多個 JSON，最後一個通常是最終結果)
	lastMatch := matches[len(matches)-1]
	
	// 嘗試解析提取出的 JSON
	if err := json.Unmarshal([]byte(lastMatch), &result); err != nil {
		// 如果直接解析失敗，可能是因為包含了非 JSON 的前後綴 (雖然 regex 已經盡量匹配)
		// 這裡可以做更進一步的清理，例如尋找最外層的 {}
		start := strings.Index(lastMatch, "{")
		end := strings.LastIndex(lastMatch, "}")
		if start != -1 && end != -1 && end > start {
			jsonPart := lastMatch[start : end+1]
			if err := json.Unmarshal([]byte(jsonPart), &result); err == nil {
				return &result, nil
			}
		}
		return nil, fmt.Errorf("failed to parse extracted JSON: %v", err)
	}

	return &result, nil
}
