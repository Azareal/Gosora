@echo off
echo Installing the dependencies
echo Installing the MySQL Driver
go get -u github.com/go-sql-driver/mysql
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Installing the PostgreSQL Driver
go get -u github.com/lib/pq
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Installing the bcrypt library
go get -u golang.org/x/crypto/bcrypt
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
go get -u golang.org/x/sys/windows
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
go get -u github.com/StackExchange/wmi
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Installing the gopsutil library
go get -u github.com/shirou/gopsutil
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Installing the WebSockets library
go get -u github.com/gorilla/websocket
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the installer
go generate
go build ./install
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
install.exe
