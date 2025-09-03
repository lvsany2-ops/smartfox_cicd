package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware 生成并传递请求ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取请求ID
		requestID := c.GetHeader("X-Request-ID")

		// 如果没有，生成一个新的
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到响应头
		c.Writer.Header().Set("X-Request-ID", requestID)

		// 设置到请求上下文，方便后续使用
		c.Set("RequestID", requestID)

		c.Next()
	}
}
