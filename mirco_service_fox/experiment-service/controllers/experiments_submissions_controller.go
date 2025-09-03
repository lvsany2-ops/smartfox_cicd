package controllers

import (
	"encoding/json"
	"errors"
	"experiment-service/config"
	"experiment-service/models"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetStudentSubmission(experimentId string, studentId uint) *models.ExperimentSubmission {
	cfg := config.LoadConfig()
	submission := &models.ExperimentSubmission{}
	submissionServiceURL := fmt.Sprintf("%s/api/student/submissions/%s/%d/status", cfg.SubmissionServiceURL, experimentId, studentId)

	resp, err := http.Get(submissionServiceURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result struct {
			Status string                      `json:"status"`
			Data   models.ExperimentSubmission `json:"data"`
		}
		if err := json.Unmarshal(body, &result); err == nil && result.Status == "success" {
			*submission = result.Data
		}
	}

	return submission
}

type GSQSRequest struct {
	SubId string `json:"submission_id" binding:"required"`
	QId   string `json:"question_id" bingding:"required"`
}

func GetStudentQuestionSubmission(submissionId string, qId string) *models.QuestionSubmission {
	cfg := config.LoadConfig()
	submissionServiceURL := fmt.Sprintf("%s/api/student/submissions/%s/GetStudentAns", cfg.SubmissionServiceURL, submissionId)

	// 构建请求数据
	requestData := GSQSRequest{
		SubId: submissionId,
		QId:   qId,
	}

	// 序列化请求数据
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Printf("序列化请求数据失败: %v\n", err)
		return nil
	}

	// 创建POST请求
	resp, err := http.Post(submissionServiceURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		fmt.Printf("发送HTTP请求失败: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("请求失败，状态码: %d\n", resp.StatusCode)
		return nil
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return nil
	}

	// 解析响应
	var result struct {
		Status string                    `json:"status"`
		Data   models.QuestionSubmission `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		return nil
	}

	if result.Status != "success" {
		fmt.Printf("接口返回错误状态: %s\n", result.Status)
		return nil
	}

	return &result.Data
}
func UpdateSubmissionsInProgress(experimentId string) {
	cfg := config.LoadConfig()
	submissionServiceURL := fmt.Sprintf("%s/api/student/submissions/%s/UpdateExperimentStatus", cfg.SubmissionServiceURL, experimentId)

	// 创建HTTP请求
	req, err := http.NewRequest("PUT", submissionServiceURL, nil)
	if err != nil {
		fmt.Printf("创建HTTP请求失败: %v\n", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送HTTP请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	// 检查响应状态
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("实验状态更新成功: %s\n", string(body))
	} else {
		fmt.Printf("实验状态更新失败，状态码: %d, 响应: %s\n", resp.StatusCode, string(body))
	}
}

func GetExperimentById(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	db := config.DB
	tx := db.Begin()
	//验证试验是否存在
	var experiment models.Experiment
	if err := tx.Where("id = ?", experimentID).First(&experiment).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Experiment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Database error"})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
		tx.Commit()
	}
}

type VQrequest struct {
	ExperimentId string   `json:"experiment_id" binding:"required"`
	QuestionIds  []string `json:"question_ids" binding:"required"`
}

type VQresponse struct {
	Status           string          `json:"status"`
	ValidQuestionMap map[string]bool `json:"valid_question_map"`
}

func ValidQuestion(c *gin.Context) {
	var req VQrequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request"})
		return
	}
	db := config.DB
	tx := db.Begin()

	var validQuestions []models.Question
	if err := tx.Where("experiment_id = ?", req.ExperimentId).
		Where("id IN ?", req.QuestionIds).
		Find(&validQuestions).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to validate questions"})
		return
	}

	validQuestionMap := make(map[string]bool)
	for _, q := range validQuestions {
		validQuestionMap[q.ID] = true
	}

	c.JSON(http.StatusOK, VQresponse{ValidQuestionMap: validQuestionMap})
}

// GetQuestionDetailRequest 请求体
type GetQuestionDetailRequest struct {
	QuestionID string `json:"question_id" binding:"required"`
}

// GetQuestionDetailResponse 响应体
type GetQuestionDetailResponse struct {
	Question models.Question `json:"question"`
}

func GetQuestionDetail(c *gin.Context) {
	var req GetQuestionDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request"})
		return
	}

	db := config.DB
	var question models.Question
	if err := db.Where("id = ?", req.QuestionID).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Question not found"})
		return
	}

	c.JSON(http.StatusOK, GetQuestionDetailResponse{Question: question})
}

// GetExperimentDetailRequest 请求体
type GetExperimentDetailRequest struct {
	ExperimentID string `json:"experiment_id" binding:"required"`
}

// GetExperimentDetailResponse 响应体
type GetExperimentDetailResponse struct {
	ExperimentID string    `json:"experiment_id"`
	Permission   int       `json:"permission"`
	Deadline     time.Time `json:"deadline"`
	Title        string    `json:"title"`
	IsExpired    bool      `json:"is_expired"`
	TotalScore   int       `json:"total_score"`
}

func GetExperimentDetail(c *gin.Context) {
	var req GetExperimentDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request"})
		return
	}

	db := config.DB
	var experiment models.Experiment
	if err := db.Preload("Questions").First(&experiment, "id = ?", req.ExperimentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Experiment not found"})
		return
	}

	totalScore := 0
	for _, q := range experiment.Questions {
		totalScore += q.Score
	}
	isExpired := experiment.Permission == 0 && time.Now().After(experiment.Deadline)

	c.JSON(http.StatusOK, GetExperimentDetailResponse{
		ExperimentID: experiment.ID,
		Permission:   experiment.Permission,
		Deadline:     experiment.Deadline,
		Title:        experiment.Title,
		IsExpired:    isExpired,
		TotalScore:   totalScore,
	})
}

// callSubmissionService 调用提交服务的API删除实验相关提交记录
func callSubmissionService(experimentID string) error {
	cfg := config.LoadConfig()
	submissionServiceURL := fmt.Sprintf("%s/api/student/experiments/%s/submissions", cfg.SubmissionServiceURL, experimentID)

	// 创建HTTP请求
	req, err := http.NewRequest("DELETE", submissionServiceURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("submission service returned status: %d", resp.StatusCode)
	}

	return nil
}
