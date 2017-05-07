@echo off
echo Installing dependencies
go get -u github.com/go-sql-driver/mysql
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
go get -u golang.org/x/crypto/bcrypt
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
go get -u github.com/StackExchange/wmi
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
go get -u github.com/shirou/gopsutil
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Preparing the installer
go generate
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
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
install.exe