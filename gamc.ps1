# ========================================
# GAMC Sistema Web Centralizado
# Scripts de Gestión - PowerShell Version
# ========================================

param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    
    [Parameter(Position=1)]
    [string]$Service = ""
)

# Función para mostrar banner
function Show-Banner {
    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Blue
    Write-Host "  GAMC Sistema Web Centralizado" -ForegroundColor Blue
    Write-Host "  Gestión de Servicios Unificados" -ForegroundColor Blue
    Write-Host "==========================================" -ForegroundColor Blue
    Write-Host ""
}

# Función para mostrar estado de servicios
function Show-Status {
    Write-Host "Verificando estado de servicios..." -ForegroundColor Yellow
    docker-compose ps
    Write-Host ""
    Write-Host "Servicios por perfil:" -ForegroundColor Yellow
    Write-Host "• Básicos: postgres, redis, minio, gamc-auth-backend, gamc-auth-frontend"
    Write-Host "• Admin: pgadmin, redis-commander"
    Write-Host "• Setup: minio-client"
}

# Función para iniciar servicios básicos
function Start-Basic {
    Write-Host "Iniciando servicios básicos..." -ForegroundColor Green
    docker-compose up -d postgres redis minio gamc-auth-backend gamc-auth-frontend
    Write-Host "Servicios básicos iniciados" -ForegroundColor Green
    Show-Urls
}

# Función para iniciar todos los servicios
function Start-All {
    Write-Host "Iniciando todos los servicios..." -ForegroundColor Green
    docker-compose --profile admin --profile setup up -d
    Write-Host "Todos los servicios iniciados" -ForegroundColor Green
    Show-Urls
}

# Función para iniciar solo con herramientas de admin
function Start-WithAdmin {
    Write-Host "Iniciando servicios con herramientas de administración..." -ForegroundColor Green
    docker-compose --profile admin up -d
    Write-Host "Servicios con admin iniciados" -ForegroundColor Green
    Show-Urls
}

# Función para parar servicios
function Stop-Services {
    Write-Host "Deteniendo servicios..." -ForegroundColor Yellow
    docker-compose --profile admin --profile setup down
    Write-Host "Servicios detenidos" -ForegroundColor Green
}

# Función para reiniciar servicios
function Restart-Services {
    Write-Host "Reiniciando servicios..." -ForegroundColor Yellow
    Stop-Services
    Start-Sleep -Seconds 3
    Start-Basic
}

# Función para mostrar logs
function Show-Logs {
    param([string]$ServiceName)
    
    if ([string]::IsNullOrEmpty($ServiceName)) {
        Write-Host "Mostrando logs de todos los servicios..." -ForegroundColor Yellow
        docker-compose logs -f --tail=50
    } else {
        Write-Host "Mostrando logs de $ServiceName..." -ForegroundColor Yellow
        docker-compose logs -f --tail=50 $ServiceName
    }
}

# Función para limpiar el sistema
function Clear-System {
    $response = Read-Host "¿Estás seguro de que quieres limpiar todos los datos? (s/N)"
    if ($response -eq "s" -or $response -eq "S") {
        Write-Host "Limpiando sistema..." -ForegroundColor Yellow
        docker-compose --profile admin --profile setup down -v
        docker system prune -f
        docker volume prune -f
        Write-Host "Sistema limpiado" -ForegroundColor Green
    } else {
        Write-Host "Operación cancelada" -ForegroundColor Green
    }
}

# Función para hacer backup
function Backup-Data {
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    $backupDir = "backups\$timestamp"
    
    New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
    
    Write-Host "Realizando backup..." -ForegroundColor Yellow
    
    # Backup PostgreSQL
    docker-compose exec postgres pg_dump -U gamc_user gamc_system | Out-File -FilePath "$backupDir\postgres_backup.sql"
    
    # Backup Redis
    docker-compose exec redis redis-cli --rdb - | Set-Content -Path "$backupDir\redis_backup.rdb" -Encoding Byte
    
    Write-Host "Backup completado en: $backupDir" -ForegroundColor Green
}

# Función para actualizar servicios
function Update-Services {
    Write-Host "Actualizando imágenes..." -ForegroundColor Yellow
    docker-compose pull
    Write-Host "Reconstruyendo servicios personalizados..." -ForegroundColor Yellow
    docker-compose build --no-cache gamc-auth-backend gamc-auth-frontend
    Write-Host "Servicios actualizados" -ForegroundColor Green
}

# Función para mostrar URLs de acceso
function Show-Urls {
    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Blue
    Write-Host "  URLs de Acceso" -ForegroundColor Blue
    Write-Host "==========================================" -ForegroundColor Blue
    Write-Host "🌐 APLICACIONES PRINCIPALES:" -ForegroundColor Green
    Write-Host "  • Frontend Auth:     http://localhost:5173"
    Write-Host "  • Backend Auth API:  http://localhost:3000/api/v1"
    Write-Host "  • Backend Health:    http://localhost:3000/health"
    Write-Host ""
    Write-Host "🛠️ HERRAMIENTAS DE ADMINISTRACIÓN:" -ForegroundColor Green
    Write-Host "  • MinIO Console:     http://localhost:9001"
    Write-Host "  • PgAdmin:          http://localhost:8080"
    Write-Host "  • Redis Commander:   http://localhost:8081"
    Write-Host ""
    Write-Host "📊 SERVICIOS DE DATOS:" -ForegroundColor Green
    Write-Host "  • MinIO API:        http://localhost:9000"
    Write-Host "  • PostgreSQL:       localhost:5432"
    Write-Host "  • Redis:            localhost:6379"
    Write-Host ""
    Write-Host "🔐 CREDENCIALES POR DEFECTO:" -ForegroundColor Yellow
    Write-Host "  • PgAdmin:          admin@gamc.gov.bo / admin123"
    Write-Host "  • Redis Commander:  admin / admin123"
    Write-Host "  • MinIO:           gamc_admin / gamc_minio_password_2024"
    Write-Host ""
    Write-Host "💡 COMANDOS ÚTILES:" -ForegroundColor Blue
    Write-Host "  • Ver estado:       .\gamc.ps1 status"
    Write-Host "  • Ver logs:         .\gamc.ps1 logs [servicio]"
    Write-Host "  • Reiniciar:        .\gamc.ps1 restart"
    Write-Host "  • Backup:           .\gamc.ps1 backup"
    Write-Host ""
}

# Función para verificar configuración de red
function Test-Network {
    Write-Host "Verificando configuración de red..." -ForegroundColor Yellow
    $networkExists = docker network ls | Select-String "gamc_network"
    if ($networkExists) {
        Write-Host "✓ Red gamc_network existe" -ForegroundColor Green
    } else {
        Write-Host "✗ Red gamc_network no existe" -ForegroundColor Red
        Write-Host "Creando red..." -ForegroundColor Yellow
        docker network create gamc_network
    }
}

# Función para verificar volúmenes
function Test-Volumes {
    Write-Host "Verificando volúmenes..." -ForegroundColor Yellow
    $volumes = @("gamc_postgres_data", "gamc_redis_data", "gamc_minio_data", "gamc_pgadmin_data")
    
    foreach ($volume in $volumes) {
        $volumeExists = docker volume ls | Select-String $volume
        if ($volumeExists) {
            Write-Host "✓ Volumen $volume existe" -ForegroundColor Green
        } else {
            Write-Host "✗ Volumen $volume no existe" -ForegroundColor Red
        }
    }
}

# Función para mostrar ayuda
function Show-Help {
    Write-Host "Comandos disponibles:" -ForegroundColor Blue
    Write-Host ""
    Write-Host "start" -ForegroundColor Green -NoNewline
    Write-Host "          - Iniciar servicios básicos"
    Write-Host "start-all" -ForegroundColor Green -NoNewline
    Write-Host "      - Iniciar todos los servicios (incluyendo admin)"
    Write-Host "start-admin" -ForegroundColor Green -NoNewline
    Write-Host "    - Iniciar servicios con herramientas de admin"
    Write-Host "stop" -ForegroundColor Green -NoNewline
    Write-Host "           - Detener todos los servicios"
    Write-Host "restart" -ForegroundColor Green -NoNewline
    Write-Host "        - Reiniciar servicios"
    Write-Host "status" -ForegroundColor Green -NoNewline
    Write-Host "         - Mostrar estado de servicios"
    Write-Host "logs [servicio]" -ForegroundColor Green -NoNewline
    Write-Host " - Mostrar logs (todos o de un servicio específico)"
    Write-Host "urls" -ForegroundColor Green -NoNewline
    Write-Host "           - Mostrar URLs de acceso"
    Write-Host "backup" -ForegroundColor Green -NoNewline
    Write-Host "         - Realizar backup de datos"
    Write-Host "update" -ForegroundColor Green -NoNewline
    Write-Host "         - Actualizar servicios"
    Write-Host "cleanup" -ForegroundColor Green -NoNewline
    Write-Host "        - Limpiar todo el sistema (CUIDADO!)"
    Write-Host "network" -ForegroundColor Green -NoNewline
    Write-Host "        - Verificar configuración de red"
    Write-Host "volumes" -ForegroundColor Green -NoNewline
    Write-Host "        - Verificar volúmenes"
    Write-Host "help" -ForegroundColor Green -NoNewline
    Write-Host "           - Mostrar esta ayuda"
    Write-Host ""
    Write-Host "Ejemplos:" -ForegroundColor Yellow
    Write-Host "  .\gamc.ps1 start          # Iniciar servicios básicos"
    Write-Host "  .\gamc.ps1 logs backend   # Ver logs del backend"
    Write-Host "  .\gamc.ps1 backup         # Hacer backup"
    Write-Host ""
}

# Script principal
Show-Banner

switch ($Command.ToLower()) {
    "start" { Start-Basic }
    "start-all" { Start-All }
    "start-admin" { Start-WithAdmin }
    "stop" { Stop-Services }
    "restart" { Restart-Services }
    "status" { Show-Status }
    "logs" { Show-Logs -ServiceName $Service }
    "urls" { Show-Urls }
    "backup" { Backup-Data }
    "update" { Update-Services }
    "cleanup" { Clear-System }
    "network" { Test-Network }
    "volumes" { Test-Volumes }
    default { Show-Help }
}
