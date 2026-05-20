param(
    [ValidateSet("k3d", "k8s")]
    [string]$Mode = "k8s",
    [string]$Action = "deploy",
    [switch]$Help
)

$ErrorActionPreference = "Stop"

function Show-Banner {
    Write-Host ""
    Write-Host "=====================================================" -ForegroundColor Cyan
    Write-Host "  Generator Platform - K8s/K3D Deployment Tool" -ForegroundColor Cyan
    Write-Host "=====================================================" -ForegroundColor Cyan
    Write-Host ""
}

function Show-Help {
    Show-Banner
    Write-Host "Usage: .\start-platform.ps1 -Mode <k3d|k8s> -Action <deploy|status|stop|logs>" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Parameters:" -ForegroundColor White
    Write-Host "  -Mode   : Deployment mode" -ForegroundColor Gray
    Write-Host "            k3d - Use K3D cluster (lightweight, for dev/test)" -ForegroundColor Green
    Write-Host "            k8s - Use Docker Desktop K8s (production)" -ForegroundColor Blue
    Write-Host ""
    Write-Host "  -Action : Operation type" -ForegroundColor Gray
    Write-Host "            deploy - Build and deploy all services" -ForegroundColor Green
    Write-Host "            status - Check service status" -ForegroundColor Yellow
    Write-Host "            stop   - Stop and cleanup all resources" -ForegroundColor Red
    Write-Host "            logs   - View service logs" -ForegroundColor Cyan
    Write-Host "            port   - Start port forwarding" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor White
    Write-Host "  .\start-platform.ps1 -Mode k3d -Action deploy   # Deploy with K3D" -ForegroundColor Green
    Write-Host "  .\start-platform.ps1 -Mode k8s -Action status   # Check K8s status" -ForegroundColor Blue
    Write-Host "  .\start-platform.ps1 -Mode k3d -Action stop     # Stop K3D services" -ForegroundColor Red
    Write-Host ""
}

function Test-Prerequisites {
    Write-Host "[*] Checking prerequisites..." -ForegroundColor Yellow
    
    $errors = @()
    
    if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
        $errors += "kubectl not installed"
    }
    
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        $errors += "Docker not installed"
    }
    
    if ($Mode -eq "k3d") {
        $hasK3D = Get-Command k3d -ErrorAction SilentlyContinue
        $hasK3DContext = (kubectl config get-contexts -o name) -contains "k3d-gen-platform-test"
        
        if (-not $hasK3D -and -not $hasK3DContext) {
            $errors += "k3d not installed and no k3d cluster found"
        }
    }
    
    if ($errors.Count -gt 0) {
        foreach ($err in $errors) {
            Write-Host "[ERROR] $err" -ForegroundColor Red
        }
        exit 1
    }
    
    Write-Host "[OK] All prerequisites met" -ForegroundColor Green
}

function Switch-KubeContext {
    param([string]$TargetMode)
    
    Write-Host "[*] Switching Kubernetes context..." -ForegroundColor Yellow
    
    if ($TargetMode -eq "k3d") {
        $contextName = "k3d-gen-platform-test"
        
        $contexts = kubectl config get-contexts -o name
        
        if ($contexts -notcontains $contextName) {
            Write-Host "[!] K3D cluster does not exist, creating..." -ForegroundColor Yellow
            
            k3d cluster create gen-platform-test --agents 2 --port "30080:80@loadbalancer" --port "30443:443@loadbalancer" 2>&1 | Out-Null
            
            if ($LASTEXITCODE -ne 0) {
                Write-Host "[ERROR] Failed to create K3D cluster" -ForegroundColor Red
                exit 1
            }
            
            Start-Sleep -Seconds 5
        }
        
        kubectl config use-context $contextName | Out-Null
        Write-Host "[OK] Switched to K3D cluster ($contextName)" -ForegroundColor Green
        
    } else {
        kubectl config use-context docker-desktop | Out-Null
        Write-Host "[OK] Switched to Docker Desktop K8s" -ForegroundColor Blue
    }
    
    Start-Sleep -Seconds 2
    
    try {
        kubectl cluster-info | Out-Null
        Write-Host "[OK] Cluster connection successful" -ForegroundColor Green
    } catch {
        Write-Host "[ERROR] Cannot connect to cluster" -ForegroundColor Red
        exit 1
    }
}

function Build-DockerImages {
    Write-Host ""
    Write-Host "[*] Building Docker images..." -ForegroundColor Yellow
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    
    $services = @(
        @{Name="api-gateway"; Path="apps/api-gateway"},
        @{Name="auth-service"; Path="apps/authentication-service"},
        @{Name="user-service"; Path="apps/user-service"},
        @{Name="project-service"; Path="apps/project-service"},
        @{Name="generator-service"; Path="apps/generator-service"},
        @{Name="operations-service"; Path="apps/operations-service"},
        @{Name="cluster-service"; Path="apps/cluster-service"},
        @{Name="web-admin"; Path="apps/web-admin"}
    )
    
    foreach ($svc in $services) {
        Write-Host "[BUILD] $($svc.Name)..." -ForegroundColor White -NoNewline

        $dockerCmd = "docker build -t generator-platform/$($svc.Name):latest -f $($svc.Path)/Dockerfile ."
        Invoke-Expression $dockerCmd 2>&1 | Out-Null

        if ($LASTEXITCODE -eq 0) {
            Write-Host " [OK]" -ForegroundColor Green
        } else {
            Write-Host " [FAIL]" -ForegroundColor Red
            exit 1
        }
    }
    
    Write-Host ""
    Write-Host "[OK] All images built successfully!" -ForegroundColor Green
}

function Deploy-ToKubernetes {
    Write-Host ""
    Write-Host "[*] Deploying to Kubernetes..." -ForegroundColor Yellow
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    
    $namespace = "generator-platform"
    
    Write-Host "[*] Applying configuration files..." -ForegroundColor White
    
    $yamlFiles = @(
        "infra/k8s/namespace.yaml",
        "infra/k8s/rbac.yaml",
        "infra/k8s/postgres.yaml",
        "infra/k8s/redis.yaml",
        "infra/k8s/auth-service.yaml",
        "infra/k8s/user-service.yaml",
        "infra/k8s/project-service.yaml",
        "infra/k8s/generator-service.yaml",
        "infra/k8s/operations-service.yaml",
        "infra/k8s/cluster-service.yaml",
        "infra/k8s/api-gateway.yaml",
        "infra/k8s/web-admin.yaml",
        "infra/k8s/ingress.yaml"
    )
    
    foreach ($file in $yamlFiles) {
        if (Test-Path $file) {
            kubectl apply -f $file 2>&1 | Out-Null
            Write-Host "  [+] $(Split-Path $file -Leaf)" -ForegroundColor DarkGray
        }
    }
    
    Write-Host ""
    Write-Host "[*] Waiting for Pods to start..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
    
    Show-Status
}

function Show-Status {
    Write-Host ""
    Write-Host "[*] Service Status" -ForegroundColor Yellow
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    
    kubectl get pods -n generator-platform -o wide
    
    Write-Host ""
    Write-Host "[*] Service Endpoints:" -ForegroundColor Yellow
    
    if ($Mode -eq "k3d") {
        Write-Host "  Frontend: http://localhost:30080 (via LoadBalancer)" -ForegroundColor Green
        Write-Host "  API:      http://localhost:30080/api/v1" -ForegroundColor Cyan
    } else {
        Write-Host "  Frontend: Run port-forward command to access" -ForegroundColor Green
        Write-Host "  API:      http://localhost:8080 (requires port-forward)" -ForegroundColor Cyan
        Write-Host "" 
        Write-Host "  [TIP] Run: .\start-platform.ps1 -Mode k8s -Action port" -ForegroundColor Magenta
    }
    
    Write-Host ""
}

function Stop-Services {
    Write-Host ""
    Write-Host "[*] Stopping all services..." -ForegroundColor Red
    
    kubectl delete -f infra/k8s/ingress.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/web-admin.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/api-gateway.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/cluster-service.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/operations-service.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/generator-service.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/project-service.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/user-service.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/auth-service.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/redis.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/postgres.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/rbac.yaml --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete -f infra/k8s/namespace.yaml --ignore-not-found=true 2>&1 | Out-Null
    
    Write-Host "[OK] All resources deleted" -ForegroundColor Green
}

function Show-Logs {
    param([string]$Service = "all")
    
    if ($Service -eq "all") {
        Write-Host "[*] Showing all service logs (Ctrl+C to exit)" -ForegroundColor Yellow
        kubectl logs -n generator-platform -l app --tail=50 -f --max-log-requests=10
    } else {
        Write-Host "[*] Showing $Service logs" -ForegroundColor Yellow
        kubectl logs -n generator-platform -l app=$Service -f --tail=100
    }
}

function Start-PortForward {
    Write-Host ""
    Write-Host "[*] Starting port forwarding..." -ForegroundColor Magenta
    Write-Host ""
    
    Write-Host "Frontend (http://localhost:8888):" -ForegroundColor Green
    Start-Job -ScriptBlock { param($ns) kubectl port-forward -n $ns svc/web-admin 8888:3000 } -ArgumentList "generator-platform" | Out-Null
    
    Start-Sleep -Seconds 2
    
    Write-Host "API Gateway (http://localhost:8080):" -ForegroundColor Cyan
    kubectl port-forward -n generator-platform svc/api-gateway 8080:8080
}

if ($Help) {
    Show-Help
    exit 0
}

Show-Banner

Write-Host "[*] Mode: " -NoNewline -ForegroundColor White
if ($Mode -eq "k3d") {
    Write-Host "K3D (lightweight)" -ForegroundColor Green
} else {
    Write-Host "K8S (Docker Desktop)" -ForegroundColor Blue
}
Write-Host "[*] Action: $Action" -ForegroundColor White
Write-Host ""

Test-Prerequisites

Switch-KubeContext -TargetMode $Mode

switch ($Action) {
    "deploy" {
        Build-DockerImages
        Deploy-ToKubernetes
        
        Write-Host ""
        Write-Host "=====================================================" -ForegroundColor Cyan
        Write-Host "              [SUCCESS] Deployment Complete!" -ForegroundColor Green
        Write-Host "=====================================================" -ForegroundColor Cyan
        Write-Host "" -ForegroundColor White
        Write-Host "  Login Credentials:" -ForegroundColor White
        Write-Host "    Username: admin" -ForegroundColor Gray
        Write-Host "    Password: admin123" -ForegroundColor Gray
        Write-Host "" -ForegroundColor White
        Write-Host "  Access URL:" -ForegroundColor White
        if ($Mode -eq "k3d") {
            Write-Host "    http://localhost:30080" -ForegroundColor Green
        } else {
            Write-Host "    Run: .\start-platform.ps1 -Mode k8s -Action port" -ForegroundColor Yellow
        }
        Write-Host "=====================================================" -ForegroundColor Cyan
        Write-Host ""
    }
    
    "status" {
        Show-Status
    }
    
    "stop" {
        Stop-Services
    }
    
    "logs" {
        Show-Logs
    }
    
    "port" {
        Start-PortForward
    }
    
    default {
        Write-Host "[ERROR] Unknown action: $Action" -ForegroundColor Red
        Write-Host "Run .\start-platform.ps1 -Help for usage" -ForegroundColor Yellow
        exit 1
    }
}
