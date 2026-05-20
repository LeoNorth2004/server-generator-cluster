<#
.SYNOPSIS
    Generator Platform - Interactive Deployment Tool (Docker / K8S / K3D)
.DESCRIPTION
    Interactive menu-driven deployment script with port conflict detection.
.EXAMPLE
    .\start.ps1                    # Launch interactive mode
.EXAMPLE
    .\start.ps1 -Mode docker -Action deploy  # Non-interactive mode
#>

param(
    [ValidateSet("docker", "k8s", "k3d")]
    [string]$Mode,
    
    [ValidateSet("deploy", "status", "stop", "logs", "port", "cleanup", "check-ports", "open")]
    [string]$Action,
    
    [int]$FrontendPort = 3000,
    [int]$APIPort = 8080,
    
    [switch]$BuildImages,
    [switch]$Force,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

# Global Configuration
$Namespace = "generator-platform"
$ProjectRoot = $PSScriptRoot
if (-not $ProjectRoot) { $ProjectRoot = "." }

$Config = @{
    FrontendPort = $FrontendPort
    APIPort      = $APIPort
    DBPort       = 5432
    RedisPort    = 6379
}

$ModePorts = @{
    "docker" = @(3000, 3001, 8080, 5432, 6379)
    "k8s"    = @(3000, 8080)
    "k3d"    = @(3000, 8080, 5432, 6379, 30080, 30443)
}

$Services = @(
    @{Name="api-gateway"; Dockerfile="apps/api-gateway/Dockerfile"},
    @{Name="auth-service"; Dockerfile="apps/authentication-service/Dockerfile"},
    @{Name="user-service"; Dockerfile="apps/user-service/Dockerfile"},
    @{Name="project-service"; Dockerfile="apps/project-service/Dockerfile"},
    @{Name="generator-service"; Dockerfile="apps/generator-service/Dockerfile"},
    @{Name="operations-service"; Dockerfile="apps/operations-service/Dockerfile"},
    @{Name="cluster-service"; Dockerfile="apps/cluster-service/Dockerfile"},
    @{Name="web-admin"; Dockerfile="apps/web-admin/Dockerfile"}
)

# UI Functions
function Show-Banner {
    Clear-Host
    Write-Host ""
    Write-Host "=====================================================" -ForegroundColor Cyan
    Write-Host "  Generator Platform - Deployment Tool" -ForegroundColor Cyan
    Write-Host "=====================================================" -ForegroundColor Cyan
    Write-Host ""
}

function Show-Help {
    Show-Banner
    
    Write-Host "Usage:" -ForegroundColor Yellow
    Write-Host "  .\start.ps1                          # Interactive mode" -ForegroundColor White
    Write-Host "  .\start.ps1 -Mode docker -Action deploy # Non-interactive" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Modes:" -ForegroundColor Yellow
    Write-Host "  docker  - Docker Compose (simplest)" -ForegroundColor Green
    Write-Host "  k8s     - Docker Desktop K8s (production)" -ForegroundColor Blue
    Write-Host "  k3d     - K3D lightweight cluster (dev/test)" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "Actions:" -ForegroundColor Yellow
    Write-Host "  deploy, status, stop, logs, port, cleanup, check-ports, open" -ForegroundColor White
    Write-Host ""
    Write-Host "  open - Open browser with the frontend URL (no more clicking!)" -ForegroundColor Green
    Write-Host ""
}

function Write-Success {
    param([string]$Message)
    Write-Host "[OK]   $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Err {
    param([string]$Message)
    Write-Host "[ERR]  $Message" -ForegroundColor Red
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor White
}

# Port Detection Functions
function Test-Port {
    param([int]$Port)
    
    try {
        $connection = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue
        return ($connection -ne $null -and $connection.State -eq 'Listen')
    } catch {
        return $false
    }
}

function Get-PortProcess {
    param([int]$Port)
    
    try {
        $connection = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue |
            Where-Object { $_.State -eq 'Listen' } |
            Select-Object -First 1
        
        if ($connection) {
            $process = Get-Process -Id $connection.OwningProcess -ErrorAction SilentlyContinue
            return @{
                Port = $Port
                PID = $connection.OwningProcess
                ProcessName = if ($process) { $process.ProcessName } else { "Unknown" }
            }
        }
    } catch {}
    
    return $null
}

function Test-Ports-Available {
    param(
        [array]$RequiredPorts,
        [string]$ModeName
    )
    
    Write-Host ""
    Write-Host "Checking ports for $ModeName mode..." -ForegroundColor Yellow
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    Write-Host ""
    
    $conflicts = @()
    
    foreach ($port in $RequiredPorts) {
        $inUse = Test-Port -Port $port
        
        if ($inUse) {
            $processInfo = Get-PortProcess -Port $port
            $conflicts += @{
                Port = $port
                Process = "$($processInfo.ProcessName) (PID: $($processInfo.PID))"
            }
            Write-Host ("  [X] Port {0,-8} -> {1,-20} (PID: {2})" -f $port, $processInfo.ProcessName, $processInfo.PID) -ForegroundColor Red
        } else {
            Write-Host ("  [OK] Port {0,-8} -> Available" -f $port) -ForegroundColor Green
        }
    }
    
    Write-Host ""
    
    if ($conflicts.Count -gt 0) {
        Write-Warn -Message "Found $($conflicts.Count) port conflict(s)"
        Write-Host ""
        
        if (-not $Force) {
            Write-Host "Options:" -ForegroundColor White
            Write-Host "  1. Stop the conflicting process(es)" -ForegroundColor Gray
            Write-Host "  2. Use different ports" -ForegroundColor Gray
            Write-Host "  3. Force deploy anyway (may cause issues)" -ForegroundColor DarkRed
            Write-Host ""
            
            if (-not $Mode -and -not $Action) {
                $choice = Read-Host "Your choice (1-3, or Enter to go back)"
                
                switch ($choice) {
                    "1" { 
                        foreach ($c in $conflicts) {
                            Write-Host "Stopping process on port $($c.Port)..." -ForegroundColor Yellow
                            $proc = Get-PortProcess -Port $c.Port
                            Stop-Process -Id $proc.PID -Force -ErrorAction SilentlyContinue
                        }
                        Write-Success -Message "Processes stopped"
                        return $true
                    }
                    "2" { 
                        $newFp = Read-Host "Enter frontend port (current: $($Config.FrontendPort))"
                        $newAp = Read-Host "Enter API port (current: $($Config.APIPort))"
                        
                        if ($newFp -match '^\d+$') { $Config.FrontendPort = [int]$newFp }
                        if ($newAp -match '^\d+$') { $Config.APIPort = [int]$newAp }
                        
                        Write-Info -Message "Using custom ports: Frontend=$($Config.FrontendPort), API=$($Config.APIPort)"
                        return $true
                    }
                    "3" { 
                        Write-Warn -Message "Forcing deployment..."
                        return $true
                    }
                    default {
                        Write-Info -Message "Going back..."
                        return $false
                    }
                }
            } else {
                Write-Host "Use -Force to ignore conflicts or change ports with -FrontendPort/-APIPort" -ForegroundColor DarkGray
                return $false
            }
        } else {
            Write-Warn -Message "-Force flag set, ignoring conflicts..."
            return $true
        }
    } else {
        Write-Success -Message "All ports are available!"
        return $true
    }
}

function Show-PortSummary {
    param([string]$TargetMode)
    
    Write-Host ""
    Write-Host "Port Configuration:" -ForegroundColor Yellow
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host ("  Mode:         {0}" -f $TargetMode) -ForegroundColor White
    Write-Host ("  Frontend:     http://localhost:{0}" -f $Config.FrontendPort) -ForegroundColor Green
    Write-Host ("  API Gateway:  http://localhost:{0}/api/v1" -f $Config.APIPort) -ForegroundColor Cyan
    Write-Host ("  PostgreSQL:   localhost:{0}" -f $Config.DBPort) -ForegroundColor DarkGray
    Write-Host ("  Redis:        localhost:{0}" -f $Config.RedisPort) -ForegroundColor DarkGray
    Write-Host ""
}

# Interactive Menu Functions
function Show-MainMenu {
    Show-Banner
    
    Write-Host "Select Deployment Mode:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  [1] Docker Compose   (Simplest, recommended for beginners)" -ForegroundColor Green
    Write-Host "  [2] Kubernetes       (Docker Desktop, production-like)" -ForegroundColor Blue
    Write-Host "  [3] K3D              (Lightweight cluster, dev/testing)" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "  [4] Check Ports      (Verify port availability)" -ForegroundColor DarkCyan
    Write-Host "  [5] Cleanup Project  (Remove temp files and old images)" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "  [0] Exit" -ForegroundColor Red
    Write-Host ""
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    Write-Host ""
}

function Show-ActionMenu {
    param([string]$SelectedMode)
    
    Show-Banner
    
    $modeColor = switch ($SelectedMode) {
        "docker" { "Green" }
        "k8s"    { "Blue" }
        "k3d"    { "Magenta" }
    }
    
    Write-Host "Mode: " -NoNewline -ForegroundColor Gray
    Write-Host $SelectedMode.ToUpper() -ForegroundColor $modeColor
    Write-Host ""
    Write-Host "Select Action:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  [1] Deploy Services    (Build images and start all services)" -ForegroundColor Green
    Write-Host "  [2] Check Status       (View running pods/containers)" -ForegroundColor Yellow
    Write-Host "  [3] Stop Services      (Shutdown everything)" -ForegroundColor Red
    Write-Host "  [4] View Logs          (Real-time log output)" -ForegroundColor Cyan
    Write-Host "  [5] Port Forwarding    (For K8S/K3D local access)" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "  [6] Change Ports       (Customize frontend/API ports)" -ForegroundColor White
    Write-Host "  [7] Rebuild Images     (Force rebuild all Docker images)" -ForegroundColor DarkYellow
    Write-Host "  [8] Open Browser       (Launch browser with URL)" -ForegroundColor Green
    Write-Host ""
    Write-Host "  [0] Back to Main Menu" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    Write-Host ""
}

function Show-PortStatusDashboard {
    Show-Banner
    
    Write-Host "=====================================================" -ForegroundColor Cyan
    Write-Host "           Port Availability Dashboard              " -ForegroundColor Cyan
    Write-Host "=====================================================" -ForegroundColor Cyan
    Write-Host ""
    
    $allModes = @("docker", "k8s", "k3d")
    
    foreach ($modeName in $allModes) {
        $ports = $ModePorts[$modeName]
        $available = 0
        $busy = 0
        
        $modeLabel = switch ($modeName) {
            "docker" { "DOCKER COMPOSE" }
            "k8s"    { "KUBERNETES" }
            "k3d"    { "K3D" }
        }
        
        $modeColor = switch ($modeName) {
            "docker" { "Green" }
            "k8s"    { "Blue" }
            "k3d"    { "Magenta" }
        }
        
        Write-Host "+- $modeLabel ----------------------------------------+" -ForegroundColor $modeColor
        
        foreach ($port in $ports) {
            if (Test-Port -Port $port) {
                $proc = Get-PortProcess -Port $port
                Write-Host ("|  [X] Port {0,-8} -> {1,-20} (PID: {2})" -f $port, $proc.ProcessName, $proc.PID) -ForegroundColor Red
                $busy++
            } else {
                Write-Host ("|  [OK] Port {0,-8} -> Available" -f $port) -ForegroundColor Green
                $available++
            }
        }
        
        $statusText = if ($busy -eq 0) { "READY" } else { "CONFLICTS DETECTED" }
        $statusColor = if ($busy -eq 0) { "Green" } else { "Red" }
        
        Write-Host "+- Status: $available/$($ports.Count) available, $busy busy [$statusText]" -ForegroundColor $statusColor
        Write-Host ""
    }
    
    Write-Host "Current Configuration:" -ForegroundColor Yellow
    Write-Host ("  Frontend Port: {0}" -f $Config.FrontendPort) -ForegroundColor White
    Write-Host ("  API Port:      {0}" -f $Config.APIPort) -ForegroundColor White
    Write-Host ""
    
    Write-Host "Press any key to continue..." -ForegroundColor DarkGray
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
}

function Select-ModeInteractive {
    while ($true) {
        Show-MainMenu
        $choice = Read-Host "Enter your choice (0-5)"
        
        switch ($choice) {
            "1" { return "docker" }
            "2" { return "k8s" }
            "3" { return "k3d" }
            "4" { 
                Show-PortStatusDashboard
                continue 
            }
            "5" { 
                Write-Host "" 
                Write-Warn -Message "This will delete temporary files!"
                $confirm = Read-Host "Confirm cleanup? (y/n)"
                if ($confirm -eq 'y') {
                    & "$ProjectRoot\cleanup-project.ps1" -Force
                }
                continue 
            }
            "0" { 
                Write-Host "Goodbye!" -ForegroundColor Cyan
                exit 0 
            }
            default {
                Write-Err -Message "Invalid choice. Please enter 0-5."
                Start-Sleep -Seconds 1
            }
        }
    }
}

function Select-ActionInteractive {
    param([string]$SelectedMode)
    
    while ($true) {
        Show-ActionMenu -SelectedMode $SelectedMode
        $choice = Read-Host "Enter your choice (0-8)"
        
        switch ($choice) {
            "1" { return "deploy" }
            "2" { return "status" }
            "3" { return "stop" }
            "4" { return "logs" }
            "5" { return "port" }
            "6" {
                Write-Host ""
                $newFp = Read-Host "Enter frontend port (current: $($Config.FrontendPort), Enter to keep)"
                $newAp = Read-Host "Enter API port (current: $($Config.APIPort), Enter to keep)"
                
                if ($newFp -match '^\d+$' -and [int]$newFp -gt 0 -and [int]$newFp -lt 65536) { 
                    $Config.FrontendPort = [int]$newFp 
                    Write-Success -Message "Frontend port set to $newFp"
                }
                if ($newAp -match '^\d+$' -and [int]$newAp -gt 0 -and [int]$newAp -lt 65536) { 
                    $Config.APIPort = [int]$newAp 
                    Write-Success -Message "API port set to $newAp"
                }
                
                Start-Sleep -Seconds 1
                continue
            }
            "7" {
                $BuildImages = $true
                return "deploy"
            }
            "8" {
                return "open"
            }
            "0" { return "back" }
            default {
                Write-Err -Message "Invalid choice. Please enter 0-8."
                Start-Sleep -Seconds 1
            }
        }
    }
}

function Show-DeploymentComplete {
    param([string]$DeployedMode)
    
    Show-Banner
    
    Write-Host "=====================================================" -ForegroundColor Green
    Write-Host "               DEPLOYMENT COMPLETE!                   " -ForegroundColor Green
    Write-Host "=====================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "  Login Credentials:" -ForegroundColor White
    Write-Host "    Username: admin" -ForegroundColor Gray
    Write-Host "    Password: admin123" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Access URLs (clickable or copy-paste):" -ForegroundColor White
    Write-Host ("    Frontend: http://localhost:{0}" -f $Config.FrontendPort) -ForegroundColor Green
    Write-Host ("    API:      http://localhost:{0}/api/v1" -f $Config.APIPort) -ForegroundColor Cyan
    Write-Host ""
    
    # Copy URL to clipboard for easy access
    $frontendUrl = ("http://localhost:{0}" -f $Config.FrontendPort)
    try {
        $frontendUrl | Set-Clipboard
        Write-Host "  [TIP] Frontend URL copied to clipboard!" -ForegroundColor Magenta
        Write-Host "        Just paste in browser (Ctrl+V)" -ForegroundColor DarkGray
    } catch {
        # Clipboard might not work in some environments, that's ok
    }
    
    if ($DeployedMode -ne "docker") {
        Write-Host ""
        Write-Host "  Next Step:" -ForegroundColor Yellow
        Write-Host "    Run this command to start port forwarding:" -ForegroundColor DarkGray
        Write-Host ("    .\start.ps1 -Mode {0} -Action port" -f $DeployedMode) -ForegroundColor Magenta
        Write-Host ""
    } else {
        Write-Host "  Status: Direct access (no port forwarding needed)" -ForegroundColor DarkGreen
        Write-Host ""
        
        # For Docker mode, offer to open browser automatically
        Write-Host "  Quick Actions:" -ForegroundColor Yellow
        Write-Host "    [O] Open browser now" -ForegroundColor Green
        Write-Host "    [Enter] Continue" -ForegroundColor Gray
        Write-Host ""
        
        $key = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        if ($key.Character -eq 'o' -or $key.Character -eq 'O') {
            try {
                Start-Process $frontendUrl
                Write-Success -Message "Browser opened!"
            } catch {
                Write-Warn -Message "Could not open browser automatically"
                Write-Host ("  Please open manually: {0}" -f $frontendUrl) -ForegroundColor DarkGray
            }
            Start-Sleep -Seconds 1
        }
    }
    
    Write-Host ""
    Write-Host "Press any key to continue..." -ForegroundColor DarkGray
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
}

# Prerequisites Check
function Test-Prerequisites {
    Write-Host "Checking prerequisites..." -ForegroundColor Yellow
    
    $errors = @()
    
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        $errors += "Docker not installed"
    }
    
    if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
        $errors += "kubectl not installed"
    }
    
    switch ($Mode) {
        "docker" {
            $composeVersion = docker compose version 2>&1
            if ($LASTEXITCODE -ne 0) {
                $errors += "Docker Compose not available"
            }
        }
        "k3d" {
            $hasK3D = Get-Command k3d -ErrorAction SilentlyContinue
            $hasK3DContext = (kubectl config get-contexts -o name 2>$null) -contains "k3d-gen-platform-test"
            
            if (-not $hasK3D -and -not $hasK3DContext) {
                $errors += "k3d not installed and no k3d cluster found"
            }
        }
        "k8s" {
            try {
                $contexts = kubectl config get-contexts -o name 2>$null
                if ($contexts -notcontains "docker-desktop") {
                    Write-Warn -Message "Docker Desktop K8s context not found"
                }
            } catch {
                $errors += "Cannot connect to Kubernetes cluster"
            }
        }
    }
    
    if ($errors.Count -gt 0) {
        foreach ($err in $errors) {
            Write-Err -Message $err
        }
        
        if (-not $Mode -or -not $Action) {
            Write-Host ""
            Write-Host "Press any key to go back..." -ForegroundColor DarkGray
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        }
        
        return $false
    }
    
    if ($Action -eq "deploy") {
        $requiredPorts = $ModePorts[$Mode]
        
        if ($FrontendPort -ne 3000 -and $requiredPorts -notcontains $FrontendPort) {
            $requiredPorts += $FrontendPort
        }
        if ($APIPort -ne 8080 -and $requiredPorts -notcontains $APIPort) {
            $requiredPorts += $APIPort
        }
        
        $portsOk = Test-Ports-Available -RequiredPorts $requiredPorts -ModeName $Mode.ToUpper()
        
        if (-not $portsOk) {
            if (-not $Mode -or -not $Action) {
                return $false
            }
            exit 1
        }
        
        Show-PortSummary -TargetMode $Mode.ToUpper()
        
        if ($Mode -and $Action) {
            Write-Host "Ready to deploy? Press any key to continue (Ctrl+C to cancel)..." -ForegroundColor Yellow
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        } else {
            Start-Sleep -Seconds 2
        }
    }
    
    # For deploy action, verify cluster connection (mode conflicts already handled by Stop-AllOtherModes)
    if ($Action -eq "deploy") {
        # Verify we're connected to the correct cluster (for K8S/K3D modes)
        if ($Mode -eq "k8s" -or $Mode -eq "k3d") {
            if (-not (Confirm-CorrectCluster -ExpectedMode $Mode)) {
                return $false
            }
        }
    }
    
    Write-Success -Message "All checks passed!"
    return $true
}

# Force stop all other deployment modes before starting a new one
function Stop-AllOtherModes {
    param([string]$TargetMode)
    
    Write-Host ""
    Write-Host "=====================================================" -ForegroundColor Magenta
    Write-Host "  STOPPING ALL OTHER DEPLOYMENT MODES" -ForegroundColor Red
    Write-Host "  Target: $TargetMode (exclusive mode)" -ForegroundColor Cyan
    Write-Host "=====================================================" -ForegroundColor Magenta
    Write-Host ""
    
    $stoppedCount = 0
    
    # 1. Stop Docker Compose if target is not docker
    if ($TargetMode -ne "docker") {
        Write-Host "[1/3] Checking Docker Compose..." -ForegroundColor White
        
        Push-Location $ProjectRoot
        try {
            $composeOutput = docker compose ps 2>&1
            $hasRunning = $false
            
            foreach ($line in $composeOutput) {
                if ($line -match "\b(Up|running)\b" -and $line -notmatch "level=warning") {
                    $hasRunning = $true
                    break
                }
            }
            
            if ($hasRunning) {
                Write-Host "  Stopping Docker Compose containers..." -ForegroundColor Yellow
                docker compose down 2>&1 | Out-Null
                Start-Sleep -Seconds 3
                
                # Verify stopped
                $checkOutput = docker compose ps 2>&1
                $stillRunning = $false
                foreach ($line in $checkOutput) {
                    if ($line -match "\b(Up|running)\b" -and $line -notmatch "level=warning") {
                        $stillRunning = $true
                    }
                }
                
                if (-not $stillRunning) {
                    Write-Success -Message "Docker Compose: All containers stopped"
                    $stoppedCount++
                } else {
                    Write-Warn -Message "Docker Compose: Some containers may still be running"
                }
            } else {
                Write-Host "  [OK] No Docker Compose containers running" -ForegroundColor Green
            }
        } finally {
            Pop-Location
        }
    } else {
        Write-Host "[1/3] Skipping Docker Compose (target mode)" -ForegroundColor DarkGray
    }
    
    # 2. Stop Kubernetes (Docker Desktop) if target is not k8s
    if ($TargetMode -ne "k8s") {
        Write-Host ""
        Write-Host "[2/3] Checking Kubernetes (Docker Desktop)..." -ForegroundColor White
        
        try {
            # Save current context
            $savedContext = kubectl config current-context 2>$null
            
            # Check if docker-desktop context exists and has pods
            $contexts = kubectl config get-contexts -o name 2>$null
            if ($contexts -contains "docker-desktop") {
                kubectl config use-context docker-desktop 2>$null | Out-Null
                Start-Sleep -Seconds 1
                
                $podOutput = kubectl get pods -n $Namespace --no-headers 2>$null
                $runningPods = ($podOutput | Where-Object { $_ -match "\bRunning\b" }).Count
                
                if ($runningPods -gt 0) {
                    Write-Host ("  Stopping {0} Kubernetes pods..." -f $runningPods) -ForegroundColor Yellow
                    
                    kubectl config use-context docker-desktop 2>$null | Out-Null
                    
                    # Delete all resources in namespace
                    $yamlFiles = @(
                        "infra/k8s/ingress.yaml",
                        "infra/k8s/web-admin.yaml",
                        "infra/k8s/api-gateway.yaml",
                        "infra/k8s/cluster-service.yaml",
                        "infra/k8s/operations-service.yaml",
                        "infra/k8s/generator-service.yaml",
                        "infra/k8s/project-service.yaml",
                        "infra/k8s/user-service.yaml",
                        "infra/k8s/auth-service.yaml",
                        "infra/k8s/redis.yaml",
                        "infra/k8s/postgres.yaml"
                    )
                    
                    Push-Location $ProjectRoot
                    try {
                        foreach ($file in $yamlFiles) {
                            if (Test-Path $file) {
                                kubectl delete -f $file --ignore-not-found=true --grace-period=0 2>$null | Out-Null
                            }
                        }
                        
                        Start-Sleep -Seconds 3
                        
                        # Verify
                        $remainingPods = (kubectl get pods -n $Namespace --no-headers 2>$null | Where-Object { $_ -match "\bRunning\b" }).Count
                        
                        if ($remainingPods -eq 0) {
                            Write-Success -Message "Kubernetes (Docker Desktop): All pods stopped"
                            $stoppedCount++
                        } else {
                            Write-Warn -Message ("Kubernetes (Docker Desktop): {0} pods still running" -f $remainingPods)
                        }
                    } finally {
                        Pop-Location
                    }
                } else {
                    Write-Host "  [OK] No Kubernetes (Docker Desktop) pods running" -ForegroundColor Green
                }
            } else {
                Write-Host "  [OK] Kubernetes (Docker Desktop) context not found" -ForegroundColor DarkGray
            }
            
            # Restore saved context if it exists
            if ($savedContext -and $savedContext -ne "docker-desktop") {
                # Don't restore yet, let the main flow handle it
            }
        } catch {
            Write-Host "  [OK] Kubernetes (Docker Desktop): Not connected" -ForegroundColor DarkGray
        }
    } else {
        Write-Host "[2/3] Skipping Kubernetes (Docker Desktop) (target mode)" -ForegroundColor DarkGray
    }
    
    # 3. Stop K3D if target is not k3d
    if ($TargetMode -ne "k3d") {
        Write-Host ""
        Write-Host "[3/3] Checking K3D clusters..." -ForegroundColor White
        
        try {
            $k3dContexts = kubectl config get-contexts -o name 2>$null | Where-Object { $_ -like "k3d-*" }
            
            if ($k3dContexts.Count -gt 0) {
                foreach ($ctx in $k3dContexts) {
                    Write-Host ("  Checking K3D context: {0}..." -f $ctx) -ForegroundColor DarkGray
                    
                    kubectl config use-context $ctx 2>$null | Out-Null
                    Start-Sleep -Seconds 1
                    
                    $podOutput = kubectl get pods -n $Namespace --no-headers 2>$null
                    $runningPods = ($podOutput | Where-Object { $_ -match "\bRunning\b" }).Count
                    
                    if ($runningPods -gt 0) {
                        Write-Host ("  Stopping {0} K3D pods..." -f $runningPods) -ForegroundColor Yellow
                        
                        Push-Location $ProjectRoot
                        try {
                            $yamlFiles = @(
                                "infra/k8s/ingress.yaml",
                                "infra/k8s/web-admin.yaml",
                                "infra/k8s/api-gateway.yaml",
                                "infra/k8s/cluster-service.yaml",
                                "infra/k8s/operations-service.yaml",
                                "infra/k8s/generator-service.yaml",
                                "infra/k8s/project-service.yaml",
                                "infra/k8s/user-service.yaml",
                                "infra/k8s/auth-service.yaml",
                                "infra/k8s/redis.yaml",
                                "infra/k8s/postgres.yaml"
                            )
                            
                            foreach ($file in $yamlFiles) {
                                if (Test-Path $file) {
                                    kubectl delete -f $file --ignore-not-found=true --grace-period=0 2>$null | Out-Null
                                }
                            }
                            
                            Start-Sleep -Seconds 3
                            $stoppedCount++
                            Write-Success -Message ("K3D ({0}): Pods stopped" -f $ctx)
                        } finally {
                            Pop-Location
                        }
                    } else {
                        Write-Host ("  [OK] No pods running in {0}" -f $ctx) -ForegroundColor Green
                    }
                }
            } else {
                Write-Host "  [OK] No K3D contexts found" -ForegroundColor DarkGray
            }
        } catch {
            Write-Host "  [OK] K3D: Not available" -ForegroundColor DarkGray
        }
    } else {
        Write-Host "[3/3] Skipping K3D (target mode)" -ForegroundColor DarkGray
    }
    
    # 4. Kill any lingering port-forward processes on our ports
    Write-Host ""
    Write-Host "Cleaning up port-forward processes..." -ForegroundColor White
    
    $portsToCheck = @($Config.FrontendPort, $Config.APIPort)
    foreach ($port in $portsToCheck) {
        if (Test-Port -Port $port) {
            $proc = Get-PortProcess -Port $port
            if ($proc.ProcessName -match "kubectl|port-forward") {
                Write-Host ("  Killing process on port {0} ({1}, PID: {2})" -f $port, $proc.ProcessName, $proc.PID) -ForegroundColor Yellow
                Stop-Process -Id $proc.PID -Force -ErrorAction SilentlyContinue
                Start-Sleep -Milliseconds 500
            }
        }
    }
    
    Write-Host ""
    Write-Host "=====================================================" -ForegroundColor Magenta
    Write-Host ("  CLEANUP COMPLETE: Stopped {0} other mode(s)" -f $stoppedCount) -ForegroundColor Green
    Write-Host "  Mode '$TargetMode' now has exclusive access" -ForegroundColor Cyan
    Write-Host "=====================================================" -ForegroundColor Magenta
    Write-Host ""
    
    return $true
}

# Mode Conflict Detection - Check if other modes are running
function Test-ModeConflicts {
    param([string]$TargetMode)
    
    Write-Host ""
    Write-Host "Checking for conflicts with other deployment modes..." -ForegroundColor Yellow
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    Write-Host ""
    
    $conflicts = @()
    
    # Check Docker Compose
    if ($TargetMode -ne "docker") {
        Push-Location $ProjectRoot
        try {
            $composeOutput = docker compose ps 2>&1
            $hasRunningContainers = $false
            
            foreach ($line in $composeOutput) {
                if ($line -match "\b(Up|running)\b" -and $line -notmatch "level=warning") {
                    $hasRunningContainers = $true
                    break
                }
            }
            
            if ($hasRunningContainers) {
                $conflicts += @{
                    Mode = "Docker Compose"
                    Type = "containers"
                    StopCommand = ".\start.ps1 -Mode docker -Action stop"
                    Description = "Docker containers are running"
                }
                Write-Host ("  [CONFLICT] Docker Compose: Containers detected" -f $TargetMode) -ForegroundColor Red
            } else {
                Write-Host "  [OK] Docker Compose: No containers running" -ForegroundColor Green
            }
        } finally {
            Pop-Location
        }
    }
    
    # Check Kubernetes (Docker Desktop)
    if ($TargetMode -ne "k8s") {
        try {
            $currentContext = kubectl config current-context 2>$null
            
            if ($currentContext -eq "docker-desktop") {
                $podOutput = kubectl get pods -n $Namespace --no-headers 2>$null
                $runningPods = ($podOutput | Where-Object { $_ -match "\bRunning\b" }).Count
                
                if ($runningPods -gt 0) {
                    $conflicts += @{
                        Mode = "Kubernetes (Docker Desktop)"
                        Type = "pods"
                        StopCommand = ".\start.ps1 -Mode k8s -Action stop"
                        Description = "$runningPods pods are running in docker-desktop context"
                    }
                    Write-Host ("  [CONFLICT] Kubernetes (Docker Desktop): {0} pods running" -f $runningPods) -ForegroundColor Red
                } else {
                    Write-Host "  [OK] Kubernetes (Docker Desktop): No pods running" -ForegroundColor Green
                }
            } else {
                Write-Host "  [OK] Kubernetes (Docker Desktop): Not active" -ForegroundColor Green
            }
        } catch {
            Write-Host "  [OK] Kubernetes (Docker Desktop): Not connected" -ForegroundColor Green
        }
    }
    
    # Check K3D
    if ($TargetMode -ne "k3d") {
        try {
            $k3dContexts = kubectl config get-contexts -o name 2>$null | Where-Object { $_ -like "k3d-*" }
            
            foreach ($ctx in $k3dContexts) {
                kubectl config use-context $ctx 2>$null | Out-Null
                Start-Sleep -Seconds 1
                
                $podOutput = kubectl get pods -n $Namespace --no-headers 2>$null
                $runningPods = ($podOutput | Where-Object { $_ -match "\bRunning\b" }).Count
                
                if ($runningPods -gt 0) {
                    $conflicts += @{
                        Mode = "K3D ($ctx)"
                        Type = "pods"
                        StopCommand = ".\start.ps1 -Mode k3d -Action stop"
                        Description = "$runningPods pods are running in K3D cluster"
                    }
                    Write-Host ("  [CONFLICT] K3D ({0}): {1} pods running" -f $ctx, $runningPods) -ForegroundColor Red
                } else {
                    Write-Host ("  [OK] K3D ({0}): No pods running" -f $ctx) -ForegroundColor Green
                }
            }
            
            # Switch back to original context if needed
            if ($Mode -and $Mode -ne "k3d") {
                if ($Mode -eq "k8s") {
                    kubectl config use-context docker-desktop 2>$null | Out-Null
                } elseif ($Mode -eq "docker") {
                    # Keep current or switch to default
                }
            }
        } catch {
            Write-Host "  [OK] K3D: Not available or not connected" -ForegroundColor Green
        }
    }
    
    Write-Host ""
    
    if ($conflicts.Count -gt 0) {
        Write-Warn -Message "Found $($conflicts.Count) conflicting mode(s)!"
        Write-Host ""
        Write-Host "Current mode: $TargetMode" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Conflicting mode(s):" -ForegroundColor Yellow
        
        $index = 1
        foreach ($conflict in $conflicts) {
            Write-Host ("  {0}. {1}" -f $index, $conflict.Mode) -ForegroundColor White
            Write-Host ("     - {0}" -f $conflict.Description) -ForegroundColor DarkGray
            Write-Host ("     - Stop command: {0}" -f $conflict.StopCommand) -ForegroundColor DarkGray
            Write-Host ""
            $index++
        }
        
        Write-Host "Options:" -ForegroundColor Yellow
        Write-Host "  1. Stop all conflicting modes automatically (recommended)" -ForegroundColor Green
        Write-Host "  2. Continue anyway (may cause port conflicts)" -ForegroundColor Red
        Write-Host "  3. Cancel operation" -ForegroundColor Gray
        Write-Host ""
        
        if (-not $Mode -or -not $Action) {
            # Interactive mode
            $choice = Read-Host "Your choice (1-3)"
        } else {
            # Non-interactive mode with -Force flag
            if ($Force) {
                Write-Host "[FORCE] Auto-stopping conflicting modes..." -ForegroundColor Magenta
                $choice = "1"
            } else {
                Write-Host "Use -Force to auto-stop, or run interactively" -ForegroundColor DarkGray
                return $false
            }
        }
        
        switch ($choice) {
            "1" {
                Write-Host "" 
                Write-Host "Stopping conflicting modes..." -ForegroundColor Cyan
                
                # Stop Docker Compose first
                if ($conflicts.Mode -contains "Docker Compose") {
                    Write-Host "  Stopping Docker Compose..." -ForegroundColor White
                    Push-Location $ProjectRoot
                    try {
                        docker compose down 2>&1 | Out-Null
                        Write-Success -Message "Docker Compose stopped"
                    } finally {
                        Pop-Location
                    }
                    Start-Sleep -Seconds 2
                }
                
                # Stop Kubernetes/Docker Desktop
                if ($conflicts.Mode -match "Kubernetes.*Docker Desktop") {
                    Write-Host "  Stopping Kubernetes resources..." -ForegroundColor White
                    & "$ProjectRoot\start.ps1" -Mode k8s -Action stop 2>$null
                    Write-Success -Message "Kubernetes resources removed"
                    Start-Sleep -Seconds 2
                }
                
                # Stop K3D
                if ($conflicts.Mode -match "^K3D") {
                    Write-Host "  Stopping K3D resources..." -ForegroundColor White
                    & "$ProjectRoot\start.ps1" -Mode k3d -Action stop 2>$null
                    Write-Success -Message "K3D resources removed"
                    Start-Sleep -Seconds 2
                }
                
                Write-Success -Message "All conflicting modes stopped!"
                Write-Host ""
                return $true
            }
            "2" {
                Write-Warn -Message "Continuing despite conflicts..."
                Write-Host "  This may cause unexpected behavior!" -ForegroundColor DarkRed
                Write-Host ""
                return $true
            }
            default {
                Write-Info -Message "Cancelled by user"
                return $false
            }
        }
    } else {
        Write-Success -Message "No conflicts detected - safe to proceed!"
        return $true
    }
}

# Verify we're connected to the correct cluster
function Confirm-CorrectCluster {
    param([string]$ExpectedMode)
    
    Write-Host ""
    Write-Host "Verifying Kubernetes connection..." -ForegroundColor Yellow
    
    $currentContext = kubectl config current-context 2>$null
    Write-Host ("  Current context: {0}" -f $currentContext) -ForegroundColor White
    
    $expectedContext = switch ($ExpectedMode) {
        "k8s"    { "docker-desktop" }
        "k3d"    { "k3d-gen-platform-test" }
        default   { $null }
    }
    
    if ($expectedContext -and $currentContext -ne $expectedContext) {
        Write-Warn -Message ("Expected context '{0}' but connected to '{1}'" -f $expectedContext, $currentContext)
        Write-Host ""
        Write-Host "This means you might be deploying to the wrong cluster!" -ForegroundColor Red
        Write-Host ""
        
        if (-not $Mode -or -not $Action) {
            $choice = Read-Host "Switch to correct context? (y/n)"
        } elseif ($Force) {
            $choice = "y"
        } else {
            $choice = "n"
        }
        
        if ($choice -eq 'y' -or $choice -eq 'Y') {
            Write-Host ("Switching to {0}..." -f $expectedContext) -ForegroundColor Cyan
            
            if ($ExpectedMode -eq "k8s") {
                Switch-KubeContext-K8S
            } elseif ($ExpectedMode -eq "k3d") {
                Switch-KubeContext-K3D
            }
            
            $newContext = kubectl config current-context 2>$null
            if ($newContext -eq $expectedContext) {
                Write-Success -Message ("Now connected to: {0}" -f $newContext)
                return $true
            } else {
                Write-Err -Message "Failed to switch context"
                return $false
            }
        } else {
            Write-Warn -Message "Using wrong context may cause issues!"
            Write-Host ""
            $choice2 = Read-Host "Continue anyway? (y/n)"
            if ($choice2 -ne 'y' -and $choice2 -ne 'Y') {
                return $false
            }
            return $true
        }
    } else {
        Write-Success -Message ("Correctly connected to: {0}" -f $currentContext)
        return $true
    }
}

# Docker Image Building
function Build-DockerImages {
    Write-Host ""
    Write-Host "Building Docker Images..." -ForegroundColor Yellow
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    Write-Host ""
    
    $total = $Services.Count
    $current = 0
    
    foreach ($svc in $Services) {
        $current++
        
        Write-Host ("[{0}/{1}] Building {2,-25} [" -f $current, $total, $svc.Name) -NoNewline -ForegroundColor White
        
        # Web-admin 需要根据目标模式设置 DEPLOY_MODE
        if ($svc.Name -eq "web-admin") {
            $deployMode = if ($Mode -eq "docker") { "docker" } else { "k8s" }
            $buildCmd = "docker build -t generator-platform/$($svc.Name):latest --build-arg DEPLOY_MODE=$deployMode -f $($svc.Dockerfile) ."
        } else {
            $buildCmd = "docker build -t generator-platform/$($svc.Name):latest -f $($svc.Dockerfile) ."
        }
        
        Invoke-Expression $buildCmd 2>&1 | Out-Null
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host " OK]" -ForegroundColor Green
        } else {
            Write-Host " FAIL]" -ForegroundColor Red
            return $false
        }
    }
    
    Write-Host ""
    Write-Success -Message "All images built successfully!"
    return $true
}

# Mode: Docker Compose
function Deploy-Docker {
    Write-Host ""
    Write-Host "Deploying with Docker Compose..." -ForegroundColor Yellow
    
    if ($BuildImages) {
        if (-not (Build-DockerImages)) { return }
    }
    
    Push-Location $ProjectRoot
    
    try {
        $envContent = "FRONTEND_PORT=$($Config.FrontendPort)`nAPI_PORT=$($Config.APIPort)`nDB_PORT=$($Config.DBPort)`nREDIS_PORT=$($Config.RedisPort)"
        
        Set-Content -Path ".env.deploy" -Value $envContent -Encoding UTF8
        
        Write-Host ("Starting services on port {0}..." -f $Config.FrontendPort) -ForegroundColor White
        docker compose --env-file .env.deploy up -d --build 2>&1 | Out-Null
        
        Remove-Item ".env.deploy" -ErrorAction SilentlyContinue
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success -Message "Docker Compose deployment complete!"
            Start-Sleep -Seconds 5
            Show-DockerStatus
        } else {
            Write-Err -Message "Docker Compose failed!"
        }
    } finally {
        Pop-Location
    }
}

function Stop-Docker {
    Write-Host ""
    Write-Host "Stopping Docker Compose services..." -ForegroundColor Yellow
    
    Push-Location $ProjectRoot
    try {
        docker compose down 2>&1 | Out-Null
        Write-Success -Message "All Docker services stopped"
    } finally {
        Pop-Location
    }
}

function Show-DockerStatus {
    Write-Host ""
    Write-Host "Service Status:" -ForegroundColor Yellow
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    
    # Get docker compose output and filter out warnings
    $output = docker compose ps 2>&1
    $hasRunningServices = $false
    $serviceLines = @()
    
    foreach ($line in $output) {
        # Skip Docker warning messages about version attribute
        if ($line -match "level=warning|attribute.*version.*obsolete") {
            continue
        }
        
        Write-Host $line
        
        # Check if there are any running services (look for "Up" or "running")
        if ($line -match "\b(Up|running)\b") {
            $hasRunningServices = $true
        }
    }
    
    # Only show access URLs if services are actually running
    if ($hasRunningServices) {
        Write-Host ""
        Write-Host "Access URLs:" -ForegroundColor Yellow
        Write-Host ("  Frontend: http://localhost:{0}" -f $Config.FrontendPort) -ForegroundColor Green
        Write-Host ("  API:      http://localhost:{0}/api/v1" -f $Config.APIPort) -ForegroundColor Cyan
    } else {
        Write-Host ""
        Write-Warn -Message "No services are currently running"
        Write-Host "  Tip: Run 'deploy' action to start services" -ForegroundColor DarkGray
    }
    
    Write-Host ""
}

function Show-DockerLogs {
    param([string]$Service = "")
    
    Push-Location $ProjectRoot
    try {
        if ($Service) {
            docker compose logs -f --tail=100 $Service
        } else {
            docker compose logs -f --tail=50
        }
    } finally {
        Pop-Location
    }
}

# Mode: Kubernetes / K3D
function Switch-KubeContext-K8S {
    Write-Host "Switching to Docker Desktop K8s context..." -ForegroundColor Yellow
    
    kubectl config use-context docker-desktop 2>&1 | Out-Null
    Start-Sleep -Seconds 2
    
    try {
        kubectl cluster-info 2>&1 | Out-Null
        Write-Success -Message "Connected to Docker Desktop K8s"
    } catch {
        Write-Err -Message "Cannot connect to Desktop K8s"
        return $false
    }
    return $true
}

function Switch-KubeContext-K3D {
    Write-Host "Switching to K3D context..." -ForegroundColor Yellow
    
    $contextName = "k3d-gen-platform-test"
    $contexts = kubectl config get-contexts -o name 2>$null
    
    if ($contexts -notcontains $contextName) {
        Write-Warn -Message "K3D cluster not found..."
        
        $hasK3D = Get-Command k3d -ErrorAction SilentlyContinue
        if ($hasK3D) {
            Write-Host "Creating K3D cluster..." -ForegroundColor White
            
            # Check and warn about port conflicts before creating
            $k3dPorts = @(5432, 6379, 30080, 30443, 6443)
            $portConflicts = @()
            
            foreach ($port in $k3dPorts) {
                if (Test-Port -Port $port) {
                    $proc = Get-PortProcess -Port $port
                    $portConflicts += @{ Port=$port; Process=$proc.ProcessName; PID=$proc.PID }
                    Write-Warn -Message ("Port {0} is in use by {1} (PID: {2})" -f $port, $proc.ProcessName, $proc.PID)
                }
            }
            
            if ($portConflicts.Count -gt 0) {
                Write-Host ""
                Write-Host "K3D requires these ports. Options:" -ForegroundColor Yellow
                Write-Host "  1. Stop conflicting processes automatically" -ForegroundColor Green
                Write-Host "  2. Continue anyway (may fail)" -ForegroundColor Red
                Write-Host "  3. Cancel" -ForegroundColor Gray
                Write-Host ""
                
                $choice = Read-Host "Your choice (1-3)"
                
                switch ($choice) {
                    "1" {
                        foreach ($conflict in $portConflicts) {
                            Write-Host ("Stopping {0} on port {1}..." -f $conflict.Process, $conflict.Port) -ForegroundColor White
                            Stop-Process -Id $conflict.PID -Force -ErrorAction SilentlyContinue
                        }
                        Start-Sleep -Seconds 2
                    }
                    "2" {
                        Write-Warn -Message "Continuing despite conflicts..."
                    }
                    default {
                        Write-Info -Message "Cancelled"
                        return $false
                    }
                }
            }
            
            Write-Host "Creating K3D cluster 'gen-platform-test'..." -ForegroundColor Cyan
            k3d cluster create gen-platform-test --agents 2 --port "30080:80@loadbalancer" --port "30443:443@loadbalancer" 2>&1 | Out-Null
            
            if ($LASTEXITCODE -ne 0) {
                Write-Err -Message "Failed to create K3D cluster"
                return $false
            }
            
            Write-Success -Message "K3D cluster created!"
            Start-Sleep -Seconds 5
        } else {
            Write-Err -Message "K3D CLI not installed but cluster context not found"
            return $false
        }
    }
    
    # Fix K3D API server port (common issue with k3d)
    Write-Host "Checking K3D API server configuration..." -ForegroundColor DarkGray
    
    try {
        $lbOutput = docker port k3d-gen-platform-test-serverlb 2>&1
        
        foreach ($line in $lbOutput) {
            if ($line -match "6443") {
                $parts = $line -split ":"
                $actualPort = $parts[-1]
                
                Write-Host ("Fixing API server port to: {0}" -f $actualPort) -ForegroundColor DarkGray
                
                # Use proper argument format for kubectl
                $serverUrl = "https://127.0.0.1:$actualPort"
                & kubectl config set-cluster $contextName --server $serverUrl 2>&1 | Out-Null
                
                break
            }
        }
    } catch {
        Write-Warn -Message "Could not auto-detect K3D API port, using default configuration"
    }
    
    # Switch to the context
    Write-Host "Switching to K3D context..." -ForegroundColor DarkGray
    & kubectl config use-context $contextName 2>&1 | Out-Null
    Start-Sleep -Seconds 2
    
    # Test connection
    try {
        $clusterInfo = kubectl cluster-info 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success -Message "Connected to K3D cluster"
            return $true
        } else {
            throw "kubectl cluster-info failed"
        }
    } catch {
        Write-Err -Message "Cannot connect to K3D cluster"
        Write-Host "" -ForegroundColor Gray
        Write-Host "Troubleshooting:" -ForegroundColor Yellow
        Write-Host "  1. Check if k3d cluster is running: docker ps | findstr k3d" -ForegroundColor DarkGray
        Write-Host "  2. Restart k3d cluster: k3d cluster start gen-platform-test" -ForegroundColor DarkGray
        Write-Host "  3. Delete and recreate: k3d cluster delete gen-platform-test" -ForegroundColor DarkGray
        Write-Host "" -ForegroundColor Gray
        return $false
    }
}

function Deploy-Kubernetes {
    Write-Host ""
    Write-Host "Deploying to Kubernetes..." -ForegroundColor Yellow
    
    if ($BuildImages) {
        if (-not (Build-DockerImages)) { return }
    }
    
    Push-Location $ProjectRoot
    
    try {
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
        
        $total = $yamlFiles.Count
        $current = 0
        
        foreach ($file in $yamlFiles) {
            if (Test-Path $file) {
                $current++
                Write-Host ("[{0}/{1}] Applying {2}" -f $current, $total, (Split-Path $file -Leaf)) -ForegroundColor Cyan
                
                kubectl apply -f $file 2>&1 | Out-Null
            }
        }
        
        Write-Host ""
        Write-Success -Message "Kubernetes deployment complete!"
        Write-Host "Waiting for Pods to start..." -ForegroundColor Yellow
        Start-Sleep -Seconds 10
        Show-KubernetesStatus
        
        Write-Host ""
        Write-Warn -Message "Run 'port' action to access services locally"
    } finally {
        Pop-Location
    }
}

function Stop-Kubernetes {
    Write-Host ""
    Write-Host "Removing Kubernetes resources..." -ForegroundColor Yellow
    
    Push-Location $ProjectRoot
    try {
        $yamlFiles = @(
            "infra/k8s/ingress.yaml",
            "infra/k8s/web-admin.yaml",
            "infra/k8s/api-gateway.yaml",
            "infra/k8s/cluster-service.yaml",
            "infra/k8s/operations-service.yaml",
            "infra/k8s/generator-service.yaml",
            "infra/k8s/project-service.yaml",
            "infra/k8s/user-service.yaml",
            "infra/k8s/auth-service.yaml",
            "infra/k8s/redis.yaml",
            "infra/k8s/postgres.yaml",
            "infra/k8s/rbac.yaml",
            "infra/k8s/namespace.yaml"
        )
        
        foreach ($file in $yamlFiles) {
            if (Test-Path $file) {
                kubectl delete -f $file --ignore-not-found=true 2>&1 | Out-Null
            }
        }
        
        Write-Success -Message "All Kubernetes resources removed"
    } finally {
        Pop-Location
    }
}

function Show-KubernetesStatus {
    Write-Host ""
    Write-Host ("Pod Status ({0}):" -f $Namespace) -ForegroundColor Yellow
    Write-Host "-----------------------------------------------------" -ForegroundColor DarkGray
    
    $podOutput = kubectl get pods -n $Namespace -o wide 2>&1
    Write-Host $podOutput
    
    # Check if there are any running pods
    $hasRunningPods = $false
    foreach ($line in $podOutput) {
        if ($line -match "\b(Running)\b") {
            $hasRunningPods = $true
            break
        }
    }
    
    Write-Host ""
    Write-Host "Services:" -ForegroundColor Yellow
    kubectl get svc -n $Namespace 2>&1
    
    if ($hasRunningPods) {
        Write-Host ""
        Write-Host "Access URLs (after port-forward):" -ForegroundColor Yellow
        Write-Host ("  Frontend: http://localhost:{0}" -f $Config.FrontendPort) -ForegroundColor Green
        Write-Host ("  API:      http://localhost:{0}/api/v1" -f $Config.APIPort) -ForegroundColor Cyan
        Write-Host ""
        Write-Host "  Run 'port' action to start port forwarding" -ForegroundColor DarkGray
    } else {
        Write-Host ""
        Write-Warn -Message "No pods are currently running"
        Write-Host "  Tip: Run 'deploy' action to start services" -ForegroundColor DarkGray
    }
    
    Write-Host ""
}

function Open-Browser {
    param([string]$TargetMode = "docker")
    
    $frontendUrl = ("http://localhost:{0}" -f $Config.FrontendPort)
    
    Write-Host ""
    Write-Host "Opening browser..." -ForegroundColor Cyan
    Write-Host ""
    Write-Host ("  URL: {0}" -f $frontendUrl) -ForegroundColor Green
    Write-Host ("  Mode: {0}" -f $TargetMode.ToUpper()) -ForegroundColor White
    Write-Host ""
    
    # Check if port is in use (service is running)
    if (-not (Test-Port -Port $Config.FrontendPort)) {
        Write-Warn -Message "Port $($Config.FrontendPort) is not in use!"
        Write-Host "  Make sure services are running first:" -ForegroundColor DarkGray
        Write-Host ("    .\start.ps1 -Mode {0} -Action deploy" -f $TargetMode) -ForegroundColor Yellow
        Write-Host ""
        
        $choice = Read-Host "Open anyway? (y/n)"
        if ($choice -ne 'y' -and $choice -ne 'Y') {
            Write-Info -Message "Cancelled"
            return
        }
    }
    
    # Copy to clipboard as backup
    try {
        $frontendUrl | Set-Clipboard
        Write-Success -Message "URL copied to clipboard (Ctrl+V to paste)"
    } catch {
        # Clipboard might not work, that's ok
    }
    
    # Try to open browser
    try {
        Start-Process $frontendUrl
        Write-Success -Message "Browser launched!"
        Write-Host ""
        Write-Host "  Login: admin / admin123" -ForegroundColor Magenta
        Write-Host ""
    } catch {
        Write-Warn -Message "Could not open browser automatically"
        Write-Host ""
        Write-Host "Manual steps:" -ForegroundColor Yellow
        Write-Host "  1. Copy this URL:" -ForegroundColor Gray
        Write-Host ("     {0}" -f $frontendUrl) -ForegroundColor Green
        Write-Host "  2. Paste in browser address bar (Ctrl+V)" -ForegroundColor Gray
        Write-Host "  3. Press Enter" -ForegroundColor Gray
        Write-Host ""
    }
}

function Show-KubernetesLogs {
    param([string]$Service = "all")
    
    if ($Service -eq "all") {
        kubectl logs -n $Namespace -l app --tail=50 -f --max-log-requests=10 2>&1
    } else {
        kubectl logs -n $Namespace -l app=$Service -f --tail=100 2>&1
    }
}

function Start-PortForward-Kubernetes {
    param([string]$TargetMode = "k8s")
    
    Write-Host ""
    Write-Host "Starting Port Forwarding for $TargetMode..." -ForegroundColor Magenta
    Write-Host "(Press Ctrl+C to stop)" -ForegroundColor DarkGray
    Write-Host ""
    
    # Check and handle port conflicts
    $frontendConflicted = $false
    $apiConflicted = $false
    
    if (Test-Port -Port $Config.FrontendPort) {
        $proc = Get-PortProcess -Port $Config.FrontendPort
        
        # Check if it's a kubectl port-forward process
        if ($proc.ProcessName -match "kubectl|port-forward") {
            Write-Warn -Message ("Port {0} is in use by another kubectl port-forward!" -f $Config.FrontendPort)
            Write-Host ("  Process: {0} (PID: {1})" -f $proc.ProcessName, $proc.PID) -ForegroundColor Yellow
            Write-Host ""
            Write-Host "This means you have another mode's port-forward running." -ForegroundColor White
            Write-Host ""
            Write-Host "Options:" -ForegroundColor Yellow
            Write-Host "  1. Stop the old process automatically (recommended)" -ForegroundColor Green
            Write-Host "  2. Cancel and keep old process" -ForegroundColor Red
            Write-Host ""
            
            $choice = Read-Host "Your choice (1-2)"
            
            switch ($choice) {
                "1" {
                    Write-Host ("Stopping old kubectl process on port {0}..." -f $Config.FrontendPort) -ForegroundColor Cyan
                    Stop-Process -Id $proc.PID -Force -ErrorAction SilentlyContinue
                    Start-Sleep -Seconds 2
                    
                    if (-not (Test-Port -Port $Config.FrontendPort)) {
                        Write-Success -Message "Old process stopped successfully"
                    } else {
                        Write-Err -Message "Failed to stop old process"
                        return
                    }
                }
                default {
                    Write-Info -Message "Cancelled - keeping old process"
                    return
                }
            }
        } else {
            # Non-kubectl process using the port
            Write-Warn -Message ("Port {0} is already in use by {1}!" -f $Config.FrontendPort, $proc.ProcessName)
            Write-Host ("  PID: {0}" -f $proc.PID) -ForegroundColor Yellow
            Write-Host ""
            Write-Host "This is not a kubectl process. Options:" -ForegroundColor White
            Write-Host "  1. Force stop anyway (may affect other applications)" -ForegroundColor DarkRed
            Write-Host "  2. Use a different port" -ForegroundColor Yellow
            Write-Host "  3. Cancel" -ForegroundColor Gray
            Write-Host ""
            
            $choice = Read-Host "Your choice (1-3)"
            
            switch ($choice) {
                "1" {
                    Stop-Process -Id $proc.PID -Force -ErrorAction SilentlyContinue
                    Start-Sleep -Seconds 2
                    Write-Success -Message "Process stopped"
                }
                "2" {
                    $newPort = Read-Host "Enter new frontend port"
                    if ($newPort -match '^\d+$') {
                        $Config.FrontendPort = [int]$newPort
                        Write-Success -Message ("Using port {0}" -f $newPort)
                    }
                }
                default {
                    Write-Info -Message "Cancelled"
                    return
                }
            }
        }
    }
    
    # Same check for API port
    if (Test-Port -Port $Config.APIPort) {
        $proc = Get-PortProcess -Port $Config.APIPort
        
        if ($proc.ProcessName -match "kubectl|port-forward") {
            Write-Warn -Message ("Port {0} is in use by kubectl, stopping..." -f $Config.APIPort)
            Stop-Process -Id $proc.PID -Force -ErrorAction SilentlyContinue
            Start-Sleep -Seconds 1
        } else {
            Write-Warn -Message ("Port {0} is in use by {1} (PID: {2})" -f $Config.APIPort, $proc.ProcessName, $proc.PID)
            Write-Host "  Skipping API port-forward. You can still access the frontend." -ForegroundColor DarkGray
            $apiConflicted = $true
        }
    }
    
    # Start port forwarding
    Write-Host ""
    Write-Host ("Mode: {0}" -f $TargetMode.ToUpper()) -ForegroundColor Cyan
    Write-Host ("Frontend -> http://localhost:{0}" -f $Config.FrontendPort) -ForegroundColor Green
    
    $pfJob = Start-Job -ScriptBlock {
        param($ns, $fp) 
        kubectl port-forward -n $ns svc/web-admin ${fp}:3000
    } -ArgumentList $Namespace, $Config.FrontendPort
    
    Start-Sleep -Seconds 3
    
    if (-not $apiConflicted) {
        Write-Host ("API      -> http://localhost:{0}/api/v1" -f $Config.APIPort) -ForegroundColor Cyan
        kubectl port-forward -n $Namespace svc/api-gateway "$($Config.APIPort):8080" 2>&1
    } else {
        Write-Host "API: Not forwarded (port in use)" -ForegroundColor DarkGray
        Write-Host "  You can access API through the frontend proxy." -ForegroundColor DarkGray
        
        # Wait for frontend job only
        while ($pfJob.State -eq 'Running') {
            Start-Sleep -Seconds 1
        }
    }
    
    if ($pfJob) {
        Remove-Job $pfJob -Force -ErrorAction SilentlyContinue
    }
}

# Main Execution Logic
if ($Help) {
    Show-Help
    exit 0
}

# Non-interactive mode (parameters provided)
if ($Mode -and $Action) {
    Show-Banner
    
    Write-Host "Mode: " -NoNewline -ForegroundColor Gray
    switch ($Mode) {
        "docker" { Write-Host "Docker Compose" -ForegroundColor Green }
        "k8s"    { Write-Host "Kubernetes (Docker Desktop)" -ForegroundColor Blue }
        "k3d"    { Write-Host "K3D (Lightweight)" -ForegroundColor Magenta }
    }
    Write-Host "Action: $Action" -ForegroundColor White
    Write-Host ""
    
    if ($Action -eq "check-ports") {
        Show-PortStatusDashboard
        exit 0
    }
    
    if ($Action -eq "cleanup") {
        & "$ProjectRoot\cleanup-project.ps1" -Force
        exit 0
    }
    
    if (-not (Test-Prerequisites)) { exit 1 }
    
    if ($Mode -eq "k8s") {
        if (-not (Switch-KubeContext-K8S)) { exit 1 }
    } elseif ($Mode -eq "k3d") {
        if (-not (Switch-KubeContext-K3D)) { exit 1 }
    }
    
    switch ($Action) {
        "deploy" {
            # CRITICAL: Force stop all other modes before deploying
            Write-Host ""
            Write-Host "[PRE-DEPLOY] Ensuring exclusive mode access..." -ForegroundColor Magenta
            Stop-AllOtherModes -TargetMode $Mode
            
            switch ($Mode) {
                "docker" { Deploy-Docker }
                "k8s"    { Deploy-Kubernetes }
                "k3d"    { Deploy-Kubernetes }
            }
            Show-DeploymentComplete -DeployedMode $Mode
        }
        "status" {
            switch ($Mode) {
                "docker" { Show-DockerStatus }
                "k8s"    { Show-KubernetesStatus }
                "k3d"    { Show-KubernetesStatus }
            }
            Write-Host "`nPress any key to continue..." -ForegroundColor DarkGray
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        }
        "stop" {
            switch ($Mode) {
                "docker" { Stop-Docker }
                "k8s"    { Stop-Kubernetes }
                "k3d"    { Stop-Kubernetes }
            }
            Write-Success -Message "Services stopped"
            Start-Sleep -Seconds 1
        }
        "logs" {
            switch ($Mode) {
                "docker" { Show-DockerLogs }
                "k8s"    { Show-KubernetesLogs }
                "k3d"    { Show-KubernetesLogs }
            }
        }
        "port" {
            switch ($Mode) {
                "docker" { 
                    Write-Warn -Message "Docker doesn't need port forwarding"
                    Write-Host ("Access directly via: http://localhost:{0}" -f $Config.FrontendPort) -ForegroundColor Green
                    Start-Sleep -Seconds 2
                }
                "k8s"    { Start-PortForward-Kubernetes -TargetMode "k8s" }
                "k3d"    { Start-PortForward-Kubernetes -TargetMode "k3d" }
            }
        }
        "open" {
            Open-Browser -TargetMode $Mode
        }
    }
    
    exit 0
}

# Interactive Mode (no parameters provided)
Write-Host "`nLaunching interactive mode...`n" -ForegroundColor Cyan
Start-Sleep -Seconds 1

while ($true) {
    $selectedMode = Select-ModeInteractive
    $selectedAction = Select-ActionInteractive -SelectedMode $selectedMode
    
    if ($selectedAction -eq "back") {
        continue
    }
    
    $Mode = $selectedMode
    $Action = $selectedAction
    
    Show-Banner
    
    Write-Host "You selected:" -ForegroundColor White
    Write-Host "  Mode:   " -NoNewline -ForegroundColor Gray
    Write-Host $selectedMode.ToUpper() -ForegroundColor $(switch ($selectedMode) { 
        "docker" { "Green" }; "k8s" { "Blue" }; "k3d" { "Magenta" } 
    })
    Write-Host "  Action: " -NoNewline -ForegroundColor Gray
    Write-Host $selectedAction.ToUpper() -ForegroundColor Yellow
    Write-Host ""
    
    if ($selectedAction -eq "check-ports") {
        Show-PortStatusDashboard
        continue
    }
    
    if ($selectedAction -eq "cleanup") {
        & "$ProjectRoot\cleanup-project.ps1" -Force
        Start-Sleep -Seconds 2
        continue
    }
    
    if (-not (Test-Prerequisites)) { continue }
    
    # For K8S/K3D modes, verify correct cluster connection after context switch
    # (Mode conflicts already handled by Stop-AllOtherModes in deploy case)
    
    if ($selectedMode -eq "k8s") {
        if (-not (Switch-KubeContext-K8S)) { continue }
        # After switching, verify we're on the right cluster
        if ($selectedAction -eq "deploy") {
            if (-not (Confirm-CorrectCluster -ExpectedMode "k8s")) { continue }
        }
    } elseif ($selectedMode -eq "k3d") {
        if (-not (Switch-KubeContext-K3D)) { continue }
        # After switching, verify we're on the right cluster
        if ($selectedAction -eq "deploy") {
            if (-not (Confirm-CorrectCluster -ExpectedMode "k3d")) { continue }
        }
    }
    
    switch ($selectedAction) {
        "deploy" {
            # CRITICAL: Force stop all other modes before deploying
            Write-Host ""
            Write-Host "[PRE-DEPLOY] Ensuring exclusive mode access..." -ForegroundColor Magenta
            Stop-AllOtherModes -TargetMode $selectedMode
            
            switch ($selectedMode) {
                "docker" { Deploy-Docker }
                "k8s"    { Deploy-Kubernetes }
                "k3d"    { Deploy-Kubernetes }
            }
            Show-DeploymentComplete -DeployedMode $selectedMode
        }
        "status" {
            switch ($selectedMode) {
                "docker" { Show-DockerStatus }
                "k8s"    { Show-KubernetesStatus }
                "k3d"    { Show-KubernetesStatus }
            }
            Write-Host "`nPress any key to continue..." -ForegroundColor DarkGray
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        }
        "stop" {
            switch ($selectedMode) {
                "docker" { Stop-Docker }
                "k8s"    { Stop-Kubernetes }
                "k3d"    { Stop-Kubernetes }
            }
            Write-Success -Message "Services stopped"
            Start-Sleep -Seconds 2
        }
        "logs" {
            switch ($selectedMode) {
                "docker" { Show-DockerLogs }
                "k8s"    { Show-KubernetesLogs }
                "k3d"    { Show-KubernetesLogs }
            }
        }
        "port" {
            switch ($selectedMode) {
                "docker" { 
                    Write-Warn -Message "Docker doesn't need port forwarding"
                    Write-Host ("Access: http://localhost:{0}" -f $Config.FrontendPort) -ForegroundColor Green
                    Start-Sleep -Seconds 2
                }
                "k8s"    { Start-PortForward-Kubernetes -TargetMode "k8s" }
                "k3d"    { Start-PortForward-Kubernetes -TargetMode "k3d" }
            }
        }
        "open" {
            Open-Browser -TargetMode $selectedMode
        }
    }
}
