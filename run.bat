@echo off
echo Generating the dynamic code
go generate
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the executable
go build -o gosora.exe
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Running Gosora
gosora.exe
pause