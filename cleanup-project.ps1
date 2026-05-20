param(
    [switch]$Force,
    [switch]$WhatIf
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "=====================================================" -ForegroundColor Cyan
Write-Host "  Generator Platform - Project Cleanup Tool" -ForegroundColor Cyan
Write-Host "=====================================================" -ForegroundColor Cyan
Write-Host ""

if (-not $Force) {
    Write-Host "[WARNING] This will delete temporary files and old versions!" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Use -Force to confirm deletion" -ForegroundColor Yellow
    Write-Host "Use -WhatIf to preview what will be deleted" -ForegroundColor Yellow
    Write-Host ""
    
    if ($WhatIf) {
        $Force = $true
    } else {
        exit 0
    }
}

$deletedCount = 0
$errorCount = 0

function Remove-FileSafe {
    param([string]$Path, [string]$Reason)
    
    if (Test-Path $Path) {
        if ($WhatIf) {
            Write-Host "  [PREVIEW] Would delete: $Path" -ForegroundColor DarkGray
            Write-Host "             Reason: $Reason" -ForegroundColor DarkGray
        } else {
            try {
                Remove-Item $Path -Force -Recurse -ErrorAction Stop
                Write-Host "  [DELETED] $Path" -ForegroundColor Green
                Write-Host "           Reason: $Reason" -ForegroundColor DarkGray
                $script:deletedCount++
            } catch {
                Write-Host "  [ERROR] Failed to delete: $Path" -ForegroundColor Red
                Write-Host "          Error: $_" -ForegroundColor DarkRed
                $script:errorCount++
            }
        }
    }
}

Write-Host "[*] Scanning for cleanup targets..." -ForegroundColor Yellow
Write-Host ""

# 1. Tar backup files in root
Write-Host "[1/6] Tar backup files..." -ForegroundColor White
$tarFiles = @(
    @{Path="cluster-service-v9.tar"; Reason="Old versioned backup (v9)"},
    @{Path="webadmin-v10.tar"; Reason="Old versioned backup (v10)"},
    @{Path="cluster-v10.tar"; Reason="Old versioned backup (v10)"},
    @{Path="webadmin.tar"; Reason="Unnamed backup file"},
    @{Path="cluster.tar"; Reason="Unnamed backup file"},
    @{Path="ops.tar"; Reason="Unnamed backup file"}
)

foreach ($file in $tarFiles) {
    Remove-FileSafe -Path $file.Path -Reason $file.Reason
}

# 2. Temporary documents
Write-Host ""
Write-Host "[2/6] Temporary document files..." -ForegroundColor White
$docFiles = @(
    @{Path="RESEARCH_PAPER.docx"; Reason="Temporary research document"},
    @{Path="RESEARCH_PAPER.md"; Reason="Temporary research document"},
    @{Path="md_to_word.cjs"; Reason="Utility script (not part of core project)"}
)

foreach ($file in $docFiles) {
    Remove-FileSafe -Path $file.Path -Reason $file.Reason
}

# 3. Old Docker images with version tags
Write-Host ""
Write-Host "[3/6] Old Docker image tags..." -ForegroundColor White
if (-not $WhatIf) {
    $oldImages = docker images generator-platform/* --format "{{.Repository}}:{{.Tag}}" | Where-Object { $_ -match "-(v\d+-new|v\d+-fix|v\d+-nobutton)$" }
    
    foreach ($image in $oldImages) {
        try {
            docker rmi $image 2>&1 | Out-Null
            Write-Host "  [REMOVED] Image: $image" -ForegroundColor Green
            $script:deletedCount++
        } catch {
            Write-Host "  [SKIP] Image in use: $image" -ForegroundColor Yellow
        }
    }
} else {
    $oldImages = docker images generator-platform/* --format "{{.Repository}}:{{.Tag}}" | Where-Object { $_ -match "-(v\d+-new|v\d+-fix|v\d+-nobutton)$" }
    foreach ($image in $oldImages) {
        Write-Host "  [PREVIEW] Would remove image: $image" -ForegroundColor DarkGray
        $script:deletedCount++
    }
}

# 4. Redundant batch files in root (keep start.bat)
Write-Host ""
Write-Host "[4/6] Redundant root-level scripts..." -ForegroundColor White
$redundantScripts = @(
    @{Path="test_system.bat"; Reason="Superseded by start-platform.ps1"},
    @{Path="start_access.bat"; Reason="Superseded by start-platform.ps1"},
    @{Path="deploy_k8s.bat"; Reason="Superseded by start-platform.ps1"},
    @{Path="cleanup_and_prepare.bat"; Reason="One-time setup script"},
    @{Path="build_all.bat"; Reason="Superseded by Makefile"}
)

foreach ($script in $redundantScripts) {
    Remove-FileSafe -Path $script.Path -Reason $script.Reason
}

# 5. Redundant scripts in scripts/ directory (keep essential ones)
Write-Host ""
Write-Host "[5/6] Redundant scripts in /scripts..." -ForegroundColor White
$redundantScriptFiles = @(
    @{Path="scripts/k8s-deploy.ps1"; Reason="Superseded by start-platform.ps1"},
    @{Path="scripts/k8s-deploy.sh"; Reason="Superseded by start-platform.ps1"},
    @{Path="scripts/deploy-k8s.ps1"; Reason="Superseded by start-platform.ps1"},
    @{Path="scripts/deploy-k8s.sh"; Reason="Superseded by start-platform.ps1"},
    @{Path="scripts/deploy.sh"; Reason="Superseded by start-platform.ps1"},
    @{Path="scripts/k8s-simple-deploy.ps1"; Reason="Superseded by start-platform.ps1"},
    @{Path="scripts/build.ps1"; Reason="Superseded by Makefile"},
    @{Path="scripts/start-docker.ps1"; Reason="Superseded by start-platform.ps1"},
    @{Path="scripts/start-k8s.ps1"; Reason="Superseded by start-platform.ps1"},
    @{Path="scripts/start-local.ps1"; Reason="Keep for local dev, but consider deprecating"},
    @{Path="scripts/start-local.bat"; Reason="Keep for Windows local dev"},
    @{Path="scripts/test-api.sh"; Reason="Test utility, can be regenerated"},
    @{Path="scripts/quick-start.ps1"; Reason="Superseded by start-platform.ps1"}
)

foreach ($script in $redundantScriptFiles) {
    # Keep start-local.ps1 and start-local.bat for now
    if ($script.Path -match "start-local\.(ps1|bat)") {
        Write-Host "  [KEEP] $($script.Path) - $($script.Reason)" -ForegroundColor Yellow
        continue
    }
    Remove-FileSafe -Path $script.Path -Reason $script.Reason
}

# 6. Binary/executable files
Write-Host ""
Write-Host "[6/6] Binary and executable files..." -ForegroundColor White
$binaryFiles = @(
    @{Path="apps/cluster-service/cluster-service-linux"; Reason="Linux binary (can be rebuilt)"},
    @{Path="apps/cluster-service/cluster-service.exe"; Reason="Windows binary (can be rebuilt)"},
    @{Path="apps/generator-service/service.go.tmp"; Reason="Temporary file"},
    @{Path="apps/web-admin/cs-binary.tar.gz"; Reason="Binary archive (can be rebuilt)"}
)

foreach ($file in $binaryFiles) {
    Remove-FileSafe -Path $file.Path -Reason $file.Reason
}

# Summary
Write-Host ""
Write-Host "=====================================================" -ForegroundColor Cyan
Write-Host "  Cleanup Summary" -ForegroundColor Cyan
Write-Host "=====================================================" -ForegroundColor Cyan
Write-Host ""
if ($WhatIf) {
    Write-Host "  Mode: PREVIEW (no files were deleted)" -ForegroundColor Yellow
} else {
    Write-Host "  Files deleted: $deletedCount" -ForegroundColor Green
    Write-Host "  Errors:        $errorCount" -ForegroundColor $(if ($errorCount -gt 0) { "Red" } else { "Green" })
}
Write-Host ""
Write-Host "[TIP] Run 'docker system prune' to clean up dangling images" -ForegroundColor Magenta
Write-Host ""
