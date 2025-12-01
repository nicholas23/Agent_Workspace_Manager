package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware 處理跨來源資源共享 (CORS) 設定
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 設定允許的來源，* 代表允許所有來源
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// 設定是否允許傳送憑證
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// 設定允許的標頭欄位
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		// 設定允許的 HTTP 方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// 如果是 OPTIONS 預檢請求，直接回傳 204 No Content
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// 繼續處理下一個處理器
		c.Next()
	}
}
