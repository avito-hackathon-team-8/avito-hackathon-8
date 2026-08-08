@echo off
setlocal EnableExtensions EnableDelayedExpansion

rem Always run relative to the repository root, even when called elsewhere.
cd /d "%~dp0"

set "COMMAND=%~1"
if not defined COMMAND set "COMMAND=help"
if not "%~1"=="" shift

set "ARG1="
set "ARG2="

:parse_args
if "%~1"=="" goto dispatch
set "ARG=%~1"
if /i "!ARG:~0,5!"=="NAME=" (
    set "NAME=!ARG:~5!"
) else if /i "!ARG:~0,5!"=="TYPE=" (
    set "TYPE=!ARG:~5!"
) else if not defined ARG1 (
    set "ARG1=!ARG!"
) else if not defined ARG2 (
    set "ARG2=!ARG!"
) else (
    echo Unexpected argument: !ARG!
    exit /b 2
)
shift
goto parse_args

:dispatch
if /i "!COMMAND!"=="help" goto help
if /i "!COMMAND!"=="up" goto up
if /i "!COMMAND!"=="down" goto down
if /i "!COMMAND!"=="restart" goto restart
if /i "!COMMAND!"=="build" goto build
if /i "!COMMAND!"=="logs" goto logs
if /i "!COMMAND!"=="ps" goto ps
if /i "!COMMAND!"=="lint" goto lint
if /i "!COMMAND!"=="lint-frontend" goto lint_frontend
if /i "!COMMAND!"=="lint-backend" goto lint_backend
if /i "!COMMAND!"=="branch" goto branch
if /i "!COMMAND!"=="feature" goto feature
if /i "!COMMAND!"=="bugfix" goto bugfix
if /i "!COMMAND!"=="chore" goto chore

echo Unknown command: !COMMAND!
echo Run make.bat help to see available commands.
exit /b 2

:help
echo make.bat up                      Build and start the application
echo make.bat down                    Stop the application
echo make.bat restart                 Restart the application
echo make.bat build                   Build all Docker images
echo make.bat logs                    Follow service logs
echo make.bat ps                      Show service status
echo make.bat lint                    Run all checks
echo make.bat feature NAME=login      Create feature/login
echo make.bat feature login           Create feature/login
echo make.bat bugfix NAME=api         Create bugfix/api
echo make.bat chore NAME=deps         Create chore/deps
echo make.bat branch TYPE=x NAME=y    Create x/y
echo make.bat branch x y              Create x/y
exit /b 0

:up
docker compose up --build -d
exit /b %errorlevel%

:down
docker compose down
exit /b %errorlevel%

:restart
docker compose restart
exit /b %errorlevel%

:build
docker compose build
exit /b %errorlevel%

:logs
docker compose logs -f
exit /b %errorlevel%

:ps
docker compose ps
exit /b %errorlevel%

:lint
call :run_lint_frontend
if errorlevel 1 exit /b %errorlevel%
call :run_lint_backend
exit /b %errorlevel%

:lint_frontend
call :run_lint_frontend
exit /b %errorlevel%

:run_lint_frontend
docker run --rm -v "%CD%\frontend:/app" -v /app/node_modules -w /app node:24-alpine sh -c "npm ci && npm run lint && npm run lint:styles"
exit /b %errorlevel%

:lint_backend
call :run_lint_backend
exit /b %errorlevel%

:run_lint_backend
docker run --rm -v "%CD%:/app:ro" -w /app/backend golangci/golangci-lint:v2.12.2-alpine golangci-lint run --config ../.golangci.yaml
exit /b %errorlevel%

:feature
set "TYPE=feature"
if not defined NAME set "NAME=!ARG1!"
goto create_branch

:bugfix
set "TYPE=bugfix"
if not defined NAME set "NAME=!ARG1!"
goto create_branch

:chore
set "TYPE=chore"
if not defined NAME set "NAME=!ARG1!"
goto create_branch

:branch
if not defined TYPE set "TYPE=!ARG1!"
if not defined NAME set "NAME=!ARG2!"

:create_branch
if not defined TYPE (
    echo TYPE is required
    exit /b 2
)
if not defined NAME (
    echo NAME is required
    exit /b 2
)

set "BRANCH_NAME=!NAME!"
powershell.exe -NoProfile -NonInteractive -Command "if ($env:BRANCH_NAME -cmatch '^[a-z0-9][a-z0-9._-]*$') { exit 0 } else { exit 1 }"
if errorlevel 1 (
    echo NAME must use lowercase letters, numbers, dots, dashes, or underscores
    exit /b 2
)

git switch -c "!TYPE!/!NAME!"
exit /b %errorlevel%
