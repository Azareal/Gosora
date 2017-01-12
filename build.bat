@echo off
go build
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
go build ./install
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Gosora was successfully built
pause