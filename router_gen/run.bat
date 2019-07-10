@echo off
echo Building the router generator
go build -ldflags="-s -w"
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo The router generator was successfully built
router_gen.exe
pause