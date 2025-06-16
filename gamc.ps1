param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    [Parameter(Position=1)]
    [string]$Service = ""
)

function Write-ColorOutput {
    param(
        [string]$ForegroundColor,
        [string]$Message
    )
    
    $color = switch($ForegroundColor) {
        "Red" { "Red" }
        "Green" { "Green" }
        "Yellow" { "Yellow" }
        "Blue" { "Blue" }
        "Cyan" { "Cyan" }
        default { "White" }
    }
    
    Write-Host $Message -ForegroundColor $color
}

function Show-Banner {
    Write-Host ""
    Write-ColorOutput "Blue" "=========================================="
    Write-ColorOutput "Blue" "  GAMC Sistema Web Centralizado"
    Write-ColorOutput "Blue" "  Gestión de Servicios Unificados"
    Write-ColorOutput "Blue" "=========================================="
    Write-Host ""
}

function Show-Status {
    Write-ColorOutput "Yellow" "Verificando estado de servicios..."
    docker compose ps
    Write-Host ""
    Write-ColorOutput "Yellow" "Servicios disponibles:"
    Write-Host "• postgres, redis, minio"
    Write-Host "• gamc-auth-backend, gamc-auth-frontend"
}

function Start-Basic {
    Write-ColorOutput "Green" "Iniciando servicios básicos..."
    docker compose up -d postgres redis minio gamc-auth-backend gamc-auth-frontend
    Write-ColorOutput "Green" "Servicios básicos iniciados"
    Show-Urls
}

function Start-All {
    Write-ColorOutput "Green" "Iniciando todos los servicios..."
    docker compose up -d
    Write-ColorOutput "Green" "Todos los servicios iniciados"
    Show-Urls
}

function Stop-Services {
    Write-ColorOutput "Yellow" "Deteniendo servicios..."
    docker compose down
    Write-ColorOutput "Green" "Servicios detenidos"
}

function Show-Logs {
    param([string]$ServiceName)
    
    if ([string]::IsNullOrEmpty($ServiceName)) {
        Write-ColorOutput "Yellow" "Mostrando logs de todos los servicios..."
        docker compose logs -f --tail=50
    } else {
        Write-ColorOutput "Yellow" "Mostrando logs de $ServiceName..."
        docker compose logs -f --tail=50 $ServiceName
    }
}

function Show-Urls {
    Write-Host ""
    Write-ColorOutput "Blue" "URLs de Acceso:"
    Write-ColorOutput "Green" "• Frontend: http://localhost:5173"
    Write-ColorOutput "Green" "• Backend API: http://localhost:3000/api/v1"
    Write-ColorOutput "Green" "• MinIO: http://localhost:9001"
    Write-Host ""
}

function Show-Help {
    Write-ColorOutput "Blue" "Comandos disponibles:"
    Write-Host ""
    Write-ColorOutput "Green" "start     - Iniciar servicios básicos"
    Write-ColorOutput "Green" "start-all - Iniciar todos los servicios"
    Write-ColorOutput "Green" "stop      - Detener servicios"
    Write-ColorOutput "Green" "status    - Estado de servicios"
    Write-ColorOutput "Green" "logs      - Ver logs"
    Write-ColorOutput "Green" "urls      - Mostrar URLs"
    Write-ColorOutput "Green" "help      - Esta ayuda"
    Write-Host ""
    Write-Host "Ejemplos:"
    Write-Host "  .\gamc.ps1 start"
    Write-Host "  .\gamc.ps1 status"
}

Show-Banner

switch ($Command.ToLower()) {
    "start" { Start-Basic }
    "start-all" { Start-All }
    "stop" { Stop-Services }
    "status" { Show-Status }
    "logs" { Show-Logs $Service }
    "urls" { Show-Urls }
    default { Show-Help }
}