@echo off
echo Generating the dynamic code
go generate
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the executable
go build
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the installer
go build ./install
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Gosora was successfully built
pause