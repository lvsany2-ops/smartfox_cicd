// src/pages/ExperimentResult/index.jsx
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Spin, Alert, Button, Tag, Progress } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import axios from '../../utils/axios';
import styles from './ExperimentResult.module.css';

export default function ExperimentResult() {
  const { experiment_id } = useParams();
  const navigate = useNavigate();
  const [experiment, setExperiment] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  // 修改mock数据，使其结构与API响应一致
  const mockData = {
    status: 'success',
    data: {
      experiment_id: '123',
      questions: [
        {
          question_id: '1',
          type: 'choice',
          content: 'What is 2+2?',
          score: 10,
          student_answer: '4',
          correct_answer: '4',
          feedback: 'Correct',
          explanation: 'Basic arithmetic'
        },
        {
          question_id: '2',
          type: 'code',
          content: 'Write a function that returns the sum of two numbers',
          score: 20,
          student_code: 'function sum(a, b) { return a + b; }',
          feedback: 'passed 3/3 test cases',
          explanation: 'Simple addition function'
        }
      ],
      total_score: 30,
      submitted_at: '2023-10-01T10:00:00Z',
      deadline: '2023-10-15T23:59:59Z'
    }
  };

  useEffect(() => {
    const fetchExperimentResult = async () => {
      try {
        // 获取实验详情（包含提交结果）
        const response = await axios.get(`/student/experiments/${experiment_id}`);
        // 模拟API响应
        // const response = { data: mockData };
  
        if (response.data.status === 'success') {
          setExperiment(response.data.data);
        }
      } catch (err) {
        setError(err.response?.data?.message || '获取结果失败');
      } finally {
        setLoading(false);
      }
    };
    fetchExperimentResult();
  }, [experiment_id]);

  const calculateActualScore = (question) => {
    const { score, feedback } = question;
    if (!feedback || !score) return 0;

    if (feedback.includes('Correct')) {
      return score;
    }
    if (feedback.includes('Incorrect')) {
      return 0;
    }
    const testCaseMatch = feedback.match(/passed (\d+)\/(\d+) test cases/i);
    if (testCaseMatch) {
      const passed = parseInt(testCaseMatch[1], 10);
      const total = parseInt(testCaseMatch[2], 10);
      if (total > 0) {
        return Math.round((passed / total) * score);
      }
    }
    return 0;
  };

  // 计算总分和得分率
  const calculateScore = () => {
    if (!experiment?.questions) return { total: 0, earned: 0, percentage: 0 };

    const total = experiment.questions.reduce((sum, q) => sum + q.score, 0);
    const earned = experiment.total_score || 0;
    const percentage = total > 0 ? Math.round((earned / total) * 100) : 0;

    return { total, earned, percentage };
  };

  // 渲染题目结果
  const renderQuestionResult = (question, index) => {
    const hasResult = question.feedback || question.explanation;
    const actualScore = calculateActualScore(question);
    const isCorrect = question.feedback && (question.feedback.includes('Correct') || actualScore === question.score);

    return (
      <Card key={question.question_id} className={styles.questionCard}>
        <div className={styles.questionHeader}>
          <div className={styles.questionTitle}>
            <span className={styles.questionNumber}>题目 {index + 1}</span>
            <span className={styles.questionType}>
              {question.type === 'choice' ? '选择题' :
               question.type === 'blank' ? '填空题' : '编程题'}
            </span>
          </div>
          <div className={styles.scoreInfo}>
            <span className={styles.score}>
              {actualScore || 0} / {question.score} 分
            </span>
            {hasResult && (
              <Tag color={isCorrect ? 'green' : 'red'} icon={isCorrect ? <CheckCircleOutlined /> : <CloseCircleOutlined />}>
                {isCorrect ? '正确' : '错误'}
              </Tag>
            )}
          </div>
        </div>

        <div className={styles.questionContent}>
          <p><strong>题目:</strong> {question.content}</p>

          {/* 显示学生答案 */}
          <div className={styles.studentAnswer}>
            <strong>你的答案:</strong>
            {question.type === 'code' ? (
              <pre className={styles.codeBlock}>
                {question.student_code || '未作答'}
              </pre>
            ) : (
              <span className={styles.answerText}>
                {question.student_answer || '未作答'}
              </span>
            )}
          </div>

          {/* 显示正确答案（如果有） */}
          {question.correct_answer && (
            <div className={styles.correctAnswer}>
              <strong>正确答案:</strong>
              <span className={styles.answerText}>{question.correct_answer}</span>
            </div>
          )}

          {/* 显示反馈 */}
          {question.feedback && (
            <div className={styles.feedback}>
              <strong>评测反馈:</strong>
              <p>{question.feedback}</p>
            </div>
          )}

          {/* 显示解释 */}
          {question.explanation && (
            <div className={styles.explanation}>
              <strong>题目解释:</strong>
              <p>{question.explanation}</p>
            </div>
          )}
        </div>
      </Card>
    );
  };

  if (loading) {
    return (
      <div className={styles.loading}>
        <Spin size="large" tip="加载评测结果..." />
      </div>
    );
  }

  if (error) {
    return (
      <div className={styles.error}>
        <Alert message={error} type="error" showIcon />
      </div>
    );
  }

  if (!experiment) {
    return (
      <div className={styles.error}>
        <Alert message="未找到实验结果" type="warning" showIcon />
      </div>
    );
  }

  const { total, earned, percentage } = calculateScore();

  return (
    <div className={styles.container}>

      {/* 总分概览 */}
      <Card className={styles.summaryCard}>
        <div className={styles.scoreOverview}>
          <div className={styles.scoreDisplay}>
            <div className={styles.mainScore}>
              <span className={styles.earned}>{earned}</span>
              <span className={styles.divider}>/</span>
              <span className={styles.total}>{total}</span>
            </div>
            <div className={styles.percentage}>{percentage}%</div>
          </div>
          <div className={styles.progressSection}>
            <Progress
              percent={percentage}
              strokeColor={percentage >= 60 ? '#52c41a' : percentage >= 40 ? '#faad14' : '#ff4d4f'}
              showInfo={false}
            />
            <div className={styles.progressLabel}>
              {percentage >= 80 ? '优秀' : percentage >= 60 ? '良好' : percentage >= 40 ? '及格' : '不及格'}
            </div>
          </div>
        </div>

        <div className={styles.experimentInfo}>
          <div>提交时间: {experiment.submitted_at ? new Date(experiment.submitted_at).toLocaleString() : '未提交'}</div>
          <div>截止时间: {new Date(experiment.deadline).toLocaleString()}</div>
        </div>
      </Card>

      {/* 题目详细结果 */}
      <div className={styles.questionsSection}>
        <h2>题目详情</h2>
        {experiment.questions?.map((question, index) => renderQuestionResult(question, index))}
      </div>
    </div>
  );
}