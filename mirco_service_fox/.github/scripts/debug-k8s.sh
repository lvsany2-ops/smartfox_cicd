#!/usr/bin/env bash
set -euo pipefail

echo "=== Kubernetes 诊断脚本 ==="

echo "1. 检查所有 Pod 状态："
kubectl get pods -o wide

echo -e "\n2. 检查 user-service Pod 详情："
kubectl describe pod -l app=user-service

echo -e "\n3. 检查 user-service 日志："
kubectl logs -l app=user-service --tail=50

echo -e "\n4. 检查 MySQL StatefulSet 状态："
kubectl get statefulset

echo -e "\n5. 检查 mysql-user Pod 详情："
kubectl describe pod -l app=mysql-user

echo -e "\n6. 检查 mysql-user 日志："
kubectl logs -l app=mysql-user --tail=30

echo -e "\n7. 检查所有服务："
kubectl get svc

echo -e "\n8. 检查存储卷："
kubectl get pvc

echo -e "\n9. 检查节点资源："
kubectl top nodes || echo "Metrics 不可用"
kubectl top pods || echo "Metrics 不可用"

echo -e "\n10. 检查事件："
kubectl get events --sort-by=.metadata.creationTimestamp
