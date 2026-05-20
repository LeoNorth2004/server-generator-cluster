.PHONY: all build push deploy clean apply delete status logs

# Docker registry prefix
REGISTRY ?= generator-platform

# Namespace
NAMESPACE ?= generator-platform

# Image tag
TAG ?= latest

all: build push deploy

# Build all Docker images
build:
	@echo "Building Docker images..."
	docker build -t $(REGISTRY)/api-gateway:$(TAG) -f apps/api-gateway/Dockerfile .
	docker build -t $(REGISTRY)/auth-service:$(TAG) -f apps/authentication-service/Dockerfile .
	docker build -t $(REGISTRY)/user-service:$(TAG) -f apps/user-service/Dockerfile .
	docker build -t $(REGISTRY)/project-service:$(TAG) -f apps/project-service/Dockerfile .
	docker build -t $(REGISTRY)/generator-service:$(TAG) -f apps/generator-service/Dockerfile .
	docker build -t $(REGISTRY)/operations-service:$(TAG) -f apps/operations-service/Dockerfile .
	docker build -t $(REGISTRY)/cluster-service:$(TAG) -f apps/cluster-service/Dockerfile .
	docker build -t $(REGISTRY)/web-admin:$(TAG) --build-arg DEPLOY_MODE=k8s -f apps/web-admin/Dockerfile .

# Push all Docker images
push:
	@echo "Pushing Docker images..."
	docker push $(REGISTRY)/api-gateway:$(TAG)
	docker push $(REGISTRY)/auth-service:$(TAG)
	docker push $(REGISTRY)/user-service:$(TAG)
	docker push $(REGISTRY)/project-service:$(TAG)
	docker push $(REGISTRY)/generator-service:$(TAG)
	docker push $(REGISTRY)/operations-service:$(TAG)
	docker push $(REGISTRY)/cluster-service:$(TAG)
	docker push $(REGISTRY)/web-admin:$(TAG)

# Deploy to Kubernetes
deploy: apply

# Apply all Kubernetes configurations
apply:
	@echo "Applying Kubernetes configurations..."
	kubectl apply -f infra/k8s/namespace.yaml
	kubectl apply -f infra/k8s/rbac.yaml
	kubectl apply -f infra/k8s/postgres.yaml
	kubectl apply -f infra/k8s/redis.yaml
	kubectl apply -f infra/k8s/auth-service.yaml
	kubectl apply -f infra/k8s/user-service.yaml
	kubectl apply -f infra/k8s/project-service.yaml
	kubectl apply -f infra/k8s/generator-service.yaml
	kubectl apply -f infra/k8s/operations-service.yaml
	kubectl apply -f infra/k8s/cluster-service.yaml
	kubectl apply -f infra/k8s/api-gateway.yaml
	kubectl apply -f infra/k8s/web-admin.yaml
	kubectl apply -f infra/k8s/ingress.yaml

# Delete all Kubernetes resources
delete:
	@echo "Deleting Kubernetes resources..."
	kubectl delete -f infra/k8s/ingress.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/web-admin.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/api-gateway.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/cluster-service.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/operations-service.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/generator-service.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/project-service.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/user-service.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/auth-service.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/redis.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/postgres.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/rbac.yaml --ignore-not-found=true
	kubectl delete -f infra/k8s/namespace.yaml --ignore-not-found=true

# Show status of all pods
status:
	@echo "Pod status in $(NAMESPACE) namespace:"
	kubectl get pods -n $(NAMESPACE)
	@echo ""
	@echo "Service status in $(NAMESPACE) namespace:"
	kubectl get svc -n $(NAMESPACE)

# Show logs of a specific service
logs:
	@echo "Logs for $(SERVICE) in $(NAMESPACE) namespace:"
	kubectl logs -n $(NAMESPACE) -l app=$(SERVICE) --tail=100

# Watch pod status
watch:
	watch -n 5 kubectl get pods -n $(NAMESPACE)

# Port forward to api-gateway
port-forward:
	kubectl port-forward -n $(NAMESPACE) svc/api-gateway 8080:8080

# Port forward to web-admin
port-forward-web:
	kubectl port-forward -n $(NAMESPACE) svc/web-admin 3000:3000

# Describe a specific pod
describe:
	kubectl describe pod -n $(NAMESPACE) -l app=$(SERVICE)

# Exec into a pod
exec:
	kubectl exec -it -n $(NAMESPACE) -l app=$(SERVICE) -- /bin/sh

# Restart a service
restart:
	kubectl rollout restart deployment/$(SERVICE) -n $(NAMESPACE)

# Scale a service
scale:
	kubectl scale deployment/$(SERVICE) -n $(NAMESPACE) --replicas=$(REPLICAS)

# Clean up completed pods
clean:
	kubectl delete pods -n $(NAMESPACE) --field-selector=status.phase==Succeeded

# Get all resources
get:
	@echo "=== All resources in $(NAMESPACE) ==="
	@echo "--- Pods ---"
	kubectl get pods -n $(NAMESPACE)
	@echo "--- Services ---"
	kubectl get svc -n $(NAMESPACE)
	@echo "--- Deployments ---"
	kubectl get deployments -n $(NAMESPACE)
	@echo "--- ConfigMaps ---"
	kubectl get configmaps -n $(NAMESPACE)
	@echo "--- Secrets ---"
	kubectl get secrets -n $(NAMESPACE)
	@echo "--- PersistentVolumeClaims ---"
	kubectl get pvc -n $(NAMESPACE)