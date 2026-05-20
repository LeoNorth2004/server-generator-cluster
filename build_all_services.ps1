$ErrorActionPreference = "Continue"
Set-Location "c:\Data\SystemData\Desktop\generator-project\project1"

$services = @(
    @{name="auth-service"; path="apps/authentication-service"},
    @{name="user-service"; path="apps/user-service"},
    @{name="project-service"; path="apps/project-service"},
    @{name="generator-service"; path="apps/generator-service"},
    @{name="operations-service"; path="apps/operations-service"},
    @{name="cluster-service"; path="apps/cluster-service"},
    @{name="web-admin"; path="apps/web-admin"}
)

foreach ($svc in $services) {
    Write-Host "`n=== Building $($svc.name) ===" -ForegroundColor Cyan
    $dockerfile = "$($svc.path)\Dockerfile"
    
    if (Test-Path $dockerfile) {
        docker build -f $dockerfile -t "generator-platform/$($svc.name):latest" . 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "[OK] $($svc.name) built successfully" -ForegroundColor Green
        } else {
            Write-Host "[FAIL] $($svc.name) build failed" -ForegroundColor Red
        }
    } else {
        Write-Host "[SKIP] Dockerfile not found: $dockerfile" -ForegroundColor Yellow
    }
}

Write-Host "`n`n========================================" -ForegroundColor Yellow
Write-Host "  Build Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Yellow

docker images --format "{{.Repository}}:{{.Tag}} ({{.Size}})" | findstr generator-platform | findstr latest
