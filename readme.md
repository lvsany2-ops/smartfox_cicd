已新增/修改的文件

ci.yml
push/PR 触发
go build 验证（各 Go 服务）
docker build 镜像（所有服务）
启动 KinD 集群，安装存储类，部署所有 k8s 清单，等待就绪
运行 e2e 烟囱测试（注册→登录→鉴权→实验列表）和 Postman（newman）集合
cd.yml
tag vX.Y.Z 触发
登录 GHCR，构建并推送镜像：ghcr.io/<owner>/smartfox-<service>:<tag>
若配置 KUBE_CONFIG（base64 kubeconfig），自动应用 k8s 清单并回滚等待
自动将部署文件中的 imagePullPolicy: Never 改为 IfNotPresent（避免外网 pull 问题）
kind-deploy-and-test.sh
KinD 集群创建、镜像加载、k8s 清单应用、就绪等待、内部 curl 健康检查
smoke-e2e.sh
port-forward 网关；注册/登录获取 Bearer Token；调用 profile 与学生实验列表
run-newman.sh
port-forward 网关；生成临时 Postman 环境；运行基础集合
basic.postman_collection.json
健康检查、注册、登录、个人信息 4 个基本用例
experiment-service 细化
routers/routers.go：补充 /health

main.go：OSS 初始化改为受 ENABLE_OSS 控制，避免 CI 外部依赖






