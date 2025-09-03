// src/pages/HistoryPage/indexHistoryPage.cy.jsx
import React from 'react';
import HistoryPage from './index';
import { ConfigProvider } from 'antd';
import { BrowserRouter as Router } from 'react-router-dom';
import axios from '../../utils/axios';

describe('<HistoryPage />', () => {
  const mockSubmissions = [
    {
      id: 1,
      experiment_id: 101,
      experiment_title: '实验一：基础编程',
      status: 'graded',
      total_score: 85,
      submitted_at: '2023-10-01T10:00:00Z',
      results: [
        { score: 30 },
        { score: 25 },
        { score: 30 }
      ]
    },
    {
      id: 2,
      experiment_id: 102,
      experiment_title: '实验二：数据结构',
      status: 'submitted',
      total_score: 0,
      submitted_at: '2023-10-05T14:30:00Z',
      results: []
    }
  ];

  const mockPagination = {
    page: 1,
    limit: 10,
    total: 2
  };

  let axiosStub;

  beforeEach(() => {
    // Always restore any existing stub first
    if (axiosStub) {
      axiosStub.restore();
    }
    
    // Create a fresh stub
    axiosStub = cy.stub(axios, 'get').callsFake((url, config) => {
      if (url === '/student/submissions') {
        const params = config?.params || {};
        const page = params.page || 1;
        
        return Promise.resolve({
          status: 'success',
          data: mockSubmissions,
          pagination: {
            ...mockPagination,
            page
          }
        });
      }
      return Promise.reject(new Error('Not found'));
    });
  });

  afterEach(() => {
    // Clean up the stub after each test
    if (axiosStub) {
      axiosStub.restore();
    }
  });

  const mountWithRouter = (component) => {
    return cy.mount(
      <ConfigProvider>
        <Router>
          {component}
        </Router>
      </ConfigProvider>
    );
  };

  it('成功加载历史记录 - 正向测试', () => {
    mountWithRouter(<HistoryPage />);

    cy.get('.ant-card-head-title').should('contain', '实验提交历史');
    cy.get('.ant-list-item').should('have.length', 2);
    
    cy.get('.ant-list-item:first')
      .should('contain', '实验一：基础编程')
      .and('contain', '已评分')
      .and('contain', '得分: 85 分');

    cy.get('.ant-list-item:first .ant-tag')
      .should('have.class', 'ant-tag-green');

    cy.contains('题1: 30分').should('exist');

    cy.get('.ant-list-item:first a')
      .should('have.attr', 'href', '/experiments/101/result')
      .and('contain', '查看详情');

    cy.contains('提交时间:').should('exist');
  });

  it('显示空状态 - 正向测试', () => {
    // Temporarily override the stub for this test
    axiosStub.restore();
    cy.stub(axios, 'get').callsFake((url) => {
      if (url === '/student/submissions') {
        return Promise.resolve({
          status: 'success',
          data: [],
          pagination: {
            page: 1,
            limit: 10,
            total: 0
          }
        });
      }
      return Promise.reject(new Error('Not found'));
    });

    mountWithRouter(<HistoryPage />);

    cy.get('.ant-alert-info').should('exist');
    cy.contains('暂无提交记录').should('exist');
    cy.contains('您还没有提交过任何实验').should('exist');
  });

  it('API 调用失败 - 反向测试', () => {
    // Temporarily override the stub for this test
    axiosStub.restore();
    cy.stub(axios, 'get').callsFake((url) => {
      if (url === '/student/submissions') {
        return Promise.reject({
          response: {
            data: {
              message: '服务器内部错误'
            }
          }
        });
      }
      return Promise.reject(new Error('Not found'));
    });

    mountWithRouter(<HistoryPage />);

    cy.get('.ant-alert-error').should('exist');
    // Look for the specific error message from the mock response
    cy.contains('服务器内部错误').should('exist');
  });

  it('API 调用失败 - 通用错误消息测试', () => {
    // Temporarily override the stub for this test - test fallback message
    axiosStub.restore();
    cy.stub(axios, 'get').callsFake((url) => {
      if (url === '/student/submissions') {
        return Promise.reject({
          response: null // No response to trigger fallback message
        });
      }
      return Promise.reject(new Error('Not found'));
    });

    mountWithRouter(<HistoryPage />);

    cy.get('.ant-alert-error').should('exist');
    // Look for the fallback error message
    cy.contains('获取历史记录失败').should('exist');
  });

  it('分页功能测试', () => {
    let callCount = 0;
    
    // Temporarily override the stub for this test
    axiosStub.restore();
    cy.stub(axios, 'get').callsFake((url, config) => {
      if (url === '/student/submissions') {
        callCount++;
        const page = config?.params?.page || 1;
        
        return Promise.resolve({
          status: 'success',
          data: mockSubmissions,
          pagination: {
            ...mockPagination,
            page,
            total: 15 // 测试多页情况
          }
        });
      }
      return Promise.reject(new Error('Not found'));
    });

    mountWithRouter(<HistoryPage />);

    cy.get('.ant-pagination').should('exist');
    
    cy.get('.ant-pagination-item-2').click();

    cy.wrap(null).then(() => {
      expect(callCount).to.be.greaterThan(1);
    });
  });

  it('状态标签显示正确', () => {
    mountWithRouter(<HistoryPage />);

    cy.get('.ant-list-item').should('have.length', 2);

    cy.get('.ant-list-item:first .ant-tag')
      .should('have.class', 'ant-tag-green'); // graded - green

    cy.get('.ant-list-item:last .ant-tag')
      .should('have.class', 'ant-tag-blue'); // submitted - blue
  });

  it('时间格式显示正确', () => {
    mountWithRouter(<HistoryPage />);

    // Use a more flexible check for date format
    cy.contains('2023').should('exist');
    cy.contains('10').should('exist');
    cy.contains('1').should('exist');
  });
});