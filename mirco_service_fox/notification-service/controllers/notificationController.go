package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"notification-service/config"
	"notification-service/database"
	"notification-service/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateNotification 创建通知
func CreateNotification(c *gin.Context) {
	var req struct {
		Title        string `json:"title" binding:"required"`
		Content      string `json:"content" binding:"required"`
		ExperimentID string `json:"experiment_id"`
		IsImportant  bool   `json:"is_important"`
		UserIDs      []uint `json:"user_ids"` // 用户ID列表

	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 通知服务调用用户服务的 API 来验证学生ID是否存在
	client := &http.Client{
		Timeout: time.Second * 5, // 设置5秒超时
	}
	cfg := config.LoadConfig()
	// 验证每个用户ID
	for _, userID := range req.UserIDs {
		// 调用用户服务API
		url := fmt.Sprintf("%s/internal/users/%d", cfg.UserServiceURL, userID)
		resp, err := client.Get(url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "无法连接用户服务",
			})
			return
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("用户 %d 不存在", userID),
			})
			return
		}

		// 解析响应
		var result struct {
			Status string `json:"status"`
			Data   struct {
				ID        uint   `json:"id"`
				Name      string `json:"name"`
				Role      string `json:"role"`
				Telephone string `json:"telephone"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "无法解析用户服务响应",
			})
			return
		}

		// 检查用户角色
		if result.Data.Role != "student" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("用户 %d 不是学生角色", userID),
			})
			return
		}
	}
	// 创建通知
	notification := models.Notification{
		ID:           uuid.New().String(),
		Title:        req.Title,
		Content:      req.Content,
		ExperimentID: req.ExperimentID,
		IsImportant:  req.IsImportant,
		CreatedAt:    time.Now(),
	}

	// 使用事务保存通知和用户关联
	tx := database.DB.Begin()
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	// 保存通知与用户的关联
	for _, userID := range req.UserIDs {
		notificationUser := models.NotificationUser{
			UserID:         strconv.FormatUint(uint64(userID), 10),
			NotificationID: notification.ID,
		}
		if err := tx.Create(&notificationUser).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification user relation"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusCreated, notification)
}

// GetTeacherNotifications 获取教师通知
func GetTeacherNotifications(c *gin.Context) {

	// 解析查询参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	experimentID := c.Query("experiment_id")
	isImportantStr := c.Query("is_important")
	createdAfter := c.Query("created_after")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的页码"})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的每页数量"})
		return
	}
	db := database.GetDB()
	// 构建查询
	query := db.Model(&models.Notification{}).Order("created_at DESC")

	// 筛选条件
	if experimentID != "" {
		query = query.Where("experiment_id = ?", experimentID)
	}
	if isImportantStr != "" {
		isImportant, err := strconv.ParseBool(isImportantStr)
		if err == nil {
			query = query.Where("is_important = ?", isImportant)
		}
	}
	if createdAfter != "" {
		if t, err := time.Parse(time.RFC3339, createdAfter); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取通知总数"})
		return
	}

	// 分页查询
	var notifications []models.Notification
	if err := query.
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取通知列表"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   notifications,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetStudentNotifications 获取学生通知
func GetStudentNotifications(c *gin.Context) {
	studentID := c.Param("student_id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student_id is required"})
		return
	}

	// 解析查询参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	experimentID := c.Query("experiment_id")
	isImportantStr := c.Query("is_important")
	createdAfter := c.Query("created_after")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	// 构建查询 with join table
	query := database.DB.
		Joins("JOIN notification_users ON notification_users.notification_id = notifications.id").
		Where("notification_users.user_id = ?", studentID).
		Order("notifications.created_at DESC")

	// 应用筛选条件
	if experimentID != "" {
		query = query.Where("notifications.experiment_id = ?", experimentID)
	}
	if isImportantStr != "" {
		isImportant, err := strconv.ParseBool(isImportantStr)
		if err == nil {
			query = query.Where("notifications.is_important = ?", isImportant)
		}
	}
	if createdAfter != "" {
		if t, err := time.Parse(time.RFC3339, createdAfter); err == nil {
			query = query.Where("notifications.created_at >= ?", t)
		}
	}

	// 获取总数
	var total int64
	if err := query.Model(&models.Notification{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count notifications"})
		return
	}

	// 分页查询
	var notifications []models.Notification
	if err := query.
		Select("notifications.*").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notifications"})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   notifications,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
