// /src/pages/StudentGroupManagement/__tests__/StudentGroupManagement.test.jsx
import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import StudentGroupManagement from "./index"; // 引用 src/pages/StudentGroupManagement/index.jsx
import { studentGroupAPI } from "../../utils/api";
import { message } from "antd";

// mock antd message
jest.mock("antd", () => {
  const antd = jest.requireActual("antd");
  return {
    ...antd,
    message: {
      success: jest.fn(),
      error: jest.fn(),
    },
  };
});

// mock API
jest.mock("../../utils/api", () => ({
  studentGroupAPI: {
    getStudents: jest.fn(() =>
      Promise.resolve({
        data: [{ user_id: "1", username: "小明", group_ids: [] }],
        pagination: { page: 1, limit: 10, total: 1 },
      })
    ),
    getGroups: jest.fn(() =>
      Promise.resolve({
        data: [],
        pagination: { page: 1, limit: 10, total: 0 },
      })
    ),
    createGroup: jest.fn(),
    updateGroup: jest.fn(),
    deleteGroup: jest.fn(),
  },
}));

describe("StudentGroupManagement", () => {
  beforeEach(async () => {
    jest.clearAllMocks();
    render(<StudentGroupManagement />);
    // 等待初始数据加载
    await waitFor(() => {
      expect(studentGroupAPI.getStudents).toHaveBeenCalled();
      expect(studentGroupAPI.getGroups).toHaveBeenCalled();
    });
  });

  test("正例：成功创建分组", async () => {
    studentGroupAPI.createGroup.mockResolvedValueOnce({ success: true });

    // 打开创建分组模态框
    fireEvent.click(screen.getByText("创建分组"));

    // 输入分组名称
    const input = screen.getByPlaceholderText("请输入分组名称");
    await userEvent.type(input, "测试分组");

    // 点击确定
    fireEvent.click(screen.getByText("创建"));

    await waitFor(() => {
      expect(studentGroupAPI.createGroup).toHaveBeenCalledWith({
        group_name: "测试分组",
        student_ids: [],
      });
      expect(message.success).toHaveBeenCalledWith("分组创建成功");
    });
  });

  test("反例：创建分组失败", async () => {
    studentGroupAPI.createGroup.mockRejectedValueOnce(new Error("后端异常"));

    fireEvent.click(screen.getByText("创建分组"));
    const input = screen.getByPlaceholderText("请输入分组名称");
    await userEvent.type(input, "失败分组");

    fireEvent.click(screen.getByText("创建"));

    await waitFor(() => {
      expect(studentGroupAPI.createGroup).toHaveBeenCalled();
      // 因为失败时只 console.error，没有 message.error
      // 所以我们检查 Modal 依然存在
      expect(screen.getByText("创建新分组")).toBeInTheDocument();
    });
  });
});
