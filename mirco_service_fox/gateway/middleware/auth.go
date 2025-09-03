package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	userRoleKey contextKey = "userRole"
)

// JWT密钥（应与用户服务中的密钥一致）
func getJWTKey() []byte {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		return []byte("a_secret_crect") // 默认密钥，生产环境不建议使用默认值
	}
	return []byte(key)
}

// Claims token的claim结构
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || strings.HasPrefix(c.Request.URL.Path, "/api/auth") {
			c.Next()
			return
		}
		// 获取authorization header
		tokenString := c.GetHeader("Authorization")
		log.Printf("Authorization header: %s", tokenString) // 添加日志
		// 验证token格式
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "权限不足"})
			c.Abort()
			return
		}

		// 提取token的有效部分
		tokenString = tokenString[7:]

		// 解析token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return getJWTKey(), nil
		})
		if err != nil {
			log.Printf("Token parsing error: %v", err) // 添加日志
			c.JSON(http.StatusUnauthorized, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Printf("Invalid token: %s", tokenString) // 添加日志
			c.JSON(http.StatusUnauthorized, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}
		log.Printf("Token validated successfully. UserID: %d, Role: %s", claims.UserID, claims.Role) // 添加日志
		// 检查路径和角色权限
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/teacher") && claims.Role != "teacher" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "需要教师权限"})
			c.Abort()
			return
		}

		if strings.HasPrefix(path, "/api/student") && claims.Role != "student" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "需要学生权限"})
			c.Abort()
			return
		}
		// 将用户信息添加到请求头中
		c.Set("userID", string(strconv.FormatUint(uint64(claims.UserID), 10)))
		c.Set("userRole", claims.Role)
		ctx := context.WithValue(c.Request.Context(), userIDKey, strconv.FormatUint(uint64(claims.UserID), 10))
		ctx = context.WithValue(ctx, userRoleKey, claims.Role)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
