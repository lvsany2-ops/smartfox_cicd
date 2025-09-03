package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"submission/config"
	"submission/global"
	"submission/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SaveAnswer(c *gin.Context) {
	db := global.DB
	experimentID := c.Param("experiment_id")
	ID, _ := strconv.ParseUint(c.GetHeader("X-User-ID"), 10, 32)
	studentID := uint(ID)
	//studentID1, _ := c.Get("studentID")
	//转为uint
	//studentID, err := strconv.ParseUint(fmt.Sprintf("%v", studentID1), 10, 32)
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid student ID"})
	//	return
	//}
	type Answer struct {
		QuestionID string `json:"question_id" binding:"required"`
		Type       string `json:"type" binding:"required,oneof=choice blank code"`
		Answer     string `json:"answer" binding:"required_if=Type choice required_if=Type blank"`
		Code       string `json:"code" binding:"required_if=Type code"`
		Language   string `json:"language" binding:"required_if=Type code,oneof=cpp java python"`
	}
	var req struct {
		Answers []Answer `json:"answers" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request"})
		return
	}

	now := time.Now()
	tx := db.Begin()
	cfg := config.LoadConfig()

	//验证试验是否存在
	//localhost:8082/api/experiments/getById/{experiment_id}
	experimentURL := fmt.Sprintf("%s/api/experiments/getById/%s", cfg.ExperimentServiceURL, experimentID)
	resp, err := http.Get(experimentURL)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to verify experiment"})
		return
	} else if resp.StatusCode == http.StatusNotFound {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Experiment not found"})
		return
	} else if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("experiments/getById service error: %s", string(bodyBytes))})
		return
	}
	defer resp.Body.Close()

	//var experiment models.Experiment
	//if err := tx.Where("id = ?", experimentID).First(&experiment).Error; err != nil {
	//	tx.Rollback()
	//	if errors.Is(err, gorm.ErrRecordNotFound) {
	//		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Experiment not found"})
	//	} else {
	//		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Database error"})
	//	}
	//	return
	//}

	// 处理实验提交记录（保持不变）
	var submission models.ExperimentSubmission
	result := tx.Where("experiment_id = ? AND student_id = ? AND status != 'submitted'", experimentID, studentID).
		Order("created_at DESC").
		First(&submission)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		submission = models.ExperimentSubmission{
			ID:           uuid.New().String(),
			ExperimentID: experimentID,
			StudentID:    studentID,
			Status:       "in_progress",
			SubmittedAt:  now,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := tx.Create(&submission).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create submission"})
			return
		}
	} else if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Database error"})
		return
	} else {
		submission.UpdatedAt = now
		if err := tx.Save(&submission).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update submission"})
			return
		}
	}
	tx.Commit()
	tx = db.Begin()
	// 3. 处理每道题的提交
	validQuestionIDs := make([]string, len(req.Answers))

	for i, ans := range req.Answers {
		validQuestionIDs[i] = ans.QuestionID
	}
	// 验证题目属于当前实验
	//localhost:8082/api/experiments/validate-questions
	validationURL := fmt.Sprintf("%s/api/experiments/validQuestion", cfg.ExperimentServiceURL)
	validationPayload := gin.H{
		"experiment_id": experimentID,
		"question_ids":  validQuestionIDs,
	}

	payloadBytes, _ := json.Marshal(validationPayload)
	resp, err = http.Post(validationURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to validate questions"})
		return
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	type ValidQuestion struct {
		ValidQuestionMap map[string]bool `json:"valid_question_map"`
	}
	var apiResp ValidQuestion
	err = json.Unmarshal(bodyBytes, &apiResp)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to parse validation response"})
		return
	}
	validQuestionMap := apiResp.ValidQuestionMap
	//var validQuestions []models.Question
	//if err := tx.Where("experiment_id = ?", experimentID).
	//	Where("id IN ?", validQuestionIDs).
	//	Find(&validQuestions).Error; err != nil {
	//	tx.Rollback()
	//	c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to validate questions"})
	//	return
	//}
	//
	//validQuestionMap := make(map[string]bool)
	//for _, q := range validQuestions {
	//	validQuestionMap[q.ID] = true
	//}

	for _, ans := range req.Answers {
		if _, exists := validQuestionMap[ans.QuestionID]; !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("Question %s does not belong to experiment %s", ans.QuestionID, experimentID),
			})
			return
		}

		//localhost:8082/api/experiments/questionDetail
		questionDetailURL := fmt.Sprintf("%s/api/experiments/questionDetail", cfg.ExperimentServiceURL)
		detailPayload := gin.H{
			"question_id": ans.QuestionID,
		}
		payloadBytes, _ := json.Marshal(detailPayload)
		resp, err = http.Post(questionDetailURL, "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch question details"})
			return
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("questionDetail service error: %s", string(bodyBytes))})
			return
		}
		type response struct {
			Question Question
		}
		var res response
		err = json.Unmarshal(bodyBytes, &res)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to parse question details"})
			return
		}
		question := res.Question
		// 验证题目存在
		if question.ID == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("Question %s not found", ans.QuestionID),
			})
			return
		}
		//var question models.Question
		//if err := tx.Where("id = ?", ans.QuestionID).First(&question).Error; err != nil {
		//	tx.Rollback()
		//	c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to validate questions"})
		//	return
		//}

		var qSubmission models.QuestionSubmission
		result := tx.Where("submission_id = ? AND question_id = ?", submission.ID, ans.QuestionID).
			First(&qSubmission)

		if result.Error == nil {
			// 更新现有记录
			switch ans.Type {
			case "choice", "blank":
				qSubmission.Answer = ans.Answer
				qSubmission.Code = ""
				qSubmission.Language = ""
			case "code":
				qSubmission.Code = ans.Code
				qSubmission.Language = ans.Language
				qSubmission.Answer = ""
			}
			qSubmission.UpdatedAt = now

			if err := tx.Save(&qSubmission).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Failed to update answer for question %s", ans.QuestionID),
				})
				return
			}
		} else {
			// 创建新记录
			qSubmission = models.QuestionSubmission{
				ID:           uuid.New().String(),
				SubmissionID: submission.ID,
				QuestionID:   ans.QuestionID,
				Type:         question.Type,
				PerfectScore: question.Score,
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			switch ans.Type {
			case "choice", "blank":
				qSubmission.Answer = ans.Answer
			case "code":
				qSubmission.Code = ans.Code
				qSubmission.Language = ans.Language
			}

			if err := tx.Create(&qSubmission).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Failed to save answer for question %s", ans.QuestionID),
				})
				return
			}
		}
	}

	// 4. 获取所有已保存题目（包括之前保存的）
	var allSubmissions []models.QuestionSubmission
	if err := tx.Where("submission_id = ?", submission.ID).Find(&allSubmissions).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch saved questions"})
		return
	}

	fullSavedQuestions := make([]gin.H, 0, len(allSubmissions))
	for _, s := range allSubmissions {
		fullSavedQuestions = append(fullSavedQuestions, gin.H{
			"question_id":            s.QuestionID,
			"question_submission_id": s.ID,
			"updated_at":             s.UpdatedAt.Format(time.RFC3339),
		})
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Transaction failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"submission_id":   submission.ID,
			"student_id":      studentID,
			"experiment_id":   experimentID,
			"updated_at":      now,
			"saved_questions": fullSavedQuestions,
		},
	})
}

type Question struct {
	ID            string `json:"id"`
	ExperimentID  string `json:"experiment_id"`
	Type          string `json:"type"` // choice, blank, code
	Content       string `json:"content"`
	Options       string `json:"options,omitempty" ` // JSON 字符串存储选择题选项
	CorrectAnswer string `json:"correct_answer,omitempty"`
	Score         int    `json:"score"`

	ImageURL    string    `json:"image_url,omitempty"`
	TestCases   string    `json:"test_cases,omitempty"` // JSON 字符串存储代码题的测试用例
	Explanation string    `json:"explanation,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type User struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Telephone string
	Password  string
	Role      string
	AvatarUrl string
	Email     string
}

type Experiment struct {
	ID         string    `json:"experiment_id"`
	Permission int       `json:"permission"`
	Deadline   time.Time `json:"deadline"`
	IsExpired  bool      `json:"is_expired"`
	TotalScore int       `json:"total_score"`
	Title      string    `json:"title"`
}

func SubmitExperiment(c *gin.Context) {
	db := global.DB
	ID, _ := strconv.ParseUint(c.GetHeader("X-User-ID"), 10, 32)
	studentID := uint(ID)
	experimentID := c.Param("experiment_id")
	cfg := config.LoadConfig()
	// 1. 检查实验是否已过期
	//localhost:8082/api/experiment/experimentDetail
	experimentURL := fmt.Sprintf("%s/api/experiments/experimentDetail", cfg.ExperimentServiceURL)
	payload := gin.H{
		"experiment_id": experimentID,
	}
	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(experimentURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch experiment details"})
		return
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("experimentDetail service error: %s", string(bodyBytes))})
		return
	}
	type GetExperimentDetailResponse struct {
		ExperimentID string    `json:"experiment_id"`
		Permission   int       `json:"permission"`
		Deadline     time.Time `json:"deadline"`
		Title        string    `json:"title"`
		IsExpired    bool      `json:"is_expired"`
		TotalScore   int       `json:"total_score"`
	}
	var experiment GetExperimentDetailResponse
	err = json.Unmarshal(bodyBytes, &experiment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to parse experiment details"})
		return
	}
	// 验证实验存在
	if experiment.ExperimentID == "" {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Experiment not found"})
		return
	}
	//if err := db.Preload("Questions").First(&experiment, "id = ?", experimentID).Error; err != nil {
	//	c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Experiment not found"})
	//	return
	//}
	totalPerfectScore := experiment.TotalScore
	//for _, q := range experiment.Questions {
	//	totalPerfectScore += q.Score
	//
	//}
	if experiment.IsExpired {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Experiment deadline has passed"})
		return
	}
	// 2. 解析请求体中的答案
	var req struct {
		Answers []struct {
			QuestionID string `json:"question_id"`
			Type       string `json:"type"`
			Answer     string `json:"answer,omitempty"`
			Code       string `json:"code,omitempty"`
			Language   string `json:"language,omitempty"`
		} `json:"answers"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}
	now := time.Now()
	// 处理实验提交记录（保持不变）
	var submission models.ExperimentSubmission
	result := db.Where("experiment_id = ? AND student_id = ? AND status != 'submitted'", experimentID, studentID).Order("created_at DESC").
		First(&submission)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		submission = models.ExperimentSubmission{
			ID:           uuid.New().String(),
			ExperimentID: experimentID,
			StudentID:    studentID,
			Status:       "in_progress",
			CreatedAt:    now,
			UpdatedAt:    now,
			SubmittedAt:  now,
		}
		if err := db.Create(&submission).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create submission"})
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Database error"})
		return
	} else {
		submission.UpdatedAt = now
		if err := db.Save(&submission).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update submission"})
			return
		}
	}
	tx := db.Begin()
	// 3. 处理每道题的提交
	totalScore := 0
	results := make([]gin.H, 0, len(req.Answers))
	validQuestionIDs := make([]string, len(req.Answers))
	for i, ans := range req.Answers {
		validQuestionIDs[i] = ans.QuestionID
	}

	// 验证题目属于当前实验
	//localhost:8082/api/experiments/validate-questions
	validationURL := fmt.Sprintf("%s/api/experiments/validateQuestion", cfg.ExperimentServiceURL)
	validationPayload := gin.H{
		"experiment_id": experimentID,
		"question_ids":  validQuestionIDs,
	}
	payloadBytes, _ = json.Marshal(validationPayload)
	resp, err = http.Post(validationURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to validate questions"})
		return
	}
	defer resp.Body.Close()
	//var validQuestions []models.Question
	//if err := tx.Where("experiment_id = ?", experimentID).
	//	Where("id IN ?", validQuestionIDs).
	//	Find(&validQuestions).Error; err != nil {
	//	tx.Rollback()
	//	c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to validate questions"})
	//	return
	//}

	//validQuestionMap := make(map[string]Question, len(validQuestions))
	//for _, q := range validQuestions {
	//	validQuestionMap[q.ID] = q
	//}

	for _, ans := range req.Answers {
		//localhost:8082/api/experiments/questionDetail
		questionDetailURL := fmt.Sprintf("%s/api/experiments/questionDetail", cfg.ExperimentServiceURL)
		detailPayload := gin.H{
			"question_id": ans.QuestionID,
		}
		payloadBytes, _ := json.Marshal(detailPayload)
		resp, err = http.Post(questionDetailURL, "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch question details"})
			return
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("questionDetail service error: %s", string(bodyBytes))})
			return
		}

		var question Question
		err = json.Unmarshal(bodyBytes, &question)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to parse question details"})
			return
		}
		var qSubmission models.QuestionSubmission
		result := tx.Where("submission_id = ? AND question_id = ?", submission.ID, ans.QuestionID).
			First(&qSubmission)

		if result.Error == nil {
			// 更新现有记录
			switch ans.Type {
			case "choice", "blank":
				qSubmission.Answer = ans.Answer
				qSubmission.Code = ""
				qSubmission.Language = ""
			case "code":
				qSubmission.Code = ans.Code
				qSubmission.Language = ans.Language
				qSubmission.Answer = ""
			}
			qSubmission.UpdatedAt = now
			qSubmission.Score, qSubmission.Feedback = getScore(question, ans)
			totalScore += qSubmission.Score
			results = append(results, gin.H{
				"question_id": ans.QuestionID,
				"type":        question.Type,
				"score":       fmt.Sprintf("%d/%d", qSubmission.Score, question.Score),
				"feedback":    qSubmission.Feedback,
			})
			if err := tx.Save(&qSubmission).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Failed to update answer for question %s", ans.QuestionID),
				})
				return
			}
		} else {
			// 创建新记录
			qSubmission = models.QuestionSubmission{
				ID:           uuid.New().String(),
				SubmissionID: submission.ID,
				QuestionID:   ans.QuestionID,
				Type:         question.Type,
				PerfectScore: question.Score,
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			switch ans.Type {
			case "choice", "blank":
				qSubmission.Answer = ans.Answer
			case "code":
				qSubmission.Code = ans.Code
				qSubmission.Language = ans.Language
			}
			qSubmission.Score, qSubmission.Feedback = getScore(question, ans)
			totalScore += qSubmission.Score
			results = append(results, gin.H{
				"question_id": ans.QuestionID,
				"type":        question.Type,
				"score":       fmt.Sprintf("%d/%d", qSubmission.Score, question.Score),
				"feedback":    qSubmission.Feedback,
			})
			if err := tx.Create(&qSubmission).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Failed to save answer for question %s", ans.QuestionID),
				})
				return
			}
		}
	}
	// 5. 更新实验提交记录的总分
	submission.TotalScore = totalScore
	submission.Status = "submitted"
	if err := tx.Save(&submission).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to save submission",
		})
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"submission_id": submission.ID,
			"total_score":   fmt.Sprintf("%d/%d", totalScore, totalPerfectScore),
			"results":       results,
			"submitted_at":  submission.SubmittedAt,
		},
	})
}

func GetSubmissions(c *gin.Context) {
	db := global.DB
	ID, _ := strconv.ParseUint(c.GetHeader("X-User-ID"), 10, 32)
	studentID := uint(ID)
	cfg := config.LoadConfig()

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// 查询条件
	query := db.Model(&models.ExperimentSubmission{}).Where("student_id = ?", studentID)

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取提交记录
	var submissions []models.ExperimentSubmission
	err := query.
		Offset(offset).
		Limit(limit).
		Order("submitted_at DESC").
		Find(&submissions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Database query failed"})
		return
	}

	// 获取所有相关的问题提交
	submissionIDs := make([]string, len(submissions))
	for i, sub := range submissions {
		submissionIDs[i] = sub.ID
	}

	var questionSubmissions []models.QuestionSubmission
	if len(submissionIDs) > 0 {
		if err := db.Where("submission_id IN ?", submissionIDs).
			Find(&questionSubmissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to get question submissions"})
			return
		}
	}

	// 按提交ID分组问题提交
	questionSubMap := make(map[string][]models.QuestionSubmission)
	for _, qs := range questionSubmissions {
		questionSubMap[qs.SubmissionID] = append(questionSubMap[qs.SubmissionID], qs)
	}
	now := time.Now()
	// 构建响应
	submissionResponses := make([]gin.H, len(submissions))
	for i, sub := range submissions {
		experimentID := sub.ExperimentID
		//获取Experiment
		//localhost:8082/api/experiment/experimentDetail
		experimentURL := fmt.Sprintf("%s/api/experiments/experimentDetail", cfg.ExperimentServiceURL)
		payload := gin.H{
			"experiment_id": experimentID,
		}
		payloadBytes, _ := json.Marshal(payload)
		resp, err := http.Post(experimentURL, "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch experiment details"})
			return
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("experimentDetail service error: %s", string(bodyBytes))})
			return
		}
		type GetExperimentDetailResponse struct {
			ExperimentID string    `json:"experiment_id"`
			Permission   int       `json:"permission"`
			Deadline     time.Time `json:"deadline"`
			Title        string    `json:"title"`
			IsExpired    bool      `json:"is_expired"`
			TotalScore   int       `json:"total_score"`
		}
		var experiment GetExperimentDetailResponse
		err = json.Unmarshal(bodyBytes, &experiment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to parse experiment details"})
			return
		}

		// 获取该提交的问题结果
		results := make([]gin.H, 0)
		if qSubs, ok := questionSubMap[sub.ID]; ok {
			for _, qs := range qSubs {
				//获取Question
				//localhost:8082/api/experiments/questionDetail
				questionDetailURL := fmt.Sprintf("%s/api/experiments/questionDetail", cfg.ExperimentServiceURL)
				detailPayload := gin.H{
					"question_id": qs.QuestionID,
				}
				payloadBytes, _ := json.Marshal(detailPayload)
				resp, err = http.Post(questionDetailURL, "application/json", bytes.NewBuffer(payloadBytes))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch question details"})
					return
				}
				defer resp.Body.Close()
				bodyBytes, _ := io.ReadAll(resp.Body)
				if resp.StatusCode != http.StatusOK {
					c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("questionDetail service error: %s", string(bodyBytes))})
					return
				}
				type GetQuestionDetailResponse struct {
					Question Question
				}
				var que GetQuestionDetailResponse
				err = json.Unmarshal(bodyBytes, &que)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to parse question details"})
					return
				}
				question := que.Question
				explanation := ""
				if now.After(experiment.Deadline) {
					explanation = question.Explanation
				}
				result := gin.H{
					"question_id": qs.QuestionID,
					"type":        question.Type,
					"score":       qs.Score,
					"feedback":    qs.Feedback,
					"explanation": explanation,
				}
				results = append(results, result)
			}
		}

		submissionResponses[i] = gin.H{
			"submission_id":    sub.ID,
			"experiment_id":    sub.ExperimentID,
			"experiment_title": experiment.Title,
			"total_score":      sub.TotalScore,
			"status":           sub.Status,
			"submitted_at":     sub.SubmittedAt.Format(time.RFC3339),
			"results":          results,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   submissionResponses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func GetSubmissionStatus(c *gin.Context) {
	// 查询学生的提交状态
	db := global.DB
	experimentID := c.Param("experiment_id")
	studentID := c.Param("student_id")
	var submission models.ExperimentSubmission
	submissionStatus := "not_started"
	if err := db.Where("experiment_id = ? AND student_id = ?", experimentID, studentID).Order("created_at DESC").
		First(&submission).Error; err == nil {
		submissionStatus = strings.ToLower(submission.Status)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Database error"})
		return

	}
	// 返回提交状态
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"submission_id":     submission.ID,
			"experiment_id":     experimentID,
			"student_id":        studentID,
			"submission_status": submissionStatus,
			"submitted_at":      submission.SubmittedAt,
			"total_score":       submission.TotalScore,
			"updated_at":        submission.UpdatedAt,
			"created_at":        submission.CreatedAt,
		},
	})
}
func GetStudentAns(c *gin.Context) {
	// 获取学生答案和反馈
	var qSubmission models.QuestionSubmission
	type Req struct {
		SubId string `json:"submission_id" binding:"required"`
		QId   string `json:"question_id" binding:"required"`
	}
	var q Req
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request"})
		return
	}

	if err := global.DB.Where("submission_id = ? AND question_id = ?", q.SubId, q.QId).
		First(&qSubmission).Error; err == nil {
	} else {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Submission or question not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   qSubmission,
	})
}

func UpdateExperimentStatusToInProgress(c *gin.Context) {
	db := global.DB
	experimentID := c.Param("experiment_id")
	//是否存在该实验
	var experiment models.ExperimentSubmission
	if err := db.Where("experiment_id = ?", experimentID).First(&experiment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":   "error",
			"message ": "Experiment not found"})
		return
	} else {

		// 更新所有该实验的提交状态为 "in_progress"
		if err := db.Model(&models.ExperimentSubmission{}).Where("experiment_id = ?", experimentID).
			Update("status", "in_progress").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":   "error",
				"message ": "Failed to update experiment status"})
			return
		}

		// 返回成功响应
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Experiment status updated to in_progress"})
	}
}

// DeleteExperimentSubmissions 删除实验相关的所有提交记录
func DeleteExperimentSubmissions(c *gin.Context) {
	db := global.DB
	experimentID := c.Param("experiment_id")

	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "实验ID不能为空",
		})
		return
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 删除关联的题目提交记录
	if err := tx.Where("submission_id IN (SELECT id FROM experiment_submissions WHERE experiment_id = ?)", experimentID).
		Delete(&models.QuestionSubmission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "删除题目提交记录失败: " + err.Error(),
		})
		return
	}

	// 2. 删除关联的实验提交记录
	if err := tx.Where("experiment_id = ?", experimentID).
		Delete(&models.ExperimentSubmission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "删除实验提交记录失败: " + err.Error(),
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "提交事务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "实验提交记录删除成功",
	})
}
