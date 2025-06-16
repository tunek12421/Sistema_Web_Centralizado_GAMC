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
    Write-ColorOutput "Green" "• Health Check: http://localhost:3000/health"
    Write-ColorOutput "Green" "• MinIO Console: http://localhost:9001"
    Write-Host ""
    Write-ColorOutput "Blue" "Credenciales por defecto:"
    Write-ColorOutput "Cyan" "• Usuario: admin | Contraseña: admin123"
    Write-ColorOutput "Cyan" "• MinIO: gamc_admin | gamc_minio_password_2024"
    Write-Host ""
}

function Reset-All {
    Write-ColorOutput "Yellow" "Reiniciando sistema completo..."
    
    # Parar servicios
    Write-ColorOutput "Yellow" "Deteniendo servicios..."
    docker compose down -v --remove-orphans
    
    # Limpiar volúmenes específicos
    Write-ColorOutput "Yellow" "Limpiando volúmenes..."
    $volumes = @("gamc_postgres_data", "gamc_redis_data", "gamc_minio_data", "gamc_pgadmin_data")
    foreach ($volume in $volumes) {
        docker volume rm $volume -f 2>$null | Out-Null
    }
    
    # Limpiar red problemática
    Write-ColorOutput "Yellow" "Limpiando red..."
    docker network rm gamc_network -f 2>$null | Out-Null
    
    # Limpiar caché de build
    Write-ColorOutput "Yellow" "Limpiando caché de Docker..."
    docker builder prune -f | Out-Null
    
    # Reconstruir imágenes
    Write-ColorOutput "Yellow" "Reconstruyendo imágenes..."
    docker compose build --no-cache
    
    # Iniciar servicios
    Start-Basic
}

function Show-Health {
    Write-ColorOutput "Blue" "Verificando salud del sistema..."
    
    # Test PostgreSQL
    try {
        $result = docker exec gamc_postgres pg_isready -U gamc_user -d gamc_system 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "Green" "✓ PostgreSQL: funcionando"
        } else {
            Write-ColorOutput "Red" "✗ PostgreSQL: no responde"
        }
    } catch {
        Write-ColorOutput "Red" "✗ PostgreSQL: error de conexión"
    }
    
    # Test Redis
    try {
        $redisPing = docker exec gamc_redis redis-cli ping 2>$null
        if ($redisPing -eq "PONG") {
            Write-ColorOutput "Green" "✓ Redis: funcionando correctamente"
        } else {
            Write-ColorOutput "Red" "✗ Redis: no responde"
        }
    } catch {
        Write-ColorOutput "Red" "✗ Redis: error de conexión"
    }
    
    # Test Backend
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:3000/health" -UseBasicParsing -TimeoutSec 5 2>$null
        if ($response.StatusCode -eq 200) {
            Write-ColorOutput "Green" "✓ Backend: API funcionando"
        } else {
            Write-ColorOutput "Red" "✗ Backend: no responde"
        }
    } catch {
        Write-ColorOutput "Red" "✗ Backend: no responde"
    }
    
    # Test Frontend
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:5173" -UseBasicParsing -TimeoutSec 5 2>$null
        if ($response.StatusCode -eq 200) {
            Write-ColorOutput "Green" "✓ Frontend: aplicación disponible"
        } else {
            Write-ColorOutput "Red" "✗ Frontend: no responde"
        }
    } catch {
        Write-ColorOutput "Red" "✗ Frontend: no responde"
    }
}

function Clean-All {
    Write-ColorOutput "Yellow" "Limpiando completamente..."
    
    # Parar todo
    docker compose down -v --remove-orphans 2>$null | Out-Null
    docker stop $(docker ps -aq) 2>$null | Out-Null
    
    # Limpiar volúmenes
    docker volume rm gamc_postgres_data gamc_redis_data gamc_minio_data gamc_pgadmin_data -f 2>$null | Out-Null
    
    # Limpiar red
    docker network rm gamc_network -f 2>$null | Out-Null
    
    Write-ColorOutput "Green" "Limpieza completada"
}

function Show-Help {
    Write-ColorOutput "Blue" "Comandos disponibles:"
    Write-Host ""
    Write-ColorOutput "Green" "start     - Iniciar servicios básicos"
    Write-ColorOutput "Green" "start-all - Iniciar todos los servicios"
    Write-ColorOutput "Green" "stop      - Detener servicios"
    Write-ColorOutput "Green" "restart   - Reiniciar servicios"
    Write-ColorOutput "Green" "reset     - Reinicio completo (limpia todo)"
    Write-ColorOutput "Green" "clean     - Limpiar todo sin reconstruir"
    Write-ColorOutput "Green" "status    - Estado de servicios"
    Write-ColorOutput "Green" "health    - Verificar salud del sistema"
    Write-ColorOutput "Green" "logs      - Ver logs"
    Write-ColorOutput "Green" "urls      - Mostrar URLs"
    Write-ColorOutput "Green" "help      - Esta ayuda"
    Write-Host ""
    Write-Host "Ejemplos:"
    Write-Host "  .\gamc.ps1 start"
    Write-Host "  .\gamc.ps1 logs postgres"
    Write-Host "  .\gamc.ps1 health"
    Write-Host "  .\gamc.ps1 reset"
    Write-Host "  .\gamc.ps1 clean"
}

# Función principal
Show-Banner

switch ($Command.ToLower()) {
    "start" { Start-Basic }
    "start-all" { Start-All }
    "stop" { Stop-Services }
    "restart" { Stop-Services; Start-Sleep -Seconds 3; Start-Basic }
    "reset" { Reset-All }
    "clean" { Clean-All }
    "status" { Show-Status }
    "health" { Show-Health }
    "logs" { Show-Logs $Service }
    "urls" { Show-Urls }
    default { Show-Help }
}