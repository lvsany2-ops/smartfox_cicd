package controllers

import (
	"encoding/json"
	"errors"
	"experiment-service/config"
	"experiment-service/models"
	"experiment-service/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetExperiments_Student(c *gin.Context) {
	db := config.DB
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.DefaultQuery("status", "all")

	offset := (page - 1) * limit
	// 从请求头获取用户ID（微服务架构）
	studentIDStr := c.GetHeader("X-User-ID")
	if studentIDStr == "" {
		studentIDStr = c.Query("user_id") // 也支持查询参数
	}
	if studentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "User ID is required",
		})
		return
	}
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}
	now := time.Now()

	// 查询分配给该学生的实验
	var experiments []models.Experiment
	// 使用JSON查询来查找包含当前学生ID的实验
	// 对于 JSON 数组中包含整数值的查询，我们需要直接传递整数值
	query := db.Model(&models.Experiment{}).Where("JSON_CONTAINS(user_ids, CAST(? AS JSON))", studentID)

	// 状态筛选逻辑
	switch status {
	case "active":
		query = query.Where("deadline > ?", now)
	case "expired":
		query = query.Where("deadline <= ?", now)
	}

	var total int64
	query.Count(&total)

	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&experiments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "数据库查询失败",
		})
		return
	}

	// 构建响应数据
	experimentResponses := make([]gin.H, len(experiments))
	for i, exp := range experiments {
		// // 查询学生的提交状态
		// var submission models.ExperimentSubmission
		// submissionStatus := "not_started"
		// if err := db.Where("experiment_id = ? AND student_id = ?", exp.ID, studentID).Order("created_at DESC").
		// 	First(&submission).Error; err == nil {
		// 	submissionStatus = strings.ToLower(submission.Status)
		// }
		submissionStatus := GetStudentSubmission(exp.ID, uint(studentID)).Status

		// 确定实验状态
		expStatus := "active"
		if exp.Deadline.Before(now) {
			expStatus = "expired"
		}

		experimentResponses[i] = gin.H{
			"experiment_id":     exp.ID,
			"title":             exp.Title,
			"description":       exp.Description,
			"deadline":          exp.Deadline.Format(time.RFC3339),
			"status":            expStatus,
			"submission_status": submissionStatus,
		}
	}

	// 返回分页响应
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   experimentResponses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetExperiment 获取实验详细信息
func GetExperimentDetail_Student(c *gin.Context) {
	// 获取实验 ID
	db := config.DB
	// 从请求头获取用户ID（微服务架构）
	studentIDStr := c.GetHeader("X-User-ID")
	if studentIDStr == "" {
		studentIDStr = c.Query("user_id") // 也支持查询参数
	}
	if studentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "User ID is required",
		})
		return
	}
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}
	experimentID := c.Param("experiment_id")
	var experiment models.Experiment
	// 查询实验详情，包括关联的阶段和资源
	result := db.Preload("Questions").Preload("Attachments").
		Where("ID = ?", experimentID).
		First(&experiment)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Experiment not found"})
		return
	}

	// 获取学生提交记录
	var submission = GetStudentSubmission(experimentID, uint(studentID))
	submissionStatus := "not_started"
	totalScore := 0

	submissionStatus = strings.ToLower(submission.Status)
	totalScore = submission.TotalScore

	// 获取学生答案
	questionResponses := make([]gin.H, len(experiment.Questions))
	for i, q := range experiment.Questions {
		questionData := gin.H{
			"question_id": q.ID,
			"type":        q.Type,
			"content":     q.Content,
			"score":       q.Score,
			"image_url":   q.ImageURL,
		}

		// 选择题添加选项
		if q.Type == "choice" {
			var options []string
			json.Unmarshal([]byte(q.Options), &options)
			questionData["options"] = options
		}
		if experiment.Deadline.Before(time.Now()) {
			if q.Type != "code" {
				questionData["correct_answer"] = q.CorrectAnswer
			}
			questionData["explanation"] = q.Explanation
		}
		// 获取学生答案和反馈
		var qSubmission = GetStudentQuestionSubmission(submission.ID, q.ID)

		if qSubmission != nil {
			if q.Type == "code" {
				questionData["student_code"] = qSubmission.Code
				questionData["student_language"] = qSubmission.Language
			} else {
				questionData["student_answer"] = qSubmission.Answer
			}
			// if experiment.Deadline.Before(time.Now()) {
			questionData["feedback"] = qSubmission.Feedback
			// }
		} else {
			// 如果没有提交记录，设置默认值
			if q.Type == "code" {
				questionData["student_code"] = ""
				questionData["student_language"] = ""
			} else {
				questionData["student_answer"] = ""
			}
			questionData["feedback"] = ""
		}

		questionResponses[i] = questionData
	}
	attachmentResponses := make([]gin.H, len(experiment.Attachments))
	for i, a := range experiment.Attachments {
		attachmentResponses[i] = gin.H{
			"id":   a.ID,
			"name": a.Name,
			"url":  a.URL,
		}
	}
	// 构建响应
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"experiment_id":     experiment.ID,
			"permission":        experiment.Permission,
			"title":             experiment.Title,
			"description":       experiment.Description,
			"deadline":          experiment.Deadline.Format(time.RFC3339),
			"questions":         questionResponses,
			"attachments":       attachmentResponses,
			"submission_status": submissionStatus,
			"total_score":       totalScore,
		},
	})
}

// CreateExperiment 创建实验
func CreateExperiment(c *gin.Context) {

	// QuestionInput 题目输入结构体
	type QuestionInput struct {
		Type          string            `json:"type" binding:"required,oneof=choice blank code"`
		Content       string            `json:"content" binding:"required"`
		Options       []string          `json:"options" binding:"required_if=Type choice"`
		CorrectAnswer string            `json:"correct_answer" binding:"required_if=Type choice required_if=Type blank"`
		Score         int               `json:"score" binding:"required,gt=0"`
		ImageURL      string            `json:"image_url" binding:"omitempty"`
		Explanation   string            `json:"explanation" binding:"omitempty"`
		TestCases     []models.TestCase `json:"test_cases" binding:"required_if=Type code"`
	}
	// CreateExperimentRequest 请求结构体
	type CreateExperimentRequest struct {
		Title       string          `json:"title" binding:"required"`
		Description string          `json:"description"`
		Permission  *int            `json:"permission" binding:"required,oneof=1 0"`
		Deadline    time.Time       `json:"deadline" binding:"required"`
		StudentIDs  []string        `json:"student_ids" binding:"required"`
		Questions   []QuestionInput `json:"questions" binding:"required,dive"`
	}
	// ExperimentResponseData 响应数据
	type ExperimentResponseData struct {
		ExperimentID string    `json:"experiment_id"`
		Title        string    `json:"title"`
		CreatedAt    time.Time `json:"created_at"`
	}
	// CreateExperimentResponse 响应结构体
	type CreateExperimentResponse struct {
		Status  string                 `json:"status"`
		Data    ExperimentResponseData `json:"data,omitempty"`
		Message string                 `json:"message,omitempty"`
	}

	db := config.DB
	var req CreateExperimentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, CreateExperimentResponse{
			Status:  "error",
			Message: fmt.Sprintf("Invalid request data: %s. Request Method: %s, Request URL: %s, Request Body: %s", err.Error(), c.Request.Method, c.Request.URL.String(), c.Request.Body),
		})
		return
	}
	// 验证截止日期
	if req.Deadline.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, CreateExperimentResponse{
			Status:  "error",
			Message: "截止日期必须在未来",
		})
		return
	}
	experimentID := uuid.New().String()
	// 处理附件上传
	form, err := c.MultipartForm()
	var attachments []models.Attachment
	if err == nil && form.File["attachments"] != nil {
		for _, file := range form.File["attachments"] {
			// 生成唯一文件名
			fileExt := filepath.Ext(file.Filename)
			fileName := uuid.New().String() + fileExt
			filePath := filepath.Join("uploads", fileName)

			// 确保上传目录存在
			if err := os.MkdirAll("uploads", 0755); err != nil {
				c.JSON(http.StatusInternalServerError, CreateExperimentResponse{
					Status:  "error",
					Message: "无法创建上传目录",
				})
				return
			}
			// 保存文件
			if err := c.SaveUploadedFile(file, filePath); err != nil {
				c.JSON(http.StatusInternalServerError, CreateExperimentResponse{
					Status:  "error",
					Message: "无法保存附件",
				})
				return
			}
			// 生成附件 URL（假设服务器地址为 localhost:8080）
			fileURL := fmt.Sprintf("/uploads/%s", fileName)
			attachments = append(attachments, models.Attachment{
				ExperimentID: experimentID,
				Name:         file.Filename,
				URL:          fileURL,
			})
		}
	}

	// 创建实验
	experiment := models.Experiment{
		ID:          experimentID,
		Title:       req.Title,
		Permission:  *req.Permission,
		Description: req.Description,
		Deadline:    req.Deadline,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Attachments: attachments,
	}
	// 处理题目
	for _, q := range req.Questions {
		question := models.Question{
			ID:           uuid.NewString(),
			ExperimentID: experimentID,
			Type:         q.Type,
			Content:      q.Content,
			Score:        q.Score,
			ImageURL:     q.ImageURL,
			Explanation:  q.Explanation,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if q.Type == "choice" {
			// 将选项序列化为 JSON 字符串
			optionsJSON, _ := json.Marshal(q.Options)
			question.Options = string(optionsJSON)
			question.CorrectAnswer = q.CorrectAnswer
		}
		if q.Type == "blank" {
			question.CorrectAnswer = q.CorrectAnswer
		}
		if q.Type == "code" && len(q.TestCases) > 0 {
			testCasesJSON, _ := json.Marshal(q.TestCases)
			question.TestCases = string(testCasesJSON)
		}
		experiment.Questions = append(experiment.Questions, question)
	}
	var userIDs models.JSONIntSlice
	for _, studentID := range req.StudentIDs {
		userIDs = append(userIDs, int(utils.StrToUint(studentID)))
	}
	experiment.UserIDs = userIDs
	// 保存到数据库
	if err := db.Create(&experiment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, CreateExperimentResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	//下发通知
	notificationRequest := map[string]interface{}{
		"title":         fmt.Sprintf("新实验发布：%s", experiment.Title),
		"content":       fmt.Sprintf("您有一个新的实验《%s》，请在 %s 前完成提交。", experiment.Title, experiment.Deadline.Format("2006-01-02 15:04")),
		"experiment_id": experiment.ID,
		"is_important":  false,
		"user_ids":      userIDs, // 直接复用前面查到的学生
	}

	//调用通知接口
	if err := callNotificationService(notificationRequest); err != nil {
		fmt.Printf("创建通知失败: %v\n", err)
	}

	// 返回成功响应
	c.JSON(http.StatusCreated, CreateExperimentResponse{
		Status: "success",
		Data: ExperimentResponseData{
			ExperimentID: experiment.ID,
			Title:        experiment.Title,
			CreatedAt:    experiment.CreatedAt,
		},
	})

}

// UpdateExperiment 更新实验
func UpdateExperiment(c *gin.Context) {
	type UpdateQuestionInput struct {
		QuestionID    string            `json:"question_id" binding:"omitempty,required_if=Type ''"`
		Type          string            `json:"type" binding:"omitempty,oneof=choice blank code"`
		Content       string            `json:"content" binding:"omitempty,min=1"`
		Options       []string          `json:"options" binding:"omitempty,required_if=Type choice"`
		CorrectAnswer string            `json:"correct_answer" binding:"omitempty,required_if=Type choice required_if=Type blank"`
		Score         int               `json:"score" binding:"omitempty,gt=0"`
		ImageURL      string            `json:"image_url" binding:"omitempty"`
		Explanation   string            `json:"explanation" binding:"omitempty"`
		TestCases     []models.TestCase `json:"test_cases" binding:"omitempty,required_if=Type code"`
	}
	type UpdateExperimentRequest struct {
		Title           string                `json:"title" binding:"omitempty,min=1"`
		Description     string                `json:"description" binding:"omitempty"`
		Deadline        time.Time             `json:"deadline" binding:"omitempty"`
		Questions       []UpdateQuestionInput `json:"questions" binding:"omitempty,dive"`
		RemoveQuestions []string              `json:"remove_questions" binding:"omitempty"`
		Permission      *int                  `json:"permission" binding:"omitempty,oneof=0 1"`
	}
	type UpdateExperimentResponse struct {
		Status       string    `json:"status"`
		ExperimentID string    `json:"experiment_id,omitempty"`
		Title        string    `json:"title,omitempty"`
		UpdatedAt    time.Time `json:"updated_at,omitempty"`
		Message      string    `json:"message,omitempty"`
	}

	db := config.DB
	experimentID := c.Param("experiment_id")

	// 检查实验是否存在
	var experiment models.Experiment
	if err := db.Preload("Questions").Preload("Attachments").
		Where("id = ?", experimentID).First(&experiment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, UpdateExperimentResponse{
				Status:  "error",
				Message: "Experiment not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, UpdateExperimentResponse{
				Status:  "error",
				Message: "Failed to fetch experiment",
			})
		}
		return
	}

	var req UpdateExperimentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, UpdateExperimentResponse{
			Status:  "error",
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		// 更新实验基本信息
		if req.Title != "" {
			experiment.Title = req.Title
		}
		if req.Description != "" {
			experiment.Description = req.Description
		}
		if req.Permission != nil {
			experiment.Permission = *req.Permission
		}
		if !req.Deadline.IsZero() {
			if req.Deadline.Before(time.Now()) {
				return errors.New("deadline must be in the future")
			}
			experiment.Deadline = req.Deadline
		}

		// 处理附件上传
		form, err := c.MultipartForm()
		var newAttachments []models.Attachment
		if err == nil && form.File["attachments"] != nil {
			for _, file := range form.File["attachments"] {
				fileExt := filepath.Ext(file.Filename)
				fileName := uuid.New().String() + fileExt
				filePath := filepath.Join("uploads", fileName)

				if err := os.MkdirAll("uploads", 0755); err != nil {
					return err
				}
				if err := c.SaveUploadedFile(file, filePath); err != nil {
					return err
				}
				fileURL := fmt.Sprintf("/uploads/%s", fileName)
				newAttachments = append(newAttachments, models.Attachment{
					ExperimentID: experimentID,
					Name:         file.Filename,
					URL:          fileURL,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				})
			}
			experiment.Attachments = append(experiment.Attachments, newAttachments...)
		}

		// 映射已存在题目
		existingQuestions := make(map[string]*models.Question)
		for i, q := range experiment.Questions {
			existingQuestions[q.ID] = &experiment.Questions[i]
		}

		//var newQuestions []models.Question
		for _, q := range req.Questions {
			if q.QuestionID != "" {
				// 更新现有题目
				if question, exists := existingQuestions[q.QuestionID]; exists {
					updated := false

					if q.Content != "" {
						question.Content = q.Content
						updated = true
					}
					if q.CorrectAnswer != "" {
						question.CorrectAnswer = q.CorrectAnswer
						updated = true
					}
					if q.Score > 0 {
						question.Score = q.Score
						updated = true
					}
					if q.Type != "" {
						question.Type = q.Type
						updated = true
					}
					if q.ImageURL != "" {
						question.ImageURL = q.ImageURL
						updated = true
					}
					if q.Explanation != "" {
						question.Explanation = q.Explanation
						updated = true
					}
					if len(q.Options) > 0 {
						optionsJSON, err := json.Marshal(q.Options)
						if err != nil {
							c.JSON(http.StatusInternalServerError, UpdateExperimentResponse{
								Status:  "error",
								Message: "Unable to serialize options",
							})
							return err
						}
						question.Options = string(optionsJSON)
						updated = true
					}
					if len(q.TestCases) > 0 {
						testCasesJSON, err := json.Marshal(q.TestCases)
						if err != nil {
							c.JSON(http.StatusInternalServerError, UpdateExperimentResponse{
								Status:  "error",
								Message: "Unable to serialize test cases",
							})
							return err
						}
						question.TestCases = string(testCasesJSON)
						updated = true
					}
					// 显式保存更改到数据库
					if updated {
						if err := tx.Save(question).Error; err != nil {
							c.JSON(http.StatusInternalServerError, UpdateExperimentResponse{
								Status:  "error",
								Message: "Failed to update question " + question.ID,
							})
							return err
						}
					}
				} else {
					return fmt.Errorf("question %s not found", q.QuestionID)
				}
			} else {
				// 新增题目
				newQ := models.Question{
					ID:           uuid.NewString(),
					ExperimentID: experimentID,
					Type:         q.Type,
					Content:      q.Content,
					Score:        q.Score,
					ImageURL:     q.ImageURL,
					Explanation:  q.Explanation,
				}
				if q.Type == "choice" {
					optionsJSON, err := json.Marshal(q.Options)
					if err != nil {
						return fmt.Errorf("failed to serialize options: %w", err)
					}
					newQ.Options = string(optionsJSON)
					newQ.CorrectAnswer = q.CorrectAnswer
				}
				if q.Type == "blank" {
					newQ.CorrectAnswer = q.CorrectAnswer
				}
				if q.Type == "code" && len(q.TestCases) > 0 {
					testCasesJSON, _ := json.Marshal(q.TestCases)
					newQ.TestCases = string(testCasesJSON)
				}
				if err := tx.Create(&newQ).Error; err != nil {
					return fmt.Errorf("failed to create new question: %w", err)
				}
				experiment.Questions = append(experiment.Questions, newQ)
			}
		}
		// 删除题目，并从 experiment.Questions 中移除
		var remainingQuestions []models.Question
		for _, question := range experiment.Questions {
			shouldDelete := false
			for _, qID := range req.RemoveQuestions {
				if question.ID == qID {
					shouldDelete = true

					// // 1. 先删除与该题目相关的所有提交记录
					// if err := tx.Where("question_id = ?", qID).Delete(&models.QuestionSubmission{}).Error; err != nil {
					// 	return fmt.Errorf("failed to delete question submissions for question %s: %w", qID, err)
					// }
					//调用
					// 2. 再删除题目本身
					if err := tx.Delete(&models.Question{}, "id = ? AND experiment_id = ?", qID, experimentID).Error; err != nil {
						return fmt.Errorf("failed to delete question %s: %w", qID, err)
					}

					break
				}
			}
			if !shouldDelete {
				remainingQuestions = append(remainingQuestions, question)
			}
		}
		experiment.Questions = remainingQuestions
		experiment.UpdatedAt = time.Now()
		// 保存实验本体
		if err := tx.Save(&experiment).Error; err != nil {
			return fmt.Errorf("failed to save experiment: %w", err)
		}
		return nil
	})
	//更新experiment_submissions表中对应实验状态为in_progress
	UpdateSubmissionsInProgress(experimentID)
	// db.Model(&models.ExperimentSubmission{}).Where("experiment_id = ?", experimentID).Update("status", "in_progress")
	if err != nil {
		c.JSON(http.StatusBadRequest, UpdateExperimentResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UpdateExperimentResponse{
		Status:       "success",
		ExperimentID: experiment.ID,
		Title:        experiment.Title,
		UpdatedAt:    experiment.UpdatedAt,
	})
}

// DeleteExperiment 删除实验及关联数据
func DeleteExperiment(c *gin.Context) {
	db := config.DB

	// 需要验证嘛？
	// 验证用户角色
	userRole := c.GetHeader("X-User-Role")
	if userRole != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "仅允许教师操作",
		})
		return
	}

	// 获取实验ID
	experimentID := c.Param("experiment_id")

	// 查询实验是否存在
	var experiment models.Experiment
	result := db.Preload("Questions").Preload("Attachments").First(&experiment, "id = ?", experimentID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Experiment not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "数据库查询失败",
		})
		return
	}
	tx := db.Begin()

	// 1. 调用submission-service删除关联的提交记录
	if err := callSubmissionService(experimentID); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "删除实验提交记录失败: " + err.Error(),
		})
		return
	}

	// 2. 删除关联题目
	if err := tx.Where("experiment_id = ?", experimentID).Delete(&models.Question{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "删除题目失败",
		})
		return
	}

	// 3. 删除关联附件
	if err := tx.Where("experiment_id = ?", experimentID).Delete(&models.Attachment{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "删除附件失败",
		})
		return
	}

	// 4. 删除实验本身
	if err := tx.Delete(&experiment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "删除实验失败",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Experiment deleted",
	})
}

func HandleTeacherUpload(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Experiment ID is required"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error getting file: %v", err)})
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to open uploaded file: %v", err)})
		return
	}
	defer src.Close()

	// 使用 filepath.Base 获取安全的文件名
	filename := filepath.Base(file.Filename)

	// 当 OSS 未初始化（NO_OSS_INIT）时，写入本地文件系统作为降级方案
	if config.Bucket == nil {
		localDir := filepath.Join("uploads", experimentID)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create upload dir: %v", err)})
			return
		}
		dstPath := filepath.Join(localDir, filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create file: %v", err)})
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save file: %v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":      "File uploaded successfully (local)",
			"experimentId": experimentID,
			"filename":     filename,
			"path":         dstPath,
		})
		return
	}

	// 构建 OSS 中的对象键 (Object Key)
	objectKey := fmt.Sprintf("%s%s/%s", config.OssExperimentPrefix, experimentID, filename)
	if err = config.Bucket.PutObject(objectKey, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload file to OSS: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":      "File uploaded successfully to OSS",
		"experimentId": experimentID,
		"filename":     filename,
		"objectKey":    objectKey,
	})
}

// handleStudentListFilesOSS 列出指定实验下 OSS 中的所有文件
func HandleStudentListFiles(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Experiment ID is required"})
		return
	}
	// 本地降级：列出 uploads/<experimentID> 下的文件
	if config.Bucket == nil {
		localDir := filepath.Join("uploads", experimentID)
		entries, err := os.ReadDir(localDir)
		files := []string{}
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					files = append(files, e.Name())
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"experimentId": experimentID, "files": files})
		return
	}

	// 构造OSS对象前缀，用于列举
	prefixToList := fmt.Sprintf("%s%s/", config.OssExperimentPrefix, experimentID)
	var files []string
	marker := ""
	for {
		lsRes, err := config.Bucket.ListObjects(oss.Marker(marker), oss.Prefix(prefixToList))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list objects from OSS: %v", err)})
			return
		}
		for _, object := range lsRes.Objects {
			if !strings.HasSuffix(object.Key, "/") || object.Size > 0 {
				fileName := strings.TrimPrefix(object.Key, prefixToList)
				if fileName != "" {
					files = append(files, fileName)
				}
			}
		}
		if !lsRes.IsTruncated {
			break
		}
		marker = lsRes.NextMarker
	}
	c.JSON(http.StatusOK, gin.H{"experimentId": experimentID, "files": files})
}

// handleStudentDownloadFileOSS 生成预签名URL供学生下载文件
func HandleStudentDownloadFile(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	filename := c.Param("filename") // 获取原始文件名

	if experimentID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Experiment ID and filename are required"})
		return
	}
	// 本地降级：直接读取 uploads/<experimentID>/<filename> 并下发
	if config.Bucket == nil {
		localPath := filepath.Join("uploads", experimentID, filepath.Base(filename))
		f, err := os.Open(localPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			}
			return
		}
		defer f.Close()
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(filename)))
		if _, err := io.Copy(c.Writer, f); err != nil {
			// 写流失败
			return
		}
		return
	}

	// 构建完整的OSS对象键
	objectKey := fmt.Sprintf("%s%s/%s", config.OssExperimentPrefix, experimentID, filepath.Base(filename))
	isExist, err := config.Bucket.IsObjectExist(objectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking object existence"})
		return
	}
	if !isExist {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found in OSS"})
		return
	}
	signedURL, err := config.Bucket.SignURL(objectKey, oss.HTTPGet, config.SignedURLExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to sign URL: %v", err)})
		return
	}
	c.Redirect(http.StatusFound, signedURL)

}
