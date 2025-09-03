#!/usr/bin/env bash
set -euo pipefail

# This script deploys the microservices into a local KinD cluster and runs basic smoke tests.

ROOT_DIR=$(cd "$(dirname "$0")/../.." && pwd)
# ROOT_DIR already points to the mirco_service_fox folder where services live
cd "$ROOT_DIR"

CLUSTER_NAME=${CLUSTER_NAME:-smartfox-ci}

echo "[kind] Ensuring cluster $CLUSTER_NAME exists"
if ! kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
  kind create cluster --name "$CLUSTER_NAME"
fi

echo "[kind] Installing local-path storage provisioner"
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.26/deploy/local-path-storage.yaml
kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'

echo "[kind] Building images and loading into cluster"
docker build -t gateway:latest ./gateway
docker build -t experiment-service:latest ./experiment-service
docker build -t user-service:latest ./user-service
docker build -t notification-service:latest ./notification-service
docker build -t submission-service:latest ./submission-service
docker build -t judge-service:latest ./judge-service

kind load docker-image gateway:latest --name "$CLUSTER_NAME"
kind load docker-image experiment-service:latest --name "$CLUSTER_NAME"
kind load docker-image user-service:latest --name "$CLUSTER_NAME"
kind load docker-image notification-service:latest --name "$CLUSTER_NAME"
kind load docker-image submission-service:latest --name "$CLUSTER_NAME"
kind load docker-image judge-service:latest --name "$CLUSTER_NAME"

echo "[k8s] Applying manifests"
kubectl apply -f user-service/k8s/
kubectl apply -f experiment-service/k8s/
kubectl apply -f submission-service/k8s/
kubectl apply -f notification-service/k8s/
kubectl apply -f judge-service/k8s/
kubectl apply -f gateway/k8s/

echo "[k8s] Wait for MySQL StatefulSets to be ready first"
for mysql in mysql-user mysql-experiment mysql-submission mysql-notification; do
  echo "等待 $mysql StatefulSet 准备就绪..."
  kubectl wait --for=condition=ready pod -l app=$mysql --timeout=600s || {
    echo "$mysql StatefulSet 未能就绪，运行诊断脚本："
    chmod +x .github/scripts/debug-k8s.sh
    .github/scripts/debug-k8s.sh
    exit 1
  }
done

echo "[k8s] Wait for deployments"
for d in user-service experiment-service submission-service notification-service judge-service gateway; do
  echo "等待部署 $d..."
  kubectl rollout status deploy/$d -n default --timeout=600s || {
    echo "部署 $d 失败，运行诊断脚本："
    chmod +x .github/scripts/debug-k8s.sh
    .github/scripts/debug-k8s.sh
    exit 1
  }
done

echo "[tests] Health checks"
kubectl run tmp-curl --rm -i --restart=Never --image=curlimages/curl:8.8.0 -- sh -lc \
  'for url in http://gateway-service.default.svc.cluster.local/health; do echo "GET $url"; curl -sS -m 10 $url; echo; done'

echo "[tests] API smoke via gateway"
# Try a minimal public endpoint set (health and maybe list with missing auth should 401)
kubectl run tmp-curl2 --rm -i --restart=Never --image=curlimages/curl:8.8.0 -- sh -lc \
  'set -e; curl -sS -m 10 http://gateway-service.default.svc.cluster.local/health | grep -i OK'

echo "All good"
