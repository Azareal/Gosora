@echo off
echo Building the router generator
go build
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo The router generator was successfully built
pause