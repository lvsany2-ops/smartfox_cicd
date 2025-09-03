package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONIntSlice 用于处理 JSON 整数切片
type JSONIntSlice []int

// Value 实现 driver.Valuer 接口
func (j JSONIntSlice) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONIntSlice) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan into JSONIntSlice")
	}

	return json.Unmarshal(bytes, j)
}

// Experiment 实验模型
type Experiment struct {
	ID          string    `json:"experiment_id" gorm:"primaryKey;type:char(36)"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FileURL     string    `json:"file_url,omitempty"`
	Permission  int       `json:"permission"`
	Deadline    time.Time `json:"deadline"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time
	Questions   []Question   `json:"questions" gorm:"foreignKey:ExperimentID"`
	Attachments []Attachment `json:"attachments" gorm:"foreignKey:ExperimentID"`
	UserIDs     JSONIntSlice `json:"user_ids" gorm:"type:json"`
}

// Question 题目模型
type Question struct {
	ID            string `json:"id" gorm:"primaryKey;type:char(36)"`
	ExperimentID  string `json:"experiment_id"`
	Type          string `json:"type"` // choice, blank, code
	Content       string `json:"content"`
	Options       string `json:"options,omitempty" gorm:"type:text"` // JSON 字符串存储选择题选项
	CorrectAnswer string `json:"correct_answer,omitempty"`
	Score         int    `json:"score"`

	ImageURL    string `json:"image_url,omitempty"`
	TestCases   string `json:"test_cases,omitempty"` // JSON 字符串存储代码题的测试用例
	Explanation string `json:"explanation,omitempty"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Attachment 附件模型
type Attachment struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	ExperimentID string `json:"experiment_id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TestCase 测试用例结构体
type TestCase struct {
	Input          interface{} `json:"input"`
	ExpectedOutput interface{} `json:"expected_output"`
}
