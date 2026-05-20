#!/bin/bash

echo "=========================================="
echo "  Generator Platform - K8s Deployment"
echo "=========================================="
echo ""

# Check kubectl
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl not found. Please install kubectl first."
    exit 1
fi

# Check cluster connection
echo "[1/5] Checking Kubernetes cluster..."
kubectl cluster-info
if [ $? -ne 0 ]; then
    echo "Error: Cannot connect to Kubernetes cluster."
    echo "Make sure your K8s cluster is running (Docker Desktop, minikube, etc.)"
    exit 1
fi
echo ""

# Create namespace
echo "[2/5] Creating namespace: generator-platform"
kubectl apply -f infra/k8s/namespace.yaml
echo ""

# Deploy infrastructure (PostgreSQL, Redis)
echo "[3/5] Deploying infrastructure..."
kubectl apply -f infra/k8s/postgres.yaml
kubectl apply -f infra/k8s/redis.yaml
echo ""

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
kubectl wait --namespace=generator-platform --for=condition=ready pod -l app=postgres --timeout=120s
echo ""

# Deploy all services
echo "[4/5] Deploying microservices..."
kubectl apply -f infra/k8s/api-gateway.yaml
kubectl apply -f infra/k8s/auth-service.yaml
kubectl apply -f infra/k8s/user-service.yaml
kubectl apply -f infra/k8s/project-service.yaml
kubectl apply -f infra/k8s/generator-service.yaml
kubectl apply -f infra/k8s/operations-service.yaml
kubectl apply -f infra/k8s/cluster-service.yaml
kubectl apply -f infra/k8s/web-admin.yaml
echo ""

# Deploy RBAC and Ingress (if exists)
if [ -f infra/k8s/rbac.yaml ]; then
    kubectl apply -f infra/k8s/rbac.yaml
fi

if [ -f infra/k8s/ingress.yaml ]; then
    kubectl apply -f infra/k8s/ingress.yaml
fi
echo ""

# Show status
echo "[5/5] Deployment status:"
echo ""
echo "Namespace: generator-platform"
kubectl get namespace generator-platform
echo ""
echo "Pods:"
kubectl get pods -n generator-platform
echo ""
echo "Services:"
kubectl get svc -n generator-platform
echo ""
echo "Deployments:"
kubectl get deployments -n generator-platform
echo ""

echo "=========================================="
echo "  Access Information"
echo "=========================================="
echo ""

# Get web-admin service info
WEB_ADMIN_IP=$(kubectl get svc web-admin -n generator-platform -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
if [ -z "$WEB_ADMIN_IP" ]; then
    WEB_ADMIN_PORT=$(kubectl get svc web-admin -n generator-platform -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null)
    if [ -n "$WEB_ADMIN_PORT" ]; then
        echo "Web Admin: http://localhost:$WEB_ADMIN_PORT"
    else
        echo "Web Admin: Port-forward with: kubectl port-forward svc/web-admin 3000:3000 -n generator-platform"
    fi
else
    echo "Web Admin: http://$WEB_ADMIN_IP:3000"
fi

echo ""
echo "API Gateway: Internal service on port 8080"
echo "To access API: kubectl port-forward svc/api-gateway 8080:8080 -n generator-platform"
echo ""

echo "=========================================="
echo "  Done! System deployed to K8s"
echo "=========================================="
