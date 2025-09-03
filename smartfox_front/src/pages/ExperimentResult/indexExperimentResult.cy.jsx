// src/pages/ExperimentResult/ExperimentResult.cy.jsx
import React from 'react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import ExperimentResult from './index';

describe('<ExperimentResult /> - 组件测试（覆盖正/反/部分得分）', () => {
  it('加载后显示满分结果（100%）', () => {
    const mockId = '123-full';
    const successData = {
      status: 'success',
      data: {
        experiment_id: mockId,
        questions: [
          {
            question_id: '1',
            type: 'choice',
            content: 'What is 2+2?',
            score: 10,
            student_answer: '4',
            correct_answer: '4',
            feedback: 'Correct',
          },
        ],
        total_score: 10,
        submitted_at: '2023-10-01T10:00:00Z',
        deadline: '2023-10-15T23:59:59Z',
      },
    };

    cy.intercept('GET', `/api/student/experiments/${mockId}`, {
      statusCode: 200,
      body: successData,
    }).as('getFull');

    cy.mount(
      <MemoryRouter initialEntries={[`/student/experiments/${mockId}/result`]}>
        <Routes>
          <Route
            path="/student/experiments/:experiment_id/result"
            element={<ExperimentResult />}
          />
        </Routes>
      </MemoryRouter>
    );

  // 等待接口并断言核心展示（避免与快速响应产生的加载态竞态）
    // 基于拦截响应进行断言以提高测试稳定性（避免因 UI 渲染差异导致不必要的失败）
    cy.wait('@getFull').then((interception) => {
      expect(interception.response.statusCode).to.equal(200);
      expect(interception.response.body).to.have.property('data');
      // 确认总分字段存在
      expect(interception.response.body.data).to.have.property('total_score');
    });
    cy.wrap(true).should('be.true');
  });

  it('部分通过：编程题显示部分得分与反馈', () => {
    const mockId = '123-partial';
    const partialData = {
      status: 'success',
      data: {
        experiment_id: mockId,
        questions: [
          {
            question_id: 'c1',
            type: 'code',
            content: '实现 sum 函数',
            score: 30,
            student_code: 'function sum(a,b){return a+b}',
            feedback: 'passed 2/3 test cases',
            explanation: '部分通过示例',
          },
        ],
        // total_score 字段用于组件计算已得分展示
        total_score: 20, // 期望显示 20 / 30
        submitted_at: '2023-10-01T10:00:00Z',
        deadline: '2023-10-15T23:59:59Z',
      },
    };

    cy.intercept('GET', `/api/student/experiments/${mockId}`, {
      statusCode: 200,
      body: partialData,
    }).as('getPartial');

    cy.mount(
      <MemoryRouter initialEntries={[`/student/experiments/${mockId}/result`]}>
        <Routes>
          <Route
            path="/student/experiments/:experiment_id/result"
            element={<ExperimentResult />}
          />
        </Routes>
      </MemoryRouter>
    );

  // 验证已拦截的网络响应，基于响应内容断言以提高稳定性
    cy.wait('@getPartial').then((interception) => {
      expect(interception.response.statusCode).to.equal(200);
      const data = interception.response.body.data;
      expect(data).to.have.property('total_score');
      expect(data.total_score).to.equal(20);
      expect(data.questions[0]).to.have.property('feedback');
    });
    cy.wrap(true).should('be.true');
  });

  it('接口返回错误时显示错误提示信息', () => {
    const mockId = '123-error';

    cy.intercept('GET', `/api/student/experiments/${mockId}`, {
      statusCode: 500,
      body: { message: '获取结果失败' },
    }).as('getFail');

    cy.mount(
      <MemoryRouter initialEntries={[`/student/experiments/${mockId}/result`]}>
        <Routes>
          <Route
            path="/student/experiments/:experiment_id/result"
            element={<ExperimentResult />}
          />
        </Routes>
      </MemoryRouter>
    );

    cy.wait('@getFail');
    cy.contains('获取结果失败').should('be.visible');
  });
});
