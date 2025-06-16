@echo off
REM ========================================
REM GAMC Sistema Web Centralizado
REM Scripts de Gestión - Windows Version
REM ========================================

setlocal enabledelayedexpansion

REM Colores para Windows (usando echo con caracteres especiales)
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "RED=[91m"
set "NC=[0m"

:show_banner
echo.
echo %BLUE%==========================================
echo   GAMC Sistema Web Centralizado
echo   Gestión de Servicios Unificados
echo ==========================================%NC%
echo.
goto :eof

:show_status
echo %YELLOW%Verificando estado de servicios...%NC%
docker-compose ps
echo.
echo %YELLOW%Servicios por perfil:%NC%
echo • Básicos: postgres, redis, minio, gamc-auth-backend, gamc-auth-frontend
echo • Admin: pgadmin, redis-commander
echo • Setup: minio-client
goto :eof

:start_basic
echo %GREEN%Iniciando servicios básicos...%NC%
docker-compose up -d postgres redis minio gamc-auth-backend gamc-auth-frontend
echo %GREEN%Servicios básicos iniciados%NC%
call :show_urls
goto :eof

:start_all
echo %GREEN%Iniciando todos los servicios...%NC%
docker-compose --profile admin --profile setup up -d
echo %GREEN%Todos los servicios iniciados%NC%
call :show_urls
goto :eof

:start_with_admin
echo %GREEN%Iniciando servicios con herramientas de administración...%NC%
docker-compose --profile admin up -d
echo %GREEN%Servicios con admin iniciados%NC%
call :show_urls
goto :eof

:stop_services
echo %YELLOW%Deteniendo servicios...%NC%
docker-compose --profile admin --profile setup down
echo %GREEN%Servicios detenidos%NC%
goto :eof

:restart_services
echo %YELLOW%Reiniciando servicios...%NC%
call :stop_services
timeout /t 3 /nobreak > nul
call :start_basic
goto :eof

:show_logs
if "%~1"=="" (
    echo %YELLOW%Mostrando logs de todos los servicios...%NC%
    docker-compose logs -f --tail=50
) else (
    echo %YELLOW%Mostrando logs de %1...%NC%
    docker-compose logs -f --tail=50 %1
)
goto :eof

:cleanup
echo %RED%¿Estás seguro de que quieres limpiar todos los datos? (s/N)%NC%
set /p response=
if /i "!response!"=="s" (
    echo %YELLOW%Limpiando sistema...%NC%
    docker-compose --profile admin --profile setup down -v
    docker system prune -f
    docker volume prune -f
    echo %GREEN%Sistema limpiado%NC%
) else (
    echo %GREEN%Operación cancelada%NC%
)
goto :eof

:backup_data
set "BACKUP_DIR=backups\%date:~10,4%%date:~4,2%%date:~7,2%_%time:~0,2%%time:~3,2%%time:~6,2%"
set "BACKUP_DIR=%BACKUP_DIR: =0%"
mkdir "%BACKUP_DIR%" 2>nul

echo %YELLOW%Realizando backup...%NC%

REM Backup PostgreSQL
docker-compose exec postgres pg_dump -U gamc_user gamc_system > "%BACKUP_DIR%\postgres_backup.sql"

REM Backup Redis
docker-compose exec redis redis-cli --rdb - > "%BACKUP_DIR%\redis_backup.rdb"

echo %GREEN%Backup completado en: %BACKUP_DIR%%NC%
goto :eof

:update_services
echo %YELLOW%Actualizando imágenes...%NC%
docker-compose pull
echo %YELLOW%Reconstruyendo servicios personalizados...%NC%
docker-compose build --no-cache gamc-auth-backend gamc-auth-frontend
echo %GREEN%Servicios actualizados%NC%
goto :eof

:show_urls
echo.
echo %BLUE%==========================================
echo   URLs de Acceso
echo ==========================================%NC%
echo %GREEN%🌐 APLICACIONES PRINCIPALES:%NC%
echo   • Frontend Auth:     http://localhost:5173
echo   • Backend Auth API:  http://localhost:3000/api/v1
echo   • Backend Health:    http://localhost:3000/health
echo.
echo %GREEN%🛠️ HERRAMIENTAS DE ADMINISTRACIÓN:%NC%
echo   • MinIO Console:     http://localhost:9001
echo   • PgAdmin:          http://localhost:8080
echo   • Redis Commander:   http://localhost:8081
echo.
echo %GREEN%📊 SERVICIOS DE DATOS:%NC%
echo   • MinIO API:        http://localhost:9000
echo   • PostgreSQL:       localhost:5432
echo   • Redis:            localhost:6379
echo.
echo %YELLOW%🔐 CREDENCIALES POR DEFECTO:%NC%
echo   • PgAdmin:          admin@gamc.gov.bo / admin123
echo   • Redis Commander:  admin / admin123
echo   • MinIO:           gamc_admin / gamc_minio_password_2024
echo.
echo %BLUE%💡 COMANDOS ÚTILES:%NC%
echo   • Ver estado:       gamc status
echo   • Ver logs:         gamc logs [servicio]
echo   • Reiniciar:        gamc restart
echo   • Backup:           gamc backup
echo.
goto :eof

:check_network
echo %YELLOW%Verificando configuración de red...%NC%
docker network ls | findstr gamc_network >nul
if !errorlevel! equ 0 (
    echo %GREEN%✓ Red gamc_network existe%NC%
) else (
    echo %RED%✗ Red gamc_network no existe%NC%
    echo %YELLOW%Creando red...%NC%
    docker network create gamc_network
)
goto :eof

:check_volumes
echo %YELLOW%Verificando volúmenes...%NC%
for %%v in (gamc_postgres_data gamc_redis_data gamc_minio_data gamc_pgadmin_data) do (
    docker volume ls | findstr %%v >nul
    if !errorlevel! equ 0 (
        echo %GREEN%✓ Volumen %%v existe%NC%
    ) else (
        echo %RED%✗ Volumen %%v no existe%NC%
    )
)
goto :eof

:show_help
echo %BLUE%Comandos disponibles:%NC%
echo.
echo %GREEN%start%NC%          - Iniciar servicios básicos
echo %GREEN%start-all%NC%      - Iniciar todos los servicios (incluyendo admin)
echo %GREEN%start-admin%NC%    - Iniciar servicios con herramientas de admin
echo %GREEN%stop%NC%           - Detener todos los servicios
echo %GREEN%restart%NC%        - Reiniciar servicios
echo %GREEN%status%NC%         - Mostrar estado de servicios
echo %GREEN%logs [servicio]%NC% - Mostrar logs (todos o de un servicio específico)
echo %GREEN%urls%NC%           - Mostrar URLs de acceso
echo %GREEN%backup%NC%         - Realizar backup de datos
echo %GREEN%update%NC%         - Actualizar servicios
echo %GREEN%cleanup%NC%        - Limpiar todo el sistema (CUIDADO!)
echo %GREEN%network%NC%        - Verificar configuración de red
echo %GREEN%volumes%NC%        - Verificar volúmenes
echo %GREEN%help%NC%           - Mostrar esta ayuda
echo.
echo %YELLOW%Ejemplos:%NC%
echo   gamc start          # Iniciar servicios básicos
echo   gamc logs backend   # Ver logs del backend
echo   gamc backup         # Hacer backup
echo.
goto :eof

:main
call :show_banner

if "%1"=="" goto show_help
if "%1"=="start" goto start_basic
if "%1"=="start-all" goto start_all
if "%1"=="start-admin" goto start_with_admin
if "%1"=="stop" goto stop_services
if "%1"=="restart" goto restart_services
if "%1"=="status" goto show_status
if "%1"=="logs" goto show_logs %2
if "%1"=="urls" goto show_urls
if "%1"=="backup" goto backup_data
if "%1"=="update" goto update_services
if "%1"=="cleanup" goto cleanup
if "%1"=="network" goto check_network
if "%1"=="volumes" goto check_volumes
if "%1"=="help" goto show_help
goto show_help

call :main %*
pause
