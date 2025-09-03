# 网关编写完成
## 测试方法：
### 1. 本地测试
在两个服务下分别执行go run main.go，即可进行localhost:8080/8081/8083测试
### 2. 容器测试
1. docker network create app-network,构建两个服务的通信网络
2. 在gateway notification-service文件夹下分别执行docker-compose up --build
理论效果同本地测试，均应为8083输出结果，8080进行鉴权和转发
### 3. k8s测试
#### 由于我的k8s环境为minikube，在docker desktop上的效果未知
1. 在gateway notification-service user-service文件夹下分别执行docker-compose up --build
2. minikube start --driver=docker
3. 导入镜像
- minikube image load mysql:8.0
- minikube image load gateway:latest
- minikube image load notification-service:latest
- minikube image load user-service:latest
4. 通过minikube ssh docker images检验上述三个镜像是否成功加载
5. 进入gateway/k8s，执行
- kubectl apply -f gateway-secrets.yaml
- kubectl apply -f gateway-config.yaml
- kubectl apply -f gateway-deployment.yaml
- kubectl apply -f gateway-service.yaml
6. 进入notification_service/k8s，执行
- kubectl apply -f notification-db-secret.yaml
- kubectl apply -f notification-mysql-statefulset.yaml
- kubectl apply -f notification-deployment.yaml
- kubectl apply -f notification-service.yaml
7. 等待10s后运行kubectl get pods，观察六个ready是否都是1/1，如果部分不是，继续等待数秒后再尝试kubectl get pods
8. 确认加载完成后，执行kubectl port-forward svc/gateway-service 8080:80
9. 此时就可以通过curl或postman确认转发是否正常
- curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTY3OTc1MzgsInJvbGUiOiJzdHVkZW50IiwidXNlcl9pZCI6MTIzfQ.r7N48myzfuY_KFvunNcoZXbYrnkZFmPRb5pytK-jyzg" http://localhost:8080/api/student/experiments/notifications/1
- 期望输出：{"data":[],"pagination":{"limit":10,"page":1,"total":0},"status":"success"}
### 上述过程至少在我的电脑上能够成功运行
## 没有完成的部分：
1. 目前只验证了网关能够转发服务，还没有验证能从请求头中获取信息
2. 目前的服务中没有验证能否在服务之间进行通信
3. 要求使用ingress统一入口，目前没有使用
4. Jenkins开发尚未开始

