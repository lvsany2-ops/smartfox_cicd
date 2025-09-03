// jest.config.js
module.exports = {
  testEnvironment: "jsdom", // 模拟浏览器环境
  setupFilesAfterEnv: ["<rootDir>/src/setupTests.js"], // 每次运行测试前执行的配置文件
  moduleFileExtensions: ["js", "jsx", "json"],
  transform: {
    "^.+\\.(js|jsx)$": "babel-jest", // 使用 babel 处理 ES6/JSX
  },
  moduleNameMapper: {
    // 处理静态资源（防止导入图片/CSS时报错）
    "\\.(css|less|scss|sass)$": "identity-obj-proxy",
    "\\.(gif|ttf|eot|svg|png)$": "<rootDir>/__mocks__/fileMock.js",
  },
  testPathIgnorePatterns: ["/node_modules/", "/dist/"], // 忽略目录
};
