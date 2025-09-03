# SmartFox CI/CD - Self-hosted Runner 版本

本项目配置为使用 **self-hosted runner** 运行 GitHub Actions，提供更好的性能控制和定制化环境。

## 🚀 快速开始

### 对于 macOS 用户

1. **自动设置环境**（推荐）：
   ```bash
   ./scripts/setup-macos-runner.sh
   ```

2. **手动验证环境**：
   ```bash
   ./scripts/verify-runner-env-macos.sh
   ```

3. **设置 GitHub Actions Runner**：
   按照 [macOS 设置指南](docs/self-hosted-runner-setup-macos.md) 进行配置

### 对于其他操作系统

查看通用设置指南：[Self-hosted Runner 设置](docs/self-hosted-runner-setup.md)

## 📋 系统要求

### 最低要求
- **内存**: 8GB RAM
- **CPU**: 4 核心
- **存储**: 50GB 可用空间
- **网络**: 稳定的互联网连接

### 推荐配置
- **内存**: 16GB RAM
- **CPU**: 8 核心
- **存储**: 100GB SSD
- **网络**: 高速宽带连接

## 🛠️ 必需软件

### 核心依赖
- **Docker** (>= 20.10) + Docker Desktop
- **Go** (>= 1.19)
- **Node.js** (>= 16.x)
- **Python** (>= 3.8)
- **kubectl** (>= 1.25)
- **KinD** (Kubernetes in Docker)
- **jq** (JSON 处理工具)

### 可选工具
- **Newman** (Postman CLI 测试)
- **Helm** (Kubernetes 包管理)
- **yq** (YAML 处理工具)

## 🏗️ 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                    GitHub Repository                        │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   CI Workflow   │    │   CD Workflow   │                │
│  │   (on push)     │    │   (on tag)      │                │
│  └─────────────────┘    └─────────────────┘                │
└─────────────────┬───────────────┬─────────────────────────────┘
                  │               │
                  ▼               ▼
┌─────────────────────────────────────────────────────────────┐
│                Self-hosted Runner                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Docker    │  │   KinD      │  │   Tools     │        │
│  │   Images    │  │   Cluster   │  │   (Go, Node,│        │
│  │   Build     │  │   Deploy    │  │   Python)   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## 🔄 CI/CD 流程

### 持续集成 (CI)
触发条件：Push 到 `main` 分支或 Pull Request

1. **环境验证**: 检查所有必需工具
2. **代码检出**: 获取最新代码
3. **依赖安装**: Go、Node.js、Python 依赖
4. **代码构建**: 编译所有服务
5. **镜像构建**: 构建 Docker 镜像
6. **集群部署**: 使用 KinD 创建测试集群
7. **烟雾测试**: 基础功能测试
8. **E2E 测试**: 端到端测试
9. **Postman 测试**: API 集成测试

### 持续部署 (CD)
触发条件：推送版本标签 (如 `v1.0.0`)

1. **镜像构建**: 构建生产镜像
2. **镜像推送**: 推送到 GitHub Container Registry
3. **生产部署**: 部署到生产 Kubernetes 集群（如果配置了 KUBE_CONFIG）

## 📊 工作流状态

### CI 工作流
- ✅ **构建验证**: 所有服务编译通过
- ✅ **镜像构建**: Docker 镜像构建成功
- ✅ **本地部署**: KinD 集群部署测试
- ✅ **自动化测试**: 包括烟雾测试和 E2E 测试

### CD 工作流
- ✅ **版本发布**: 自动构建和推送镜像
- 🔧 **生产部署**: 需要配置 `KUBE_CONFIG` 密钥

## ⚙️ 配置说明

### GitHub Secrets

在仓库的 Settings > Secrets and variables > Actions 中配置：

| 密钥名称 | 描述 | 必需 |
|---------|------|------|
| `GITHUB_TOKEN` | 自动提供，用于推送镜像 | ✅ |
| `KUBE_CONFIG` | 生产 Kubernetes 配置 (base64 编码) | 🔧 |

### Runner 标签

支持的 runner 标签：
- `self-hosted`: 基础标签
- `macOS`: macOS 系统
- `Linux`: Linux 系统  
- `Windows`: Windows 系统
- `ARM64`: ARM 架构（Apple Silicon）
- `X64`: x86_64 架构

## 🐳 服务架构

### 微服务组件
- **Gateway**: API 网关 (端口 80)
- **User Service**: 用户管理 (端口 8081)
- **Experiment Service**: 实验管理 (端口 8082)
- **Notification Service**: 通知服务 (端口 8083)
- **Submission Service**: 提交管理 (端口 8084)
- **Judge Service**: 评判服务 (端口 8085)

### 前端应用
- **SmartFox Frontend**: React 应用 (端口 3000)

### 数据存储
- **MySQL**: 每个服务独立的数据库实例
- **持久化存储**: 使用 Kubernetes PVC

## 🔧 本地开发

### 启动开发环境
```bash
# 清理现有集群
kind delete cluster --name smartfox-ci

# 运行完整测试流程
cd mirco_service_fox
.github/scripts/kind-deploy-and-test.sh
```

### 单独构建服务
```bash
# 构建特定服务
cd mirco_service_fox/user-service
go build ./...

# 构建 Docker 镜像
docker build -t user-service:dev .
```

### 前端开发
```bash
cd smartfox_front
npm install
npm start
```

## 📈 性能监控

### 资源使用情况
- **内存监控**: `top -o MEM`
- **Docker 资源**: `docker stats`
- **磁盘使用**: `df -h`
- **网络状态**: `netstat -tuln`

### 日志查看
```bash
# Runner 日志
tail -f ~/actions-runner/_diag/Runner_*.log

# 作业日志  
tail -f ~/actions-runner/_diag/Worker_*.log

# Kubernetes 日志
kubectl logs -f deployment/user-service
```

## 🛠️ 故障排除

### 常见问题

1. **Docker 未启动**
   ```bash
   # macOS
   open /Applications/Docker.app
   
   # 验证
   docker info
   ```

2. **端口被占用**
   ```bash
   # 查看占用端口的进程
   lsof -i :8080
   
   # 终止进程
   sudo kill -9 <PID>
   ```

3. **内存不足**
   ```bash
   # 清理 Docker 缓存
   docker system prune -a -f
   
   # 清理 KinD 集群
   kind delete cluster --name smartfox-ci
   ```

4. **构建失败**
   ```bash
   # 清理 Go 缓存
   go clean -modcache
   
   # 清理 npm 缓存
   npm cache clean --force
   ```

### 获取帮助

- 📖 **文档**: 查看 `docs/` 目录中的详细指南
- 🐛 **问题报告**: 在 GitHub Issues 中提交问题
- 💬 **讨论**: 使用 GitHub Discussions 进行讨论

## 🚀 部署到生产环境

### 发布新版本
```bash
# 创建版本标签
git tag v1.0.0
git push origin v1.0.0

# CD 工作流将自动触发
```

### 生产环境配置
1. 准备 Kubernetes 集群
2. 配置 `KUBE_CONFIG` 密钥
3. 推送版本标签触发部署

## 📚 参考资料

- [Self-hosted Runner 设置指南](docs/self-hosted-runner-setup-macos.md)
- [GitHub Actions 官方文档](https://docs.github.com/en/actions)
- [Docker 官方文档](https://docs.docker.com/)
- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [KinD 用户指南](https://kind.sigs.k8s.io/)

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

⭐ 如果这个项目对您有帮助，请给它一个星标！
