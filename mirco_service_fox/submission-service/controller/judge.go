package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"submission/config"
)

type TestCase struct {
	Input          interface{} `json:"input"`
	ExpectedOutput interface{} `json:"expected_output"`
}
type EvaluationRequest struct {
	Language   string     `json:"language"`
	SourceCode string     `json:"source_code"`
	TestCases  []TestCase `json:"test_cases"`
	TimeLimit  int        `json:"time_limit,omitempty"`
}

type EvaluationResponse struct {
	CaseResults []struct {
		Status    string  `json:"status"`
		Stdout    string  `json:"stdout"`
		Stderr    string  `json:"stderr"`
		TimeTaken float64 `json:"time_taken"`
	} `json:"case_results"`
	Summary struct {
		TotalCases  int     `json:"total_cases"`
		PassedCases int     `json:"passed_cases"`
		PassRate    float64 `json:"pass_rate_percent"`
		Status      string  `json:"overall_status"`
	} `json:"summary"`
}

func getScore(question Question, ans struct {
	QuestionID string "json:\"question_id\""
	Type       string "json:\"type\""
	Answer     string "json:\"answer,omitempty\""
	Code       string "json:\"code,omitempty\""
	Language   string "json:\"language,omitempty\""
}) (int, string) {
	score := 0
	feedback := ""
	switch question.Type {
	case "choice", "blank":
		if ans.Answer == question.CorrectAnswer {
			score = question.Score
			feedback = "Correct"
		} else {
			feedback = "Incorrect"
		}
	case "code":
		// 调用评测服务进行代码评测
		result, err := evaluateCode(ans.Code, ans.Language, question.TestCases)
		if err != nil {
			feedback = fmt.Sprintf("Evaluation error: %v", err)
		} else {
			score = int(float64(question.Score) * result.Summary.PassRate / 100)
			feedback = fmt.Sprintf("Passed %d/%d test cases", result.Summary.PassedCases, result.Summary.TotalCases)
		}
	}
	return score, feedback
}

// evaluateCode 调用评测服务进行代码评测
func evaluateCode(code, language, testCasesJSON string) (*EvaluationResponse, error) {
	// 解析测试用例
	var testCases []TestCase
	if err := json.Unmarshal([]byte(testCasesJSON), &testCases); err != nil {
		return nil, fmt.Errorf("invalid test cases format")
	}

	// 准备评测请求
	request := EvaluationRequest{
		Language:   language,
		SourceCode: code,
		TestCases:  testCases,
		TimeLimit:  2, // 默认2秒超时
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	cfg := config.LoadConfig()
	// 调用评测服务
	judgeURL := cfg.JudgeServiceURL + "/evaluate"
	resp, err := http.Post(judgeURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("evaluation service returned status: %d", resp.StatusCode)
	}

	// 解析响应
	var response EvaluationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}
