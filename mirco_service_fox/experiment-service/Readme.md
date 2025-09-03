# Experiment Service

Go 微服务，负责实验、题目、附件相关 RESTful API。

- 目录结构：models、controllers、routers、config、middleware、utils
- 连接 MySQL（experiment_db）
- 支持文件上传
- 题目结构抽象：选择题/填空题/代码题

## 启动
```sh
cd experiment_service
# 构建镜像
docker build -t experiment_service .
# 运行服务
docker run -p 8080:8080 experiment_service
```
