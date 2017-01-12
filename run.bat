@echo off
go build
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
gosora.exe
pause